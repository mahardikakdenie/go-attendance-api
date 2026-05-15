# System Permissions Reference

This document lists all available permission keys in the Go-Attendance API system. These keys are used for Role-Based Access Control (RBAC) and to control menu visibility.

## Module: Attendance & Monitoring (`attendance`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `attendance.view` | view | View attendance records and logs |
| `attendance.create` | create | Clock-in/out and create manual logs |
| `attendance.edit` | edit | Edit existing attendance records |
| `attendance.delete` | delete | Delete attendance records |
| `attendance.export` | export | Export attendance data to CSV/Excel |

## Module: Leave Management (`leave`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `leave.view` | view | View leave requests and balances |
| `leave.create` | create | Submit new leave requests |
| `leave.approve` | approve | Approve pending leave requests |
| `leave.reject` | reject | Reject pending leave requests |

## Module: Overtime & Extra Hours (`overtime`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `overtime.view` | view | View overtime logs and requests |
| `overtime.create` | create | Submit overtime requests |
| `overtime.approve` | approve | Approve overtime hours |

## Module: Payroll & Finance (`payroll`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `payroll.view` | view | View payroll records and slips |
| `payroll.calculate` | calculate | Run payroll calculations |
| `payroll.approve` | approve | Approve payroll for disbursement |
| `payroll.edit` | edit | Edit payroll entries and adjustments |

## Module: User Management (`user`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `user.view` | view | View user profiles and directory |
| `user.view.detail` | view | View granular user details (DNA, documents) |
| `user.create` | create | Create new user accounts |
| `user.edit` | edit | Update user information |
| `user.delete` | delete | Deactivate/Delete user accounts |

## Module: Organization & SaaS (`tenant`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `tenant.view` | view | View organization details |
| `tenant.edit` | edit | Edit organization information |
| `tenant.settings.view` | view | View tenant-specific settings and policies |
| `billing.manage` | manage | Manage organization billing and plans |
| `calendar.manage` | manage | Manage holiday calendars and work days |
| `lifecycle.manage` | manage | Manage employee lifecycle and transitions |

## Module: Plans & Billing (`subscription`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `subscription.manage` | manage | Manage system-wide subscription plans (Superadmin) |

## Module: Roles & Permissions (`role`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `role.view` | view | View available roles |
| `role.manage` | manage | Create and update roles and permissions |
| `platform.roles.view` | view | View system-level roles (Superadmin) |

## Module: Support & Helpdesk (`support`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `support.manage` | manage | Access support tickets and desk |

## Module: Analytics & Reports (`analytics`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `analytics.view` | view | Access workforce analytics and dashboards |

## Module: Work Schedules (`schedule`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `schedule.view` | view | View work shifts and rosters |

## Module: Project Management (`project`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `project.view` | view | View project assignments |
| `project.manage` | manage | Manage projects and team assignments |

## Module: Time Tracking (`timesheet`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `timesheet.view` | view | View timesheet entries |
| `timesheet.create` | create | Submit timesheet entries |
| `timesheet.manage` | manage | Audit and manage team timesheets |

## Module: Finance Operations (`finance`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `finance.view` | view | View financial transactions |
| `finance.manage` | manage | Manage claims and expenses |
| `expense.view` | view | View expense claims |
| `loan.view` | view | View employee loans |

## Module: Performance & Goals (`performance`)
| Permission ID | Action | Description |
|---------------|--------|-------------|
| `performance.view` | view | View performance reviews |
| `performance.manage` | manage | Manage strategic goals and appraisals |
