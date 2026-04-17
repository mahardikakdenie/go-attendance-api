terjadi ketidak konsistennan antara data berdasarkan Filter bisa fix 

Fetching 1 
curl 'http://localhost:3000/api/v1/hr/roster?start_date=2026-04-13&end_date=2026-04-19' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7' \
  -H 'Connection: keep-alive' \
  -b 'i18n_redirected=en; refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2NmUyYjBkMC1iMTY1LTQyYzgtYmE3NC02ZDEwZmMzZTdiZTciLCJ0eXBlIjoicmVmcmVzaCIsImp0aSI6IjdiYjdmZmVmLWQyMTctNDE4YS1iNzExLTYxN2Q4OWUyMGYzYSIsImlhdCI6MTc3NjA2NTEyNywiZXhwIjoxNzc2NjY5OTI3fQ.rXZ20sWtyJss-c4U-s2TZXbbHPhYekYELa1YfFAJ4Rs; access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzY0MjI3MTAsImlhdCI6MTc3NjMzNjMxMCwicm9sZSI6ImFkbWluIiwidGVuYW50X2lkIjoyLCJ1c2VyX2lkIjoyfQ.s2UrwlGOCafWm7AGPPx4F52TyHpRq0f7AqIIq3a_UFM; __next_hmr_refresh_hash__=118; _dd_s=logs=1&id=b5ce8712-e6a9-4f40-af19-4f743eb1570a&created=1776336293129&expire=1776337882860' \
  -H 'Referer: http://localhost:3000/schedules' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36' \
  -H 'X-Request-ID: 2600877d-8994-4054-912b-0a20fc2163be' \
  -H 'X-Timestamp: 1776336983182' \
  -H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"'

  dengan data 

  {
    "success": true,
    "meta": {
        "message": "Roster fetched successfully",
        "code": 200,
        "status": "success"
    },
    "data": [
        {
            "id": 3,
            "name": "HR Manager",
            "avatar": "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png",
            "department": "HRD",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        },
        {
            "id": 4,
            "name": "Employee User",
            "avatar": "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png",
            "department": "Operations",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "leave",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        },
        {
            "id": 2,
            "name": "Admin PT Friendship",
            "avatar": "",
            "department": "Owner",
            "weeklyRoster": {
                "friday": "leave",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "leave",
                "sunday": "leave",
                "thursday": "leave",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        },
        {
            "id": 8,
            "name": "kucing",
            "avatar": "",
            "department": "HR",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        }
    ]
}

Fetching 2 
curl 'http://localhost:3000/api/v1/hr/roster?start_date=2026-04-12&end_date=2026-04-18' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7' \
  -H 'Connection: keep-alive' \
  -b 'i18n_redirected=en; refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2NmUyYjBkMC1iMTY1LTQyYzgtYmE3NC02ZDEwZmMzZTdiZTciLCJ0eXBlIjoicmVmcmVzaCIsImp0aSI6IjdiYjdmZmVmLWQyMTctNDE4YS1iNzExLTYxN2Q4OWUyMGYzYSIsImlhdCI6MTc3NjA2NTEyNywiZXhwIjoxNzc2NjY5OTI3fQ.rXZ20sWtyJss-c4U-s2TZXbbHPhYekYELa1YfFAJ4Rs; access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzY0MjI3MTAsImlhdCI6MTc3NjMzNjMxMCwicm9sZSI6ImFkbWluIiwidGVuYW50X2lkIjoyLCJ1c2VyX2lkIjoyfQ.s2UrwlGOCafWm7AGPPx4F52TyHpRq0f7AqIIq3a_UFM; __next_hmr_refresh_hash__=118; _dd_s=logs=1&id=b5ce8712-e6a9-4f40-af19-4f743eb1570a&created=1776336293129&expire=1776337903353' \
  -H 'Referer: http://localhost:3000/' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36' \
  -H 'X-Request-ID: a784d73e-c2c6-4504-acef-dbfa16fb1689' \
  -H 'X-Timestamp: 1776337003928' \
  -H 'sec-ch-ua: "Google Chrome";v="147", "Not.A/Brand";v="8", "Chromium";v="147"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Windows"'

  Response 

  {
    "success": true,
    "meta": {
        "message": "Roster fetched successfully",
        "code": 200,
        "status": "success"
    },
    "data": [
        {
            "id": 3,
            "name": "HR Manager",
            "avatar": "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png",
            "department": "HRD",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "leave",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "leave",
                "wednesday": "leave"
            }
        },
        {
            "id": 4,
            "name": "Employee User",
            "avatar": "http://i.ibb.co.com/p6119B1C/attendance-1775556680532.png",
            "department": "Operations",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "leave"
            }
        },
        {
            "id": 2,
            "name": "Admin PT Friendship",
            "avatar": "",
            "department": "Owner",
            "weeklyRoster": {
                "friday": "leave",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "leave",
                "sunday": "leave",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        },
        {
            "id": 8,
            "name": "kucing",
            "avatar": "",
            "department": "HR",
            "weeklyRoster": {
                "friday": "work_shift_tenant (00:00 - 17:00)",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "work_shift_tenant (00:00 - 17:00)",
                "sunday": "work_shift_tenant (00:00 - 17:00)",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }
        }
    ]
}

perhatikan di user Admin PT Friendship

"weeklyRoster": {
                "friday": "leave",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "leave",
                "sunday": "leave",
                "thursday": "leave",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }

			 "weeklyRoster": {
                "friday": "leave",
                "monday": "work_shift_tenant (00:00 - 17:00)",
                "saturday": "leave",
                "sunday": "leave",
                "thursday": "work_shift_tenant (00:00 - 17:00)",
                "tuesday": "work_shift_tenant (00:00 - 17:00)",
                "wednesday": "work_shift_tenant (00:00 - 17:00)"
            }


			di thuesday padahal dengan filter yang beda cuman 1 hari, bisa ga di perjelas mungkin data balikan Ke Users biar tidak ada kesalah pahaman soalnya hari ini Akun Admin PT Friendship harus nya CUTI bukan kerja 

			bisa fix 
