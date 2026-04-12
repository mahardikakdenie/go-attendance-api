# Backend Task: Advanced HR Operations (Shifts, Calendar & Lifecycle)

## 📝 Overview
To support enterprise-grade workforce management, the backend needs to implement three core operational modules: Shift Management, Public Holidays, and Employee Checklists. The Frontend (FE) has already prepared the UI and types based on the following contracts.

---

## 🛠️ 1. Shift & Schedule Management

### A. Shift Templates
Endpoints to manage master data for available work shifts.

- **GET** `/api/v1/hr/shifts`
- **Response**: `APIResponse<WorkShift[]>`
```json
[
  {
    "id": "uuid-s1",
    "name": "Morning Shift",
    "startTime": "06:00",
    "endTime": "14:00",
    "type": "Morning",
    "color": "bg-emerald-500",
    "isDefault": false
  }
]
```

### B. Weekly Rostering
Endpoints to fetch and save assignments for a specific week.

- **GET** `/api/v1/hr/roster?start_date=2024-04-15&end_date=2024-04-21&department_id=optional`
- **Response**: `APIResponse<EmployeeSchedule[]>`
```json
[
  {
    "id": 1,
    "name": "Bagus Fikri",
    "avatar": "...",
    "department": "Engineering",
    "weeklyRoster": {
      "monday": "uuid-s1",
      "tuesday": "uuid-s1",
      "wednesday": "off",
      "thursday": "uuid-s1",
      "friday": "uuid-s1",
      "saturday": "off",
      "sunday": "off"
    }
  }
]
```

- **POST** `/api/v1/hr/roster/save`
- **Payload**:
```json
{
  "start_date": "2024-04-15",
  "assignments": [
    {
      "user_id": 1,
      "roster": {
        "monday": "uuid-s1",
        "tuesday": "uuid-s1",
        "wednesday": "off",
        "thursday": "uuid-s1",
        "friday": "uuid-s1",
        "saturday": "off",
        "sunday": "off"
      }
    }
  ]
}
```

---

## 🛠️ 2. Company & Holiday Calendar

- **GET** `/api/v1/hr/calendar?year=2024`
- **Response**:
```json
[
  {
    "id": "uuid-h1",
    "date": "2024-04-10",
    "name": "Idul Fitri 1445H",
    "type": "NATIONAL_HOLIDAY",
    "is_paid": true
  }
]
```

- **POST** `/api/v1/hr/calendar`
- **Payload**: `{"date": "2024-08-17", "name": "Independence Day", "type": "NATIONAL_HOLIDAY"}`

**Logic**: The attendance engine MUST check this table. If a date is a holiday, users should NOT be marked as "Absent" if they don't clock in.

---

## 🛠️ 3. Employee Lifecycle (Onboarding/Offboarding)

- **GET** `/api/v1/hr/employees/{id}/lifecycle`
- **Response**:
```json
{
  "employee_id": 1,
  "status": "ONBOARDING",
  "tasks": [
    {
      "id": "task-1",
      "task_name": "Laptop Handover",
      "category": "ONBOARDING",
      "is_completed": true,
      "completed_at": "2024-04-01T10:00:00Z"
    },
    {
      "id": "task-2",
      "task_name": "ID Card Creation",
      "category": "ONBOARDING",
      "is_completed": false,
      "completed_at": null
    }
  ]
}
```

- **PATCH** `/api/v1/hr/employees/{id}/lifecycle/tasks/{task_id}`
- **Payload**: `{"is_completed": true}`

---

## 🚀 Key Integrations Status
- `/schedules` -> **UI Ready** (Dummy Integrated)
- `/tenant-settings/calendar` -> **Placeholder Ready**
- `/tenant-settings/lifecycle` -> **Placeholder Ready**

---
**Documented by**: CEO (Gemini CLI Agent)
**Alignment**: Synced with `src/types/schedules.ts`
