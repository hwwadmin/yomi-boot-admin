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

// recordFailureScript 在一次调用内累加一次凭证失败涉及的全部维度，任一维度达到阈值即写锁，原子生效。
//
// KEYS: ip.fail, ip.lock [, pair.fail, pair.lock, account.fail, account.lock, account.level]
// ARGV: ipWindowMs, ipThreshold, ipLockMs
//
//	[, pairWindowMs, pairThreshold, pairLockMs,
//	   accountWindowMs, accountThreshold, accountBaseLockMs, accountMaxLockMs, accountLevelTtlMs]
//
// KEYS 为 2 个时只记 IP 维度（空账号）；7 个时同时记 IP、IP+账号、账号三个维度。
// 返回与维度顺序一致的锁时长（毫秒）数组，0 表示该维度未触发锁定。
const recordFailureScript = `
local function recordFixed(failKey, lockKey, windowMs, threshold, lockMs)
	local count = redis.call('INCR', failKey)
	if count == 1 then redis.call('PEXPIRE', failKey, windowMs) end
	if count < tonumber(threshold) then return 0 end
	redis.call('PSETEX', lockKey, tonumber(lockMs), '1')
	redis.call('DEL', failKey)
	return tonumber(lockMs)
end

local function recordEscalating(failKey, lockKey, levelKey, windowMs, threshold, baseLockMs, maxLockMs, levelTtlMs)
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
locks[1] = recordFixed(KEYS[1], KEYS[2], ARGV[1], ARGV[2], ARGV[3])
if #KEYS >= 7 then
	locks[2] = recordFixed(KEYS[3], KEYS[4], ARGV[4], ARGV[5], ARGV[6])
	locks[3] = recordEscalating(KEYS[5], KEYS[6], KEYS[7], ARGV[7], ARGV[8], ARGV[9], ARGV[10], ARGV[11])
end
return locks
`
