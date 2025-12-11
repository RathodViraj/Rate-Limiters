-- KEYS[1] = base key prefix, e.g. "user:1.2.3.4"
-- ARGV[1] = current timestamp (seconds)
-- ARGV[2] = window size (seconds)
-- ARGV[3] = limit (max requests allowed)
-- returns: { allowed(1/0), total_count, retry_after_seconds }

local base = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

local current_window = now - (now % window)
local prev_window = current_window - window

local current_key = base .. ":w:" .. current_window
local prev_key = base .. ":w:" .. prev_window

local current_count = tonumber(redis.call("GET", current_key) or "0")
local prev_count = tonumber(redis.call("GET", prev_key) or "0")

local elapsed = now - current_window
local weight = elapsed / window

-- Weighted sliding window formula
local total = current_count + (1 - weight) * prev_count

-- Check limit
if total >= limit then
    local retry_after = window - elapsed
    return {0, total, retry_after}  -- request rejected
end

redis.call("INCR", current_key)
redis.call("EXPIRE", current_key, window * 2)

-- Return success
return {1, total + 1, 0}
