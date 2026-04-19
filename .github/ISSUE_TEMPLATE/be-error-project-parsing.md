bisa fix error ini di curl

curl 'http://localhost:3001/api/v1/projects' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -b 'access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzY2MjkzOTcsImlhdCI6MTc3NjU0Mjk5Nywicm9sZSI6ImhyIiwidGVuYW50X2lkIjoyLCJ1c2VyX2lkIjozfQ.aQqWy0zHvJyZnU1VhQLzsZ_vSG0J-SSoiT47E7-hoko' \
  -H 'Origin: http://localhost:3001' \
  -H 'Referer: http://localhost:3001/projects' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36' \
  -H 'X-Request-ID: c1097508-46f9-43e4-afe1-5194f2da0275' \
  -H 'X-Timestamp: 1776543034592' \
  -H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  --data-raw '{"name":"kjandand","client_name":"aidhadnain","budget":1321312312,"description":"ajdadn","start_date":"2026-04-19","status":"ACTIVE","end_date":"2026-04-20"}'


"parsing time \"2026-04-19\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\""

dengan response 

{
    "success": false,
    "meta": {
        "message": "Invalid request",
        "code": 400,
        "status": "error"
    },
    "data": "parsing time \"2026-04-19\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\""
}
