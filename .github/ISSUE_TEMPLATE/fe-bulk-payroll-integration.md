# Frontend Task: Bulk Payroll Generation Integration

## 📌 Context
The backend has implemented a new endpoint for **Bulk Payroll Generation**. This feature allows HR/Admin to generate payroll records for multiple employees at once, supporting various run types (Regular, THR, Bonus, All) and calculation methods (Gross, Net).

We need to update the Payroll Management UI to include a "Bulk Generate" action that interfaces with this new endpoint.

---

## 🎯 Required API Integration

### `POST /v1/payroll/bulk-generate` (NEW)
**Description:** Generates draft payroll records for a specific period and set of employees.
**Authentication:** Required (Admin / HR / Finance).

**Request Body:**
```json
{
  "period": "2026-05",        // String (Required) - Format: YYYY-MM
  "run_type": "Regular",      // String (Optional) - Pilihan: "Regular", "THR", "Bonus", "All"
  "method": "Gross",          // String (Optional) - Pilihan: "Gross", "Net"
  "user_ids": [],             // Array of Int (Optional) - Empty array = All employees in tenant
  "bonus": 0,                 // Number (Optional) - Manual bonus override
  "incentives": 0             // Number (Optional) - Manual incentives override
}
```

**API Contract (Success Response):**
```json
{
  "success": true,
  "message": "15 payroll records generated as Draft",
  "count": 15
}
```

---

## 🎨 UI/UX Requirements

### 1. Bulk Action Trigger
- Add a "Bulk Generate" button in the Payroll List view.
- If employees are selected via checkboxes in the table, the button should indicate it will process only the "Selected" employees.
- If no employees are selected, it should default to "All" employees.

### 2. Configuration Modal
When clicking "Bulk Generate", show a modal with:
- **Period Picker:** Month/Year selection (Required).
- **Calculation Method:** Radio buttons or Toggle for "Gross" vs "Net (Gross Up)".
- **Run Type:** Dropdown for "Regular", "THR", "Bonus", or "All".
- **Additional Earnings (Optional):** Input fields for `Bonus` and `Incentives` to be applied to this batch.

### 3. Feedback & States
- **Loading State:** Show a spinner or progress indicator during generation.
- **Success Notification:** Show a toast notification with the count of records generated.
- **Auto-Refresh:** Refresh the payroll list table automatically upon success to show the new "Draft" records.

---

## ✅ Success Criteria
- [ ] User can select a period and successfully trigger the generation.
- [ ] The `user_ids` array is correctly populated based on table selection.
- [ ] Calculation method "Net" correctly triggers the Gross-Up logic in backend.
- [ ] THR calculation correctly handles tenure (verified by checking results in list).
- [ ] Error messages are displayed if the request fails (e.g., period already published).

---

## 🛠️ References
- Backend PR: [Link to PR or Commit]
- Swagger Docs: `/docs/index.html#/payroll/BulkGenerate`
