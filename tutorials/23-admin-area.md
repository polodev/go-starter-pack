<!-- nav:start -->
**Tutorial 23 of 30** | [Index](00-index.md) | [Previous: 22 Roles and permissions](22-roles-and-permissions.md) | [Next: 24 Activity log](24-activity-log.md)

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
22. [Roles and permissions](22-roles-and-permissions.md)  
**23. [The admin area](23-admin-area.md)**  
24. [Activity log](24-activity-log.md)  
25. [Media library](25-media-library.md)  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  

</details>
<!-- nav:end -->

# 23. The admin area

- **Shell:** `admin::header` / `admin::footer`: sidebar, breadcrumb, title, status banner.
- **Sidebar:** built from every module's `AdminMenu()`; a link shows only to people who hold its permission.
- **Dashboard** (`/admin`): users, verified emails, staff, waiting and failed jobs; a card links only where you may go.
- **Users** (`/admin/users`): search, ID first, show, edit name and email (`users.manage`); with `roles.assign` also the
  role and direct exceptions.
- **Roles & permissions** (`/admin/role-permissions`): Overview, Roles, Permissions, Permission matrix, User
  assignments and Backend links (every admin route and its permission), plus the Sync button (`roles.sync`).

## An admin page of your own

```go
func AdminIndex(ctx http.Context) http.Response {
	return adminlayout.View(ctx, "shop::admin-index", "Products", []adminlayout.Crumb{{Label: "Products"}}, map[string]any{
		"products": products,
	})
}
```

```html
{{ define "shop::admin-index" }}{{ template "admin::header" . }}
  {{ template "admin::card-start" . }} <table>...</table> {{ template "admin::card-end" . }}
{{ template "admin::footer" . }}{{ end }}
```

Lists start with the ID as a link; feedback comes back as `?saved=...` or `?error=...`; every change writes an activity
entry ([24](24-activity-log.md)).

<!-- foot:start -->
---
[Previous: 22 Roles and permissions](22-roles-and-permissions.md) | [Index](00-index.md) | [Next: 24 Activity log](24-activity-log.md)
<!-- foot:end -->
