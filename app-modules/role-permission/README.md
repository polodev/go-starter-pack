# role-permission

Staff authorization, the model of the Laravel apps' role-permission module.

- `definitions.go`: permissions by area and five ranked roles (new_employee < business_development < employee <
  admin < developer), each including the one below.
- `go run . artisan role-permission:sync`: copies the definitions into `roles`, `permissions`, `permission_role`;
  removed ones are switched off.
- `go run . artisan role-permission:assign --email=you@example.com --role=developer`: the first developer.
- `go run . artisan role-permission:audit`: fails when an `/admin` route has no permission.
- One role per person (`role_user`) plus direct exceptions (`permission_user`); never `roles.*` as an exception.
- `rolepermission.Admin(func(admin *rolepermission.AdminRouter) { admin.Get("/x", "x.view", handler) })` declares
  admin routes; undeclared ones answer 403 to everyone.
- `ROLE_PERMISSION_SUPER_USER_IDS` pass every check. Goravel's Gate answers through `HasPermission`.
