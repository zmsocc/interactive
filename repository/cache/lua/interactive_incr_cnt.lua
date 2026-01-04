---
--- Created by zhang san.
--- DateTime: 2025/5/10 18:22
---
local key = KEYS[1]
local cntKey = ARGV[1]
local delta = tonumber(ARGV[2])
local exists = redis.call("EXISTS", key)
if exists == 1 then
    redis.call("HINCRBY", key, cntKey, delta)
    -- 说明自增成功了
    return 1
else
    return 0
end