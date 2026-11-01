<!-- nav:start -->
**Tutorial 5 of 30** | [Index](00-index.md) | [Previous: 4 Go tooling jargon](04-go-tooling.md) | [Next: 6 Routes, controllers and views](06-routes-controllers-views.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
**05. [Project structure](05-project-structure.md)**  
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

# 5. Project structure

```text
go-starter-pack/
  main.go                  bootstrap.Boot().Start(): web server, queue worker, scheduler
  artisan                  shortcut: ./artisan migrate == go run . artisan migrate
  app-modules/             the features, one module each (like internachi/modular)
    site-layout/ auth/ role-permission/ admin-layout/ activity-log/ media/ pdf/ ai-service/ pages/
      module.go            Name() plus routes, migrations, commands, jobs, schedule, seeders, views, admin menu
      migrations/ models/  its tables
      resources/views/     templates named "<module>::<view>"
      lang/en.json bn.json its translations, read as "<module>.<key>"
  app/
    modular/               the module system (Module, Collect, Menu, LangFS)
    web/                   web middleware group, page data, templates, translation helpers
    activity/              activity log writer any module can use
    storage/               the Cloudflare R2 filesystem driver
    services/backup/       database backups
    console/commands/      app-wide artisan commands (backup, queue:work, make:module)
    http/responses/        Laravel-shaped JSON answers
    models/                User
    support/               ses (SES SMTP password), aifake (AI provider for tests)
    facades/               one file per Goravel facade
  bootstrap/
    app.go                 the builder: WithJobs, WithSchedule, WithSeeders, WithCommands, WithMigrations, WithRouting ...
    modules.go             the module list (make:module adds to it)
    commands.go jobs.go migrations.go schedule.go seeders.go providers.go
  config/                  one file per concern: database, queue, filesystems, mail, ai, backup ...
  database/migrations/     app-wide tables (users, jobs, failed_jobs)
  routes/                  web.go (static files); pages live in their modules
  resources/css/           Tailwind input; public/css/app.css is the built file
  tests/                   test_case.go (boot + test-database check), feature/*_test.go
  deploy/supervisor.conf   production processes
  docs/CONVENTIONS.md      the rules
```

## How it boots

`bootstrap/app.go` builds the application. Each `With...` gets the app's own items plus every module's:

```go
foundation.Setup().
	WithJobs(Jobs).WithSchedule(Schedule).WithSeeders(Seeders).WithCommands(Commands).WithMigrations(Migrations).
	WithRouting(func() { routes.Web(); for _, module := range Modules() { module.Routes() } ... }).
	WithProviders(Providers).WithConfig(config.Boot).
	WithCallback(bootModules). // registers each module's views
	Create()
```

Go has no autoloading or package discovery, so new things must be listed. The generators (`make:command`,
`make:job`, `make:migration`, `make:module`) do it for you.

## Config

Each `config/*.go` file calls `facades.Config().Add("name", map[string]any{...})`, reading `.env` with
`config.Env("KEY", default)`. Read a value anywhere with `facades.Config().GetString("queue.default")`.

<!-- foot:start -->
---
[Previous: 4 Go tooling jargon](04-go-tooling.md) | [Index](00-index.md) | [Next: 6 Routes, controllers and views](06-routes-controllers-views.md)
<!-- foot:end -->
