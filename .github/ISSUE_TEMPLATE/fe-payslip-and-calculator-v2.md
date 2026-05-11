# Frontend Task: Payslip Display & Smart Payroll Calculator (v2)

## 📌 Context
The backend payroll engine has been upgraded. **All API request keys have been standardized to snake_case.** 

The calculator is now "Smart": by sending a `user_id`, the backend automatically pulls profile data (salary, bank info, position, company logo), so the Frontend needs to send less data.

---

## 🎯 API Contract Updates (IMPORTANT)

### 1. `POST /v1/payroll/calculate` (The Engine)
⚠️ **BREAKING CHANGE:** You MUST use `snake_case`. Requests using `camelCase` will return 0 results.

**Request Payload Example:**
```json
{
  "user_id": 7,                     // ID employee yang dipilih
  "run_type": "Regular",            // "Regular", "THR", "Bonus", "All"
  "method": "Net",                  // "Gross" (Potong Pajak) or "Net" (Gross-up)
  "working_days_in_month": 22,      // Hari kerja sebulan
  "attendance_days": 20,            // Hari hadir
  "overtime_hours": 10.5,           // Total jam lembur
  "unpaid_leave_days": 0,           // Hari cuti tidak berbayar
  "bonus": 1000000,                 // Opsional: Bonus tambahan
  "incentives": 500000              // Opsional: Insentif tambahan
}
```

### 2. Standardized Response (Automatic Context)
The response now includes everything you need to render a **Slip Gaji Preview** immediately:
```typescript
interface CalculationResponse {
  success: boolean;
  data: {
    run_type: string;
    method: string;
    company_context: {
        name: string;
        logo_url: string;
    };
    user: {
        full_name: string;
        employee_id: string;
        position: string;
        bank_name: string;
        bank_account_number: string;
    };
    breakdown: {
        earnings: {
            basic_salary: number;
            fixed_allowances: number;
            variable_allowances: number;
            overtime_pay: number;
            incentives: number;
            bonus: number;
            thr: number;
            tax_allowance?: number; // Only for "Net"
            bpjs_allowance?: number; // Only for "Net"
            gross_income: number;
        };
        deductions: {
            pph21_amount: number;
            bpjs_health_employee: number;
            bpjs_jht_employee: number;
            bpjs_jp_employee: number;
            unpaid_leave_deduction: number;
            total_deductions: number;
        };
    };
    net_salary: number;
  };
}
```

---

## 🎨 UI/UX Recommendations

### 1. The Payslip (Two-Column Layout)
- **Left Column (Income):** List all fields from `breakdown.earnings`.
- **Right Column (Deductions):** List all fields from `breakdown.deductions`.
- **Footer:** Display `net_salary` prominently in a large, bold font.

### 2. The Smart Calculator Modal
- **Dynamic Fields:** If `run_type === 'THR'`, **disable** the `attendance_days` and `overtime_hours` inputs to prevent confusion.
- **Method Toggle:** Provide a clear switch between **Gross** and **Net**. 
- **Real-time Preview:** As the user types (use debounce), hit the `/calculate` API and show the breakdown preview on the right side of the modal.

---

## ✅ Success Criteria
- [ ] UI correctly uses `snake_case` in all payroll POST requests.
- [ ] Preview slip displays company logo and employee bank info fetched from backend.
- [ ] Tax Allowance is visible when selecting the "Net" calculation method.
- [ ] No `any` types used in the Payroll Redux/Zustand state.
