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

// recordFailureScript 累加失败次数，达到阈值后写入固定时长的锁。
// KEYS: fail, lock；ARGV: windowMs, threshold, lockMs
const recordFailureScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
if count < tonumber(ARGV[2]) then return 0 end

local lockMs = tonumber(ARGV[3])
redis.call('PSETEX', KEYS[2], lockMs, '1')
redis.call('DEL', KEYS[1])
return lockMs
`

// recordAccountFailureScript 累加账号失败次数，达到阈值后按等级翻倍锁定。
// KEYS: fail, lock, level；ARGV: windowMs, threshold, baseLockMs, maxLockMs, levelTtlMs
const recordAccountFailureScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
if count < tonumber(ARGV[2]) then return 0 end

local level = redis.call('INCR', KEYS[3])
redis.call('PEXPIRE', KEYS[3], ARGV[5])
local lockMs = math.floor(math.min(tonumber(ARGV[4]), tonumber(ARGV[3]) * 2 ^ (level - 1)))

redis.call('PSETEX', KEYS[2], lockMs, '1')
redis.call('DEL', KEYS[1])
return lockMs
`
