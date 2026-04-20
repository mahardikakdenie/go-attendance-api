curl 'http://localhost:3000/api/v1/users' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7' \
  -H 'Connection: keep-alive' \
  -H 'Content-Type: application/json' \
  -b 'i18n_redirected=en; refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2NmUyYjBkMC1iMTY1LTQyYzgtYmE3NC02ZDEwZmMzZTdiZTciLCJ0eXBlIjoicmVmcmVzaCIsImp0aSI6IjdiYjdmZmVmLWQyMTctNDE4YS1iNzExLTYxN2Q4OWUyMGYzYSIsImlhdCI6MTc3NjA2NTEyNywiZXhwIjoxNzc2NjY5OTI3fQ.rXZ20sWtyJss-c4U-s2TZXbbHPhYekYELa1YfFAJ4Rs; Profile=%7B%22id%22%3A%222efe7aa3-469f-4b20-afb1-21b50dc48252%22%2C%22account_id%22%3A%2206ae66af-928a-499d-9792-7dcbde3598af%22%2C%22name%22%3A%22Dev-Dika%22%2C%22email%22%3A%22dev-dika%40yopmail.com%22%2C%22phone_number%22%3A%22%2B628123456789%22%2C%22role%22%3A%22user%22%2C%22division%22%3A%22d175de19-a7fb-4980-a77e-332996bb2650%22%2C%22created_at%22%3A%222026-01-20T02%3A08%3A48.079Z%22%2C%22updated_at%22%3A%222026-01-20T02%3A08%3A48.079Z%22%2C%22deleted_at%22%3Anull%7D; _SID_Teman=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImRldi1kaWthQHlvcG1haWwuY29tIiwicGhvbmVfbnVtYmVyIjoiKzYyODEyMzQ1Njc4OSIsInN1YiI6IjA2YWU2NmFmLTkyOGEtNDk5ZC05NzkyLTdkY2JkZTM1OThhZiIsIm5hbWUiOiJEZXYtRGlrYSIsInJvbGUiOiJVc2VyIiwiY2hhbm5lbCI6IjQwZWVlNWJmLTJiOTItNGQyMy1iZTU1LWY5Y2FhOWQzZWE4OCIsInBlcm1pc3Npb25fbGlzdCI6WyJUaWNrZXRpbmcudGlja2V0Lm1lbnUuYXNzaWduZWRUb01lIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24uYWNjZXB0YW5jZSIsIlRpY2tldGluZy50aWNrZXQuYWN0aW9uLmNvbXBsZXRlIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24ucmVxdWVzdCIsIlRpY2tldGluZy5kYXNoYm9hcmQucmVhZCIsIlRpY2tldGluZy5sb2cucmVhZCIsIlRpY2tldGluZy50aWNrZXQuYWN0aW9uLnJldmlldyJdLCJhY2NvdW50X2luc3VyZXJzIjpbXSwiYWNjb3VudF9jaGFubmVscyI6W3siY2hhbm5lbCI6IjQwZWVlNWJmLTJiOTItNGQyMy1iZTU1LWY5Y2FhOWQzZWE4OCJ9XSwibGFzdF9sb2dpbiI6IjIwMjYtMDQtMjBUMDM6MzY6MzUuMjEwWiIsImlhdCI6MTc3NjY1NzExMywiZXhwIjoxNzc2NzQzNTEzfQ.APfxABnNfJqcVUTAqMuwD7yKlFf-C0aArpyMnLyuLqQ; access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzY3NTE2ODMsImlhdCI6MTc3NjY2NTI4Mywicm9sZSI6InN1cGVyYWRtaW4iLCJ0ZW5hbnRfaWQiOjEsInVzZXJfaWQiOjF9.uk7JOWNuxq6FLniOa5xL4-QBKvlW2UrkL-tRlV3UpBI; __next_hmr_refresh_hash__=440; _dd_s=logs=1&id=37749e8f-ad32-4e11-8ba1-d83b47336ea4&created=1776665274897&expire=1776669804766' \
  -H 'Origin: http://localhost:3000' \
  -H 'Referer: http://localhost:3000/employees' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36' \
  -H 'X-Request-ID: cc1dc395-7026-4435-8f33-ba0a77ff5f10' \
  -H 'X-Timestamp: 1776668904767' \
  -H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"' \
  --data-raw '{"name":"mahardika kessuma","email":"dd@yopmail.com","employee_id":"asdasd","department":"asdad","phone_number":"08123456789","address":"adnad","base_salary":2000,"role_id":3,"password":""}'

  bisa fix ERROR masalah ini 
  dengan error 

  {
    "success": false,
    "meta": {
        "message": "Failed to create user",
        "code": 400,
        "status": "error"
    },
    "data": "failed to create user payroll profile: ERROR: insert or update on table \"user_payroll_profiles\" violates foreign key constraint \"fk_user_payroll_profiles_user\" (SQLSTATE 23503)"
}
