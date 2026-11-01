<!-- nav:start -->
**Tutorial 19 of 30** | [Index](00-index.md) | [Previous: 18 Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md) | [Next: 20 Layouts, pages and translations](20-layouts-and-translations.md)

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
**19. [Modules (app-modules)](19-modules.md)**  
20. [Layouts, pages and translations](20-layouts-and-translations.md)  
21. [Web login, verification and password reset](21-web-login.md)  
22. [Roles and permissions](22-roles-and-permissions.md)  
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

# 19. Modules (app-modules)

Your Laravel apps keep features in `app-modules/<name>` with internachi/modular: each is a Composer package with its own
service provider, routes, views and migrations. This starter does the same in Go.

| internachi/modular | Here |
|---|---|
| `app-modules/<name>`, a Composer package | `app-modules/<name>`, a Go package |
| `php artisan make:module shop` | `go run . artisan make:module shop` |
| service provider: `loadRoutesFrom`, `loadViewsFrom`, `loadMigrationsFrom`, `commands` | a `Module` with `Routes()`, `ViewsPath()`, `Migrations()`, `Commands()` ... |
| package discovery | `bootstrap/modules.go` lists the modules (Go has no discovery) |
| `view('shop::index')` | template `{{ define "shop::index" }}` |
| `__('shop::messages.title')` | `{{ t .lang "shop.title" }}` from `app-modules/shop/lang/en.json` |

## A module

```go
package shop

type Module struct{}

func (m *Module) Name() string      { return "shop" }
func (m *Module) ViewsPath() string { return "app-modules/shop/resources/views" }

func (m *Module) Migrations() []schema.Migration { return []schema.Migration{&migrations.M2026...CreateProductsTable{}} }
func (m *Module) Commands() []console.Command    { return []console.Command{&ImportProductsCommand{}} }
func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "Content", Label: "Products", Href: "/admin/products", Permission: "products.view"}}
}
func (m *Module) Routes() {
	web.Group(func(router route.Router) { router.Get("/shop", Index) })
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) { admin.Get("/products", "products.view", AdminIndex) })
}
```

A module implements only what it needs: `HasRoutes`, `HasMigrations`, `HasCommands`, `HasJobs`, `HasSchedule`,
`HasSeeders`, `HasViews`, `HasAdminMenu` (see `app/modular/modular.go`).

## Rules

- **No cycles.** Go refuses import cycles: `site-layout`, `auth` < `role-permission` < `admin-layout` < modules with
  admin screens. Something many modules need goes in `app/` (the activity writer is `app/activity`).
- **A module owns its tables, routes, views, translations and permissions**; others use only its exported functions.
- Add its permissions to `app-modules/role-permission/definitions.go` and run `role-permission:sync`.

<!-- foot:start -->
---
[Previous: 18 Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md) | [Index](00-index.md) | [Next: 20 Layouts, pages and translations](20-layouts-and-translations.md)
<!-- foot:end -->
