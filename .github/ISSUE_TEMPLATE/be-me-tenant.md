````
const MENUS: MenuItem[] = [
  // 1. PLATFORM ADMINISTRATION (SaaS Level)
  {
    key: "platform-group",
    label: "Platform Control",
    icon: ShieldCheck,
    roles: [ROLES.SUPERADMIN],
    children: [
      {
        key: "manage-tenants",
        label: "Tenant Directory",
        icon: Building2,
        path: "/admin/tenants",
        roles: [ROLES.SUPERADMIN],
      },
      {
        key: "subscriptions",
        label: "Global Billing",
        icon: CreditCard,
        path: "/admin/subscriptions",
        roles: [ROLES.SUPERADMIN],
      },
      {
        key: "accounts",
        label: "Platform Admins",
        icon: UserCheck,
        path: "/admin/accounts",
        roles: [ROLES.SUPERADMIN],
      },
      {
        key: "platform-roles",
        label: "System Governance",
        icon: ShieldAlert,
        path: "/admin/roles",
        roles: [ROLES.SUPERADMIN],
        permission: "platform.roles.view",
      },
      {
        key: "support-desk",
        label: "Support Desk",
        icon: MessageSquare,
        path: "/admin/support",
        roles: [ROLES.SUPERADMIN],
        permission: "support.manage",
      },
    ],
  },

  // 2. INTELLIGENCE & OVERVIEW
  {
    key: "intelligence-group",
    label: "Intelligence Hub",
    icon: LayoutDashboard,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.FINANCE, ROLES.USER],
    children: [
      {
        key: "dashboard",
        label: "Main Dashboard",
        icon: LayoutDashboard,
        path: "/",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.FINANCE, ROLES.USER],
        module: "attendance",
      },
      {
        key: "analytics",
        label: "Workforce Intel",
        icon: TrendingUp,
        path: "/analytics",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.FINANCE],
        permission: "analytics.view",
        module: "analytics",
      },
    ],
  },

  // 3. WORKFORCE MANAGEMENT (Operations)
  {
    key: "workforce-group",
    label: "Workforce Management",
    icon: Users,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
    children: [
      {
        key: "all-employees",
        label: "Staff Directory",
        icon: Users,
        path: "/employees",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "user.view",
        module: "user",
      },
      {
        key: "all-attendance",
        label: "Attendance Logs",
        icon: CalendarDays,
        path: "/attendances",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "attendance.view",
        module: "attendance",
      },
      {
        key: "work-schedules",
        label: "Shift Rosters",
        icon: Clock,
        path: "/schedules",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "schedule.view",
        module: "schedule",
      },
      {
        key: "manage-leaves",
        label: "Leave Approvals",
        icon: CalendarX,
        path: "/leaves",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "leave.view",
        module: "leave",
      },
      {
        key: "manage-overtime",
        label: "Overtime Desk",
        icon: Clock,
        path: "/overtime",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "overtime.view",
        module: "overtime",
      },
    ],
  },

  // 4. PERFORMANCE & PROJECTS (Strategic)
  {
    key: "performance-group",
    label: "Performance & Ops",
    icon: Target,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
    children: [
      {
        key: "performance-goals",
        label: "Strategic Goals",
        icon: Target,
        path: "/performance/goals",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "performance.manage",
        module: "performance",
      },
      {
        key: "performance-appraisals",
        label: "Staff Appraisals",
        icon: Star,
        path: "/performance/appraisals",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.USER],
        permission: "performance.view",
        module: "performance",
      },
      {
        key: "projects",
        label: "Project Tracker",
        icon: Briefcase,
        path: "/projects",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "project.manage",
        module: "project",
      },
      {
        key: "timesheet-monitoring",
        label: "Timesheet Audit",
        icon: ActivityIcon,
        path: "/timesheet/monitoring",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR],
        permission: "project.manage",
        module: "project",
      },
    ],
  },

  // 5. FINANCIAL HUB (Payroll & Finance)
  {
    key: "financial-group",
    label: "Financial Center",
    icon: Coins,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.FINANCE],
    children: [
      {
        key: "payroll-list",
        label: "Payroll Ledger",
        icon: FileText,
        path: "/payroll",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.FINANCE],
        permission: "payroll.view",
        module: "payroll",
      },
      {
        key: "payroll-calc",
        label: "Salary Engine",
        icon: Calculator,
        path: "/payroll/calculator",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.FINANCE],
        permission: "payroll.calculate",
        module: "payroll",
      },
      {
        key: "expenses",
        label: "Claims & Expenses",
        icon: Receipt,
        path: "/finance/expenses",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.FINANCE],
        permission: "expense.view",
        module: "finance",
      },
      {
        key: "loans",
        label: "Employee Loans",
        icon: Landmark,
        path: "/finance/loans",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.FINANCE],
        permission: "loan.view",
        module: "finance",
      },
    ],
  },

  // 6. ORGANIZATION GOVERNANCE (Settings)
  {
    key: "governance-group",
    label: "Organization Control",
    icon: Settings,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
    children: [
      {
        key: "tenant-settings-general",
        label: "General Policies",
        icon: Building2,
        path: "/tenant-settings",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
        permission: "tenant.settings.view",
      },
      {
        key: "tenant-settings-billing",
        label: "Plans & Billing",
        icon: CreditCard,
        path: "/tenant-settings/billing",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
        permission: "billing.manage",
      },
      {
        key: "company-calendar",
        label: "Holiday Calendar",
        icon: Calendar,
        path: "/tenant-settings/calendar",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
        permission: "calendar.manage",
      },
      {
        key: "employee-lifecycle",
        label: "Lifecycle Master",
        icon: ListChecks,
        path: "/tenant-settings/lifecycle",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
        permission: "lifecycle.manage",
      },
      {
        key: "tenant-roles",
        label: "Roles & Access",
        icon: ShieldAlert,
        path: "/tenant-settings/roles",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN],
        permission: "role.view",
      },
    ],
  },

  // 7. PERSONAL WORKSPACE (Employee Self-Service)
  {
    key: "personal-group",
    label: "My Personal Hub",
    icon: UserCog,
    roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.FINANCE, ROLES.USER],
    children: [
      {
        key: "my-leaves",
        label: "Leave Request",
        icon: CalendarX,
        path: "/leaves",
        roles: [ROLES.USER],
        module: "leave",
      },
      {
        key: "my-overtime",
        label: "Overtime Desk",
        icon: Clock,
        path: "/overtime",
        roles: [ROLES.USER],
        module: "overtime",
      },
      {
        key: "my-payroll",
        label: "My Salary & Slips",
        icon: Wallet,
        path: "/payroll",
        roles: [ROLES.USER],
        module: "payroll",
      },
      {
        key: "my-timesheet",
        label: "My Timesheet",
        icon: ActivityIcon,
        path: "/timesheet",
        roles: [ROLES.SUPERADMIN, ROLES.ADMIN, ROLES.HR, ROLES.FINANCE, ROLES.USER],
        module: "project",
      },
    ],
  },
];

```

sepertinya variable Static ini harus di simpan di DB agar Module tersync dengan Rapih dengan api system-module -> untuk SUPERADMIN, tenant-module -> untuk non superadmin

tolong buatkan table menus, dan seeders nya siapkan dan Refactor subscription-features jika di butuhkan dan refactor code yang harus di refactors

dan berikan task untuk FE untuk implementasi API MENUS
