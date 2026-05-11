# Backend Task: Dynamic Subscription Features Management

## 📌 Context
Currently, the subscription features available for assignment to plans are hardcoded in the frontend. To make the system scalable and allow the addition of new modules without frontend deployments, we need to manage these features in the backend database.

## 🎯 Database Requirements
Create a new table `subscription_features` to store available system modules that can be included in a subscription plan.

### Table: `subscription_features`
| Column | Type | Description |
| :--- | :--- | :--- |
| `id` | BigInt (PK) | Unique identifier |
| `feature_key` | String (Unique) | String key used by the code (e.g., `payroll`, `attendance`) |
| `label` | String | Display name (e.g., `Payroll & Slips`) |
| `description` | Text (Optional) | Brief explanation of the module |
| `is_active` | Boolean | Whether this feature is currently available for selection |
| `created_at` | Timestamp | |
| `updated_at` | Timestamp | |

## 🌱 Mandatory Seeder
Please provide a seeder to populate the initial system features based on the current frontend requirements:

| feature_key | label |
| :--- | :--- |
| `user` | Employee Management |
| `attendance` | Advanced Attendance |
| `leave` | Leave Requests |
| `overtime` | Overtime Tracking |
| `payroll` | Payroll & Slips |
| `finance` | Finance & Claims |
| `analytics` | Advanced Analytics |
| `timesheet` | Project Timesheet |

---

## 🚀 Required API Endpoints (Frontend Expectation)

### 1. `GET /v1/superadmin/subscription-features`
**Description:** Fetches all available subscription features for the Plan Management UI.
**Expected Response:**
```json
{
  "meta": {
    "code": 200,
    "status": "success",
    "message": "Features retrieved successfully"
  },
  "data": [
    {
      "id": 1,
      "feature_key": "user",
      "label": "Employee Management",
      "is_active": true
    },
    {
      "id": 2,
      "feature_key": "attendance",
      "label": "Advanced Attendance",
      "is_active": true
    }
  ]
}
```

### 2. Update `GET /v1/superadmin/plans`
**Description:** Ensure that when fetching plans, the `features` array returns the `feature_key` of the assigned features.

---

## 🛠️ Actions Required
1. Run migration for `subscription_features` table.
2. Execute the seeder with the values provided above.
3. Implement the `GET` endpoint for features.
4. (Optional) Provide a simple CRUD for these features if needed in the future.
