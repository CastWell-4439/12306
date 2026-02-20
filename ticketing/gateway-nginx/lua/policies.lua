local trace_id = ngx.req.get_headers()["X-Trace-Id"]
if not trace_id or trace_id == "" then
    ngx.status = 400
    ngx.header["Content-Type"] = "application/json"
    ngx.say('{"error":"X-Trace-Id is required"}')
    return ngx.exit(400)
end

-- Minimal local rate limit policy: 20 requests per minute per trace_id.
local key = "rl:" .. trace_id
local dict = ngx.shared.rate_limit
if not dict then
    return
end

local current, err = dict:incr(key, 1, 0, 60)
if not current then
    ngx.log(ngx.ERR, "rate limit incr failed: ", err)
    return
end

if current > 20 then
    ngx.status = 429
    ngx.header["Content-Type"] = "application/json"
    ngx.say('{"error":"rate limit exceeded"}')
    return ngx.exit(429)
end


