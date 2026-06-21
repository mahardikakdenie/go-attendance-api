curl ^"http://localhost:3000/api/v1/menus/me^" ^
  -H ^"Accept: application/json, text/plain, */*^" ^
  -H ^"Accept-Language: id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7^" ^
  -H ^"Connection: keep-alive^" ^
  -b ^"i18n_redirected=en; refresh_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI2NmUyYjBkMC1iMTY1LTQyYzgtYmE3NC02ZDEwZmMzZTdiZTciLCJ0eXBlIjoicmVmcmVzaCIsImp0aSI6IjBiZjA4YWNkLWI3NTgtNGU5Mi1iNjc5LTQ5MjRlNjEyOGMxNiIsImlhdCI6MTc4MDAxODA5OCwiZXhwIjoxNzgwNjIyODk4fQ.k37xmj0wS_F0_Kb-1G2fvnKL4Yiq2uJwocraN3yIzis; terms_accepted=true; Profile=^%^7B^%^22id^%^22^%^3A^%^22e05f31ea-a362-4b9a-be5f-1a4f6fee254c^%^22^%^2C^%^22name^%^22^%^3A^%^22Budy^%^20Santoso^%^22^%^2C^%^22email^%^22^%^3A^%^22budysantoso^%^40yopmail.com^%^22^%^2C^%^22phone_number^%^22^%^3A^%^22^%^2B628123456789^%^22^%^2C^%^22division^%^22^%^3A^%^22d175de19-a7fb-4980-a77e-332996bb2650^%^22^%^2C^%^22created_at^%^22^%^3A^%^222026-01-08T03^%^3A15^%^3A34.803Z^%^22^%^2C^%^22updated_at^%^22^%^3A^%^222026-01-08T03^%^3A15^%^3A34.803Z^%^22^%^2C^%^22deleted_at^%^22^%^3Anull^%^2C^%^22account_id^%^22^%^3A^%^22a85190b1-62f9-4e7a-be2a-d5a7bcbb0157^%^22^%^2C^%^22role^%^22^%^3A^%^22head^%^22^%^2C^%^22divisions^%^22^%^3A^%^7B^%^22id^%^22^%^3A^%^22d175de19-a7fb-4980-a77e-332996bb2650^%^22^%^2C^%^22name^%^22^%^3A^%^22IT^%^20Engineering^%^22^%^7D^%^2C^%^22division_approval_pics^%^22^%^3A^%^5B^%^7B^%^22id^%^22^%^3A^%^228008a5c0-8d2e-448f-9f6c-2fc29e4fde1b^%^22^%^2C^%^22division_id^%^22^%^3A^%^22d175de19-a7fb-4980-a77e-332996bb2650^%^22^%^2C^%^22profile_id^%^22^%^3A^%^22e05f31ea-a362-4b9a-be5f-1a4f6fee254c^%^22^%^2C^%^22created_at^%^22^%^3A^%^222026-05-25T15^%^3A06^%^3A24.261Z^%^22^%^7D^%^5D^%^2C^%^22tickets_requested^%^22^%^3A^%^5B^%^5D^%^2C^%^22timesheet_approvals^%^22^%^3A^%^5B^%^5D^%^2C^%^22is_division_approval_pic^%^22^%^3Atrue^%^7D; _SID_Teman=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImJ1ZHlzYW50b3NvQHlvcG1haWwuY29tIiwicGhvbmVfbnVtYmVyIjoiKzYyODEyMzQ1Njc4OSIsInN1YiI6ImE4NTE5MGIxLTYyZjktNGU3YS1iZTJhLWQ1YTdiY2JiMDE1NyIsIm5hbWUiOiJCdWR5IFNhbnRvc28iLCJyb2xlIjoiQWRtaW4iLCJjaGFubmVsIjoiNDBlZWU1YmYtMmI5Mi00ZDIzLWJlNTUtZjljYWE5ZDNlYTg4IiwicGVybWlzc2lvbl9saXN0IjpbIlRpY2tldGluZy5kaXZpc2lvbi5jcmVhdGUiLCJUaWNrZXRpbmcuZGl2aXNpb24uZGVsZXRlIiwiVGlja2V0aW5nLnByb2plY3QudXBkYXRlIiwiVGlja2V0aW5nLmRhc2hib2FyZC5yZWFkIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24uYWNjZXB0YW5jZSIsIlRpY2tldGluZy50aWNrZXQubWVudS5hc3NpZ25lZFRvTWUiLCJUaWNrZXRpbmcucHJvamVjdC5yZWFkIiwiVGlja2V0aW5nLnByb2plY3QuZGVsZXRlIiwiVGlja2V0aW5nLmRpdmlzaW9uLnVwZGF0ZSIsIlRpY2tldGluZy5wcm9qZWN0LmNyZWF0ZSIsIlRpY2tldGluZy50aWNrZXQuYWN0aW9uLmFwcHJvdmFsIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24ucmVxdWVzdCIsIlRpY2tldGluZy5sb2cucmVhZCIsIlRpY2tldGluZy5kaXZpc2lvbi5yZWFkIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24uY29tcGxldGUiLCJVc2VyLlVwZGF0ZSIsIlVzZXIuQ2hhbmdlIFN0YXR1cyIsIlVzZXIuUmVhZCIsIlVzZXIuRGVsZXRlIiwiVGlja2V0aW5nLnRpY2tldC5hY3Rpb24ucmV2aWV3Il0sImFjY291bnRfaW5zdXJlcnMiOltdLCJhY2NvdW50X2NoYW5uZWxzIjpbeyJjaGFubmVsIjoiNDBlZWU1YmYtMmI5Mi00ZDIzLWJlNTUtZjljYWE5ZDNlYTg4In1dLCJsYXN0X2xvZ2luIjoiMjAyNi0wNS0zMFQxNzo1NTo0Ni40MjRaIiwiaWF0IjoxNzgwMTcwMDAwLCJleHAiOjE3ODAyNTY0MDB9.aSqXw-qo8W860ykTUjcLZs92nIR7EPDxQ0xYkCi9iBg; access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODAzNzM4OTYsImlhdCI6MTc4MDI4NzQ5Niwicm9sZSI6InN1cGVyYWRtaW4iLCJ0ZW5hbnRfaWQiOjEsInVzZXJfaWQiOjF9.JCV6gyxhTFE8RFu5_9KjLeYj_BQMGH1Qrlbm5ObBfKk; auth_token=eyJhbGciOiJIUzI1NiJ9.eyJhdXRoZW50aWNhdGVkIjp0cnVlLCJpYXQiOjE3ODAzMjQ2MDgsImV4cCI6MTc4MDQxMTAwOH0.hW9isppmcvHuCRp0grkvOuMh-ELNAZvqHI7rkOmpuU8; _dd_s=logs=1^&id=0774048e-36aa-47d0-aada-421a69ceb6e8^&created=1780334498327^&expire=1780337407956^" ^
  -H ^"Referer: http://localhost:3000/admin/support^" ^
  -H ^"Sec-Fetch-Dest: empty^" ^
  -H ^"Sec-Fetch-Mode: cors^" ^
  -H ^"Sec-Fetch-Site: same-origin^" ^
  -H ^"User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36^" ^
  -H ^"X-Request-ID: 261fd10c-19f0-4ee3-9039-9b95cec4fbc2^" ^
  -H ^"X-Timestamp: 1780336510517^" ^
  -H ^"sec-ch-ua: ^\^"Chromium^\^";v=^\^"148^\^", ^\^"Google Chrome^\^";v=^\^"148^\^", ^\^"Not/A)Brand^\^";v=^\^"99^\^"^" ^
  -H ^"sec-ch-ua-mobile: ?0^" ^
  -H ^"sec-ch-ua-platform: ^\^"Windows^\^"^"

  response 

  {
    "success": true,
    "meta": {
        "message": "Menus retrieved successfully",
        "code": 200,
        "status": "success"
    },
    "data": [
        {
            "id": 1,
            "key": "platform-group",
            "label": "Platform Control",
            "icon": "ShieldCheck",
            "children": [
                {
                    "id": 2,
                    "key": "manage-tenants",
                    "label": "Tenant Directory",
                    "icon": "Building2",
                    "path": "/admin/tenants",
                    "required_permission": "superadmin.access"
                },
                {
                    "id": 3,
                    "key": "subscriptions",
                    "label": "Global Billing",
                    "icon": "CreditCard",
                    "path": "/admin/subscriptions",
                    "required_permission": "superadmin.access"
                },
                {
                    "id": 4,
                    "key": "accounts",
                    "label": "Platform Admins",
                    "icon": "UserCheck",
                    "path": "/admin/accounts",
                    "required_permission": "superadmin.access"
                },
                {
                    "id": 5,
                    "key": "platform-roles",
                    "label": "System Governance",
                    "icon": "ShieldAlert",
                    "path": "/admin/roles",
                    "required_permission": "rbac.manage"
                },
                {
                    "id": 6,
                    "key": "support-desk",
                    "label": "Support Desk",
                    "icon": "MessageSquare",
                    "path": "/admin/support",
                    "required_permission": "support.view"
                }
            ]
        },
        {
            "id": 7,
            "key": "intelligence-group",
            "label": "Intelligence Hub",
            "icon": "LayoutDashboard",
            "children": [
                {
                    "id": 8,
                    "key": "dashboard",
                    "label": "Main Dashboard",
                    "icon": "LayoutDashboard",
                    "path": "/",
                    "module": "attendance"
                },
                {
                    "id": 9,
                    "key": "analytics",
                    "label": "Workforce Intel",
                    "icon": "TrendingUp",
                    "path": "/analytics",
                    "module": "analytics",
                    "required_permission": "analytics.executive"
                }
            ]
        },
        {
            "id": 10,
            "key": "workforce-group",
            "label": "Workforce Management",
            "icon": "Users",
            "children": [
                {
                    "id": 11,
                    "key": "all-employees",
                    "label": "Staff Directory",
                    "icon": "Users",
                    "path": "/employees",
                    "module": "user",
                    "required_permission": "employee.view"
                },
                {
                    "id": 12,
                    "key": "all-attendance",
                    "label": "Attendance Logs",
                    "icon": "CalendarDays",
                    "path": "/attendances",
                    "module": "attendance",
                    "required_permission": "attendance.view"
                },
                {
                    "id": 13,
                    "key": "work-schedules",
                    "label": "Shift Rosters",
                    "icon": "Clock",
                    "path": "/schedules",
                    "module": "schedule",
                    "required_permission": "schedule.view"
                },
                {
                    "id": 14,
                    "key": "manage-leaves",
                    "label": "Leave Approvals",
                    "icon": "CalendarX",
                    "path": "/leaves",
                    "module": "leave",
                    "required_permission": "leave.view"
                },
                {
                    "id": 15,
                    "key": "manage-overtime",
                    "label": "Overtime Desk",
                    "icon": "Clock",
                    "path": "/overtime",
                    "module": "overtime",
                    "required_permission": "overtime.view"
                }
            ]
        },
        {
            "id": 16,
            "key": "performance-group",
            "label": "Performance \u0026 Ops",
            "icon": "Target",
            "children": [
                {
                    "id": 17,
                    "key": "performance-goals",
                    "label": "Strategic Goals",
                    "icon": "Target",
                    "path": "/performance/goals",
                    "module": "performance",
                    "required_permission": "performance.manage"
                },
                {
                    "id": 18,
                    "key": "performance-appraisals",
                    "label": "Staff Appraisals",
                    "icon": "Star",
                    "path": "/performance/appraisals",
                    "module": "performance",
                    "required_permission": "performance.view"
                },
                {
                    "id": 19,
                    "key": "projects",
                    "label": "Project Tracker",
                    "icon": "Briefcase",
                    "path": "/projects",
                    "module": "project",
                    "required_permission": "project.manage"
                },
                {
                    "id": 20,
                    "key": "timesheet-monitoring",
                    "label": "Timesheet Audit",
                    "icon": "ActivityIcon",
                    "path": "/timesheet/monitoring",
                    "module": "project",
                    "required_permission": "project.manage"
                }
            ]
        },
        {
            "id": 21,
            "key": "financial-group",
            "label": "Financial Center",
            "icon": "Coins",
            "children": [
                {
                    "id": 22,
                    "key": "payroll-list",
                    "label": "Payroll Ledger",
                    "icon": "FileText",
                    "path": "/payroll",
                    "module": "payroll",
                    "required_permission": "payroll.view"
                },
                {
                    "id": 23,
                    "key": "payroll-calc",
                    "label": "Salary Engine",
                    "icon": "Calculator",
                    "path": "/payroll/calculator",
                    "module": "payroll",
                    "required_permission": "payroll.calculate"
                },
                {
                    "id": 24,
                    "key": "expenses",
                    "label": "Claims \u0026 Expenses",
                    "icon": "Receipt",
                    "path": "/finance/expenses",
                    "module": "finance",
                    "required_permission": "expense.view"
                },
                {
                    "id": 25,
                    "key": "loans",
                    "label": "Employee Loans",
                    "icon": "Landmark",
                    "path": "/finance/loans",
                    "module": "finance",
                    "required_permission": "loan.view"
                }
            ]
        },
        {
            "id": 26,
            "key": "governance-group",
            "label": "Organization Control",
            "icon": "Settings",
            "children": [
                {
                    "id": 37,
                    "key": "tenant-info",
                    "label": "Tenant Info",
                    "icon": "Info",
                    "path": "/tenant-settings/info",
                    "required_permission": "settings.manage"
                },
                {
                    "id": 27,
                    "key": "tenant-settings-general",
                    "label": "General Policies",
                    "icon": "Building2",
                    "path": "/tenant-settings",
                    "required_permission": "settings.manage"
                },
                {
                    "id": 28,
                    "key": "tenant-settings-billing",
                    "label": "Plans \u0026 Billing",
                    "icon": "CreditCard",
                    "path": "/tenant-settings/billing",
                    "required_permission": "billing.view"
                },
                {
                    "id": 29,
                    "key": "company-calendar",
                    "label": "Holiday Calendar",
                    "icon": "Calendar",
                    "path": "/tenant-settings/calendar",
                    "required_permission": "calendar.manage"
                },
                {
                    "id": 30,
                    "key": "employee-lifecycle",
                    "label": "Lifecycle Master",
                    "icon": "ListChecks",
                    "path": "/tenant-settings/lifecycle",
                    "required_permission": "lifecycle.manage"
                },
                {
                    "id": 31,
                    "key": "tenant-roles",
                    "label": "Roles \u0026 Access",
                    "icon": "ShieldAlert",
                    "path": "/tenant-settings/roles",
                    "required_permission": "role.view"
                }
            ]
        },
        {
            "id": 32,
            "key": "personal-group",
            "label": "My Personal Hub",
            "icon": "UserCog",
            "children": [
                {
                    "id": 33,
                    "key": "my-leaves",
                    "label": "Leave Request",
                    "icon": "CalendarX",
                    "path": "/leaves",
                    "module": "leave"
                },
                {
                    "id": 34,
                    "key": "my-overtime",
                    "label": "Overtime Desk",
                    "icon": "Clock",
                    "path": "/overtime",
                    "module": "overtime"
                },
                {
                    "id": 35,
                    "key": "my-payroll",
                    "label": "My Salary \u0026 Slips",
                    "icon": "Wallet",
                    "path": "/payroll",
                    "module": "payroll"
                },
                {
                    "id": 36,
                    "key": "my-timesheet",
                    "label": "My Timesheet",
                    "icon": "ActivityIcon",
                    "path": "/timesheet",
                    "module": "project"
                },
                {
                    "id": 38,
                    "key": "my-support",
                    "label": "Helpdesk",
                    "icon": "LifeBuoy",
                    "path": "/tickets",
                    "module": "support",
                    "required_permission": "support.access"
                }
            ]
        }
    ]
}
