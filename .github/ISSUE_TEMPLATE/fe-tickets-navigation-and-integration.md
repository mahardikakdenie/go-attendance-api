# Frontend Task: Tickets Navigation + Integration (`/menus/me` + `/tickets`)

## 📌 Context
Backend already added Helpdesk navigation into `GET /api/v1/menus/me` as a permission-driven menu item:

- `key`: `my-support`
- `label`: `Helpdesk`
- `path`: `/tickets`
- `required_permission`: `support.access`

This means sidebar visibility is no longer hardcoded by role. FE must consume menu from API and route users to the new Tickets page.

---

## ✅ Backend Status (Confirmed)
Menu entry exists in backend seeder:
- File: `internal/seeder/menu_seeder.go`
- Entry:
  - `Key: "my-support"`
  - `Path: "/tickets"`
  - `RequiredPermission: "support.access"`

So yes, `/menus/me` now includes navigation data for this page, **when user has permission**.

---

## 🎯 FE Tasks

### 1) Sidebar Navigation
- Ensure sidebar renderer from `/api/v1/menus/me` handles menu item:
  - `key: my-support`
  - `path: /tickets`
- Do not hardcode role checks for this menu.
- Visibility must be fully controlled by backend permission output.

### 2) Router Setup
- Add route page for:
  - `/tickets`
- If app still uses `/support`, add redirect:
  - `/support` -> `/tickets`

### 3) Tickets Create Form Integration
Use backend endpoint:
- `POST /api/v1/tickets`

Payload:
```json
{
  "subject": "Gagal Clock In",
  "category": "TECHNICAL",
  "priority": "HIGH",
  "message": "Saya tidak bisa clock in sejak pagi",
  "attachment_url": "https://..." 
}
```

### 4) Category & Priority Source
Replace FE hardcoded options with API source:
- `GET /api/v1/tickets/categories`
- `GET /api/v1/tickets/priorities`

### 5) User Ticket History
Integrate list page with:
- `GET /api/v1/tickets/history`

Supported query params:
- `search`
- `status`
- `priority`
- `limit`
- `offset`

### 6) User Reply Flow
Integrate reply action using:
- `POST /api/v1/tickets/{id}/reply`

Payload:
```json
{ "message": "Saya sudah coba saran sebelumnya, masih error." }
```

---

## 🔒 Permission Behavior
If `support.access` is absent:
- Menu `Helpdesk` should not appear from `/menus/me`.
- User should not be able to access `/tickets` directly.

---

## ✅ Success Criteria
- [ ] Sidebar shows `Helpdesk` menu when backend returns `my-support`.
- [ ] Clicking menu navigates to `/tickets` page.
- [ ] Ticket form submits to `POST /tickets` successfully.
- [ ] Category/Priority dropdown populated from API.
- [ ] Ticket history and reply features integrated.
- [ ] No hardcoded role logic for Helpdesk visibility.
