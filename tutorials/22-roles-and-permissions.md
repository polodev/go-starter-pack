<!-- nav:start -->
**Tutorial 22 of 30** | [Index](00-index.md) | [Previous: 21 Web login, verification and password reset](21-web-login.md) | [Next: 23 The admin area](23-admin-area.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
06. [Routes, controllers and views](06-routes-controllers-views.md)  
07. [Database: migrations, models, queries, seeding](07-database.md)  
08. [Validation and JSON responses](08-validation-and-json.md)  
09. [Token API: register, login, me](09-auth-api.md)  
10. [Queues and jobs](10-queues-and-jobs.md)  
11. [Console commands and the scheduler](11-console-commands-and-scheduler.md)  
12. [Files on Cloudflare R2](12-files-on-r2.md)  
13. [Database backups](13-database-backups.md)  
14. [Testing](14-testing.md)  
15. [Deployment with Supervisor](15-deployment.md)  
16. [MySQL or PostgreSQL](16-mysql-or-postgresql.md)  
17. [Conventions](17-conventions.md)  
18. [Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md)  
19. [Modules (app-modules)](19-modules.md)  
20. [Layouts, pages and translations](20-layouts-and-translations.md)  
21. [Web login, verification and password reset](21-web-login.md)  
**22. [Roles and permissions](22-roles-and-permissions.md)**  
23. [The admin area](23-admin-area.md)  
24. [Activity log](24-activity-log.md)  
25. [Media library](25-media-library.md)  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  

</details>
<!-- nav:end -->

# 22. Roles and permissions

This is the model of the role-permission module in your Laravel apps.

## Defined in code

`app-modules/role-permission/definitions.go`:

```go
var Permissions = map[string]map[string]string{
	"users": {"users.view": "View user accounts", "users.manage": "Create, edit and block user accounts"},
	...
}
newEmployee         = []string{"dashboard.view"}
businessDevelopment = with(newEmployee, "users.view", "pages.view", "media.view")
employee            = with(businessDevelopment, "pages.manage", "activity_logs.view", "ai_api_logs.view", "queue.view")
admin               = with(employee, "users.manage", "media.manage", ..., "queue.*")
developer           = with(admin, "roles.*")
```

Each role includes the one below (`with`, like `array_merge` in `config/roles.php`). After a change:

```sh
go run . artisan role-permission:sync                                   # php artisan role-permission:sync
go run . artisan role-permission:assign --email=you@example.com --role=developer
go run . artisan role-permission:audit                                  # advanced-authorization:audit
```

## The rules

- One role per person plus optional direct exceptions; `roles.*` is never an exception; the last developer stays.
- A developer holds `*`; `area.*` covers every action of an area; any grant opens the dashboard.
- `ROLE_PERMISSION_SUPER_USER_IDS` pass every check (Laravel: `ADVANCED_AUTH_SUPER_USER_IDS`).

## Admin routes fail closed

```go
rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
	admin.Get("/products", "products.view", AdminIndex)
	admin.Post("/products/{id}", "products.manage", AdminUpdate)
})
```

Every admin route names its permission. A route without one answers 403 to everyone, and `role-permission:audit`
fails, like your apps' `AdvancedRouteAuthorizationMiddleware`.

## Checking in code and templates

```go
rolepermission.HasPermission(user.ID, "products.manage")
facades.Gate().WithContext(ctx).Allows("products.view", nil) // Gate::before answers through the same check
```

```html
{{ if call .can "products.manage" }}<a href="/admin/products/{{ .ID }}/edit">Edit</a>{{ end }}
```

<!-- foot:start -->
---
[Previous: 21 Web login, verification and password reset](21-web-login.md) | [Index](00-index.md) | [Next: 23 The admin area](23-admin-area.md)
<!-- foot:end -->
