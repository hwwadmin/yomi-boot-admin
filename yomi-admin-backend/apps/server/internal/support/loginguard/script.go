package loginguard

// checkLocksScript 读取多个 lock key 的剩余 TTL，返回最大值（毫秒）。
// 无 TTL 的 key 属于异常状态，直接删除自愈。
const checkLocksScript = `
local maxTTL = 0
for i = 1, #KEYS do
  local ttl = redis.call('PTTL', KEYS[i])
  if ttl == -1 then
    redis.call('DEL', KEYS[i])
  elseif ttl > maxTTL then
    maxTTL = ttl
  end
end
return maxTTL
`

// reserveAttemptScript 在一次调用内完成锁定检查与失败预算预占，消除「检查→验密→记账」之间的竞态。
//
// KEYS: ip.lock, ip.fail, ip.pending
//
//	[, pair.lock, pair.fail, pair.pending, account.lock, account.fail, account.pending]
//
// ARGV: pendingTtlMs, ipThreshold [, pairThreshold, accountThreshold]
//
// 返回 {kind, ttlMs}：kind=0 允许继续验密；kind=1 命中锁定；kind=2 预算已被占满。
// 预算判定使用「已提交失败数 + 在途预占数」，保证同一窗口内进入验密的请求数不超过阈值。
const reserveAttemptScript = `
local function remainingTTL(key)
  local ttl = redis.call('PTTL', key)
  if ttl == -1 then
    redis.call('DEL', key)
    return 0
  end
  if ttl < 0 then
    return 0
  end
  return ttl
end

local lockTTL = 0
for i = 1, #KEYS, 3 do
  local ttl = remainingTTL(KEYS[i])
  if ttl > lockTTL then
    lockTTL = ttl
  end
end
if lockTTL > 0 then
  return {1, lockTTL}
end

local exhausted = false
local waitMs = 0
for i = 1, #KEYS, 3 do
  local fail = tonumber(redis.call('GET', KEYS[i + 1]) or '0')
  local pending = tonumber(redis.call('GET', KEYS[i + 2]) or '0')
  local dimension = math.floor((i - 1) / 3)
  local threshold = tonumber(ARGV[2 + dimension])
  if fail + pending >= threshold then
    exhausted = true
    local failTTL = remainingTTL(KEYS[i + 1])
    local pendingTTL = remainingTTL(KEYS[i + 2])
    if failTTL > waitMs then waitMs = failTTL end
    if pendingTTL > waitMs then waitMs = pendingTTL end
  end
end
if exhausted then
  return {2, waitMs}
end

for i = 1, #KEYS, 3 do
  redis.call('INCR', KEYS[i + 2])
  redis.call('PEXPIRE', KEYS[i + 2], tonumber(ARGV[1]))
end
return {0, 0}
`

// releaseAttemptScript 释放验密前预占的失败预算；验密成功或链路异常时调用。
//
// KEYS: ip.pending [, pair.pending, account.pending]
const releaseAttemptScript = `
local function releasePending(pendingKey)
  local pending = tonumber(redis.call('GET', pendingKey) or '0')
  if pending <= 0 then
    return
  end
  if pending == 1 then
    redis.call('DEL', pendingKey)
  else
    redis.call('DECR', pendingKey)
  end
end

for i = 1, #KEYS do
  releasePending(KEYS[i])
end
return 0
`

// recordFailureScript 把一次验密失败从预占转为正式计数：释放 pending、累加失败数，任一维度达阈值即写锁。
//
// KEYS: ip.lock, ip.fail, ip.pending
//
//	[, pair.lock, pair.fail, pair.pending, account.lock, account.fail, account.pending, account.level]
//
// ARGV: ipWindowMs, ipThreshold, ipLockMs
//
//	[, pairWindowMs, pairThreshold, pairLockMs,
//	   accountWindowMs, accountThreshold, accountBaseLockMs, accountMaxLockMs, accountLevelTtlMs]
//
// KEYS 为 3 个时只记 IP 维度（空账号）；10 个时同时记 IP、IP+账号、账号三个维度。
// 返回与维度顺序一致的锁时长（毫秒）数组，0 表示该维度未触发锁定。
const recordFailureScript = `
local function releasePending(pendingKey)
  local pending = tonumber(redis.call('GET', pendingKey) or '0')
  if pending <= 0 then
    return
  end
  if pending == 1 then
    redis.call('DEL', pendingKey)
  else
    redis.call('DECR', pendingKey)
  end
end

local function recordFixed(lockKey, failKey, pendingKey, windowMs, threshold, lockMs)
  releasePending(pendingKey)
  local count = redis.call('INCR', failKey)
  if count == 1 then redis.call('PEXPIRE', failKey, windowMs) end
  if count < tonumber(threshold) then return 0 end
  redis.call('PSETEX', lockKey, tonumber(lockMs), '1')
  redis.call('DEL', failKey)
  return tonumber(lockMs)
end

local function recordEscalating(lockKey, failKey, pendingKey, levelKey, windowMs, threshold, baseLockMs, maxLockMs, levelTtlMs)
  releasePending(pendingKey)
  local count = redis.call('INCR', failKey)
  if count == 1 then redis.call('PEXPIRE', failKey, windowMs) end
  if count < tonumber(threshold) then return 0 end
  local level = redis.call('INCR', levelKey)
  redis.call('PEXPIRE', levelKey, levelTtlMs)
  local lock = math.floor(math.min(tonumber(maxLockMs), tonumber(baseLockMs) * 2 ^ (level - 1)))
  redis.call('PSETEX', lockKey, lock, '1')
  redis.call('DEL', failKey)
  return lock
end

local locks = {}
locks[1] = recordFixed(KEYS[1], KEYS[2], KEYS[3], ARGV[1], ARGV[2], ARGV[3])
if #KEYS >= 10 then
  locks[2] = recordFixed(KEYS[4], KEYS[5], KEYS[6], ARGV[4], ARGV[5], ARGV[6])
  locks[3] = recordEscalating(KEYS[7], KEYS[8], KEYS[9], KEYS[10], ARGV[7], ARGV[8], ARGV[9], ARGV[10], ARGV[11])
end
return locks
`

// recordSuccessScript 登录链路成功后释放预占，并清空组合与账号维度的失败计数与锁定。
// IP 失败计数跨账号，成功登录不能把 IP 的喷洒记录洗白，因此只释放其预占、不删除计数。
//
// KEYS: ip.pending
//
//	[, pair.pending, pair.fail, pair.lock, account.pending, account.fail, account.lock, account.level]
const recordSuccessScript = `
local function releasePending(pendingKey)
  local pending = tonumber(redis.call('GET', pendingKey) or '0')
  if pending <= 0 then
    return
  end
  if pending == 1 then
    redis.call('DEL', pendingKey)
  else
    redis.call('DECR', pendingKey)
  end
end

releasePending(KEYS[1])
if #KEYS >= 8 then
  releasePending(KEYS[2])
  redis.call('DEL', KEYS[3], KEYS[4])
  releasePending(KEYS[5])
  redis.call('DEL', KEYS[6], KEYS[7], KEYS[8])
end
return 0
`
