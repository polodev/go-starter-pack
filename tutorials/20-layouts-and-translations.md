<!-- nav:start -->
**Tutorial 20 of 32** | [Index](00-index.md) | [Previous: 19 Modules (app-modules)](19-modules.md) | [Next: 21 Web login, verification and password reset](21-web-login.md)

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
**20. [Layouts, pages and translations](20-layouts-and-translations.md)**  
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
31. [Queue monitor](31-queue-monitor.md)  
32. [Contact form](32-contact-form.md)  

</details>
<!-- nav:end -->

# 20. Layouts, pages and translations

## The web middleware group

Laravel's `web` group (sessions, CSRF, locale) is `web.Group`:

```go
web.Group(func(router route.Router) {
	router.Get("/shop", Index)
	router.Middleware(&auth.Authenticate{}).Get("/shop/orders", Orders)
})
```

## Rendering a page

```go
func Index(ctx http.Context) http.Response {
	return web.View(ctx, "shop::index", map[string]any{"title": web.T(ctx, "shop.title"), "products": products})
}
```

`web.View` adds what every page needs: `lang`, `csrf`, `path`, `appName`, `user`, `can`, `avatarURL`.
`web.ViewStatus(ctx, 422, ...)` renders with another status (a form with errors, a 404).

```html
{{ define "shop::index" }}{{ template "site::header" . }}
  <h1>{{ .title }}</h1>
  {{ range .products }}<p>{{ .Name }}</p>{{ end }}
{{ template "site::footer" . }}{{ end }}
```

Go's `html/template` escapes everything, like Blade's `{{ }}`. Templates are compiled on the first request: restart
the server after changing one (or use `air`).

## Translations

Every module has `lang/en.json` and `lang/bn.json`; Goravel's translator reads them through `modular.LangFS`:

```json
{ "title": "Our shop", "welcome": "Hello :name" }
```

| Where | How |
|---|---|
| template | `{{ t .lang "shop.title" }}`, `{{ tr .lang "shop.welcome" "name" .user.Name }}` |
| Go, in a request | `web.T(ctx, "shop.title")` |
| Go, outside a request (an email) | `web.Translate("bn", "shop.title")` |

The language comes from `?lang=bn`, the `lang` cookie (set by `/language/bn?next=/shop`), then `Accept-Language`.
Add every key to both files. Admin screens are English only.

<!-- foot:start -->
---
[Previous: 19 Modules (app-modules)](19-modules.md) | [Index](00-index.md) | [Next: 21 Web login, verification and password reset](21-web-login.md)
<!-- foot:end -->
