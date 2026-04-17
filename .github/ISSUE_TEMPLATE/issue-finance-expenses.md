# Backend Integration: Finance Expenses & Reimbursements

## Overview
This issue tracks the implementation of backend APIs required for the Finance Expenses & Reimbursements module as seen in `src/app/(admin)/finance/expenses/page.tsx` and `src/views/finance/Expenses.tsx`.

## Endpoint Requirements

All endpoints should be prefixed with `/api/v1/finance` (or as per project convention).

### 1. GET `/expenses`
Fetch list of employee expense claims.
- **Query Parameters:**
  - `status`: Filter by status (`Pending`, `Approved`, `Rejected`).
  - `search`: Filter by Claim ID or Employee Name.
  - `page`, `limit`: Pagination parameters.
- **Response Structure:**
  ```json
  {
    "data": [
      {
        "id": "EXP-001",
        "employeeName": "Alex Johnson",
        "avatar": "url_to_image",
        "category": "Travel",
        "amount": 450000,
        "date": "2024-03-15",
        "description": "Taxi to client office",
        "status": "Pending",
        "receiptUrl": "url_to_receipt_file"
      }
    ],
    "meta": {
      "total": 100,
      "page": 1,
      "lastPage": 10
    }
  }
  ```

### 2. GET `/expenses/summary`
Fetch summary statistics for the dashboard cards.
- **Response Structure:**
  ```json
  {
    "pendingAmount": 4200000,
    "approvedThisMonthAmount": 12800000,
    "topCategory": {
      "name": "Travel",
      "percentage": 60
    }
  }
  ```

### 3. POST `/expenses`
Submit a new expense claim.
- **Request Body:**
  ```json
  {
    "category": "Travel",
    "amount": 250000,
    "date": "2024-03-16",
    "description": "Lunch with client",
    "receipt": "binary_file_or_url" 
  }
  ```

### 4. PATCH `/expenses/:id/approve`
Approve a pending expense claim.
- **Response:** `200 OK`

### 5. PATCH `/expenses/:id/reject`
Reject a pending expense claim.
- **Request Body:**
  ```json
  {
    "reason": "Missing receipt details"
  }
  ```
- **Response:** `200 OK`

## Data Models

### ExpenseStatus (Enum)
- `Pending`
- `Approved`
- `Rejected`

### ExpenseCategory (Enum)
- `Travel`
- `Medical`
- `Supplies`
- `Equipment`
- `Other`

## Security & Validation
- Ensure only users with `Finance` or `Admin` roles can approve/reject claims.
- Validate receipt attachments if required.
- All requests must include standard security headers (`X-Timestamp`, `X-Request-ID`, `X-Signature`) as handled by the Next.js proxy.

## Frontend Reference
- View: `src/views/finance/Expenses.tsx`
- Types: `src/types/finance.ts`
