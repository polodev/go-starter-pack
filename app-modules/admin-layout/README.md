# admin-layout

The admin area: the shell (sidebar from every module's `AdminMenu()`, shown by permission; breadcrumb), the
dashboard, the user screens and the roles and permissions panel.

| Route | Permission |
|---|---|
| `GET /admin` | `dashboard.view` |
| `GET /admin/users`, `/admin/users/{id}` | `users.view` |
| `GET /admin/users/{id}/edit`, `POST /admin/users/{id}` | `users.manage` (role and exceptions need `roles.assign`) |
| `GET /admin/role-permissions[/roles,/roles/{slug},/permissions,/matrix,/users,/links]` | `roles.view` |
| `POST /admin/role-permissions/sync` | `roles.sync` |

Other modules render admin pages with `adminlayout.View(ctx, "module::view", title, crumbs, values)`.
