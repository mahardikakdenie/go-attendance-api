# Backend Task: HR Operational Dashboard Integration (Manager Home)

## 📝 Overview
Dashboard operasional untuk level Manager (HR/Admin) di halaman utama membutuhkan data real-time terkait "Pulse" organisasi hari ini. Berbeda dengan Analytics bulanan, data di sini bersifat harian dan interaktif.

---

## 🛠️ API Requirements

### 1. Daily Organization Pulse
- **Endpoint**: `GET /api/v1/dashboards/hr/daily-pulse`
- **Response**:
```json
{
  "stats": {
    "present_percentage": 94.2,
    "avg_overtime_hours": 2.5,
    "pending_approvals_count": 12,
    "at_risk_count": 3
  },
  "hotline_requests": [
    {
      "id": "LV-102",
      "user_name": "Bagus Fikri",
      "avatar": "...",
      "department": "Engineering",
      "request_type": "Leave",
      "priority": "High"
    }
  ],
  "top_performers": [
    { "name": "Sarah", "avatar": "...", "department": "HR", "score": 98 }
  ]
}
```

---

## 🚀 Key Integrations in Frontend
UI sudah disiapkan di `src/views/dashboard/ManagerDashboard.tsx` dan saat ini memanggil `getHrDashboard()`. 

**Action Item:**
1. Update `HrDashboardData` di `api.ts` jika ada field tambahan dari response di atas.
2. Pastikan `fetchData` di `ManagerDashboard.tsx` memanggil endpoint harian yang baru jika performa `getHrDashboard()` (bulanan) dirasa berat untuk halaman Home.

---
**Status**: 🎨 UI Selesai | 🔄 Menunggu API Spesifik Harian
