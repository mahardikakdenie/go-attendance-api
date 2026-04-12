# Backend Task: Holiday Calendar Management (CRUD)

**Title:** Implement Holiday & Corporate Event Management API
**Priority:** Medium-High
**Component:** HR Ops / Tenant Settings

## 📝 Overview
We need a robust API to manage holidays and corporate events for each tenant. These dates are critical as they affect:
1. **Attendance System**: Employees should not be marked as "Absent" on these dates.
2. **Payroll System**: Holidays marked as `is_paid: true` must be treated as working days for salary calculation.
3. **Schedules**: Overtime worked on these dates usually incurs a different multiplier.

---

## 📋 API Contract Specifications

Standard security headers (`X-Timestamp`, `X-Request-ID`, `X-Signature`) are required for all endpoints. Data must be scoped to the `tenant_id` of the authenticated user.

### 1. Get Holidays by Year
*   **Endpoint:** `GET /v1/hr/calendar`
*   **Query Params:** 
    *   `year` (int, required): e.g., `2026`
*   **Response Structure:**
    ```json
    {
      "success": true,
      "meta": {
        "message": "Calendar data fetched",
        "code": 200,
        "status": "success"
      },
      "data": [
        {
          "id": "hol-001",
          "date": "2026-12-25",
          "name": "Christmas Day",
          "type": "National Holiday",
          "is_paid": true
        }
      ]
    }
    ```

### 2. Create Holiday
*   **Endpoint:** `POST /v1/hr/calendar`
*   **Payload:**
    ```json
    {
      "date": "2026-08-17",
      "name": "Independence Day",
      "type": "National Holiday",
      "is_paid": true
    }
    ```
*   **Response:** `201 Created` with the new object.

### 3. Update Holiday
*   **Endpoint:** `PUT /v1/hr/calendar/{id}`
*   **Payload:**
    ```json
    {
      "name": "Independence Day Celebration",
      "is_paid": true
    }
    ```
*   **Response:** `200 OK`

### 4. Delete Holiday
*   **Endpoint:** `DELETE /v1/hr/calendar/{id}`
*   **Response:** `200 OK`

---

## 🔑 Data Dictionary & Logic

| Field | Type | Description |
| :--- | :--- | :--- |
| `id` | String/UUID | Unique identifier. |
| `date` | Date | The date of the event (ISO 8601: YYYY-MM-DD). |
| `name` | String | Title of the holiday (max 100 chars). |
| `type` | Enum | `National Holiday`, `Company Event`, `Mandatory Leave`, `Other`. |
| `is_paid` | Boolean | If true, payroll calculation treats this day as working hours. |

**Business Rules:**
1. **Uniqueness**: A tenant cannot have two events on the same date (or if they can, ensure UI can handle lists).
2. **Auto-validation**: When a holiday is created, the system should ideally check if any "Absent" logs for that date need to be converted to "Holiday" status.
3. **Regional Defaults**: (Bonus) Ability to seed national holidays based on the tenant's country during onboarding.

## ✅ Acceptance Criteria
- [ ] CRUD endpoints functional and strictly scoped by `tenant_id`.
- [ ] Proper error handling for invalid date formats.
- [ ] Validation to prevent duplicate entries for the same date if applicable.
- [ ] Integration test verifying `is_paid` flag is correctly stored.
