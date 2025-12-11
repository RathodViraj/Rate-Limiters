-- KEYS[1] = rate limit key (e.g. "rl:ip:1.2.3.4")
-- ARGV[1] = current timestamp in milliseconds
-- ARGV[2] = window size in milliseconds (e.g. 5000)
-- ARGV[3] = max allowed requests (e.g. 3)

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

-- remove timestamps outside the window
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)

-- add current timestamp
redis.call('ZADD', key, now, now)

-- count requests in the window
local count = redis.call('ZCARD', key)

-- set expiration slightly > window
redis.call('EXPIRE', key, math.ceil(window / 1000) + 1)

-- return 1 if allowed, 0 if limit exceeded
if count > limit then
  return 0
else
  return 1
end
