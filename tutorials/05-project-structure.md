<!-- nav:start -->
**Tutorial 5 of 18** | [Index](00-index.md) | [Previous: 4 Go tooling jargon](04-go-tooling.md) | [Next: 6 Routes, controllers and views](06-routes-controllers-views.md)

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
18. [Cheat sheet and what to build next](18-cheatsheet-and-next-steps.md)  

</details>
<!-- nav:end -->

# 5. Project structure

```text
go-starter-pack/
  main.go                 bootstrap.Boot().Start(): web server, queue worker, scheduler
  artisan                 shortcut script: ./artisan migrate == go run . artisan migrate
  bootstrap/
    app.go                the builder: WithJobs, WithSchedule, WithCommands, WithMigrations, WithRouting ...
    commands.go jobs.go migrations.go schedule.go providers.go   lists the generators update for you
  app/
    facades/              one file per facade (Orm, Queue, Storage, ...)
    http/controllers/     handlers
    http/middleware/      middleware types
    http/responses/       Laravel-shaped JSON answers
    models/               ORM models
    services/backup/      domain logic
    console/commands/     artisan commands
    jobs/                 queue jobs
    storage/              the R2 filesystem driver
  config/                 one file per concern: database.go, queue.go, filesystems.go, backup.go ...
  database/migrations/    one file per table
  routes/                 web.go, api.go, grpc.go
  resources/views/        *.tmpl templates
  tests/                  test_case.go (boot + test-database guard), feature/*_test.go
  deploy/supervisor.conf  production processes
  docs/CONVENTIONS.md     the rules
```

## How it boots

`bootstrap/app.go` builds the application:

```go
foundation.Setup().
	WithJobs(Jobs).            // bootstrap/jobs.go
	WithSchedule(Schedule).    // bootstrap/schedule.go
	WithCommands(Commands).    // bootstrap/commands.go
	WithMigrations(Migrations).
	WithRouting(func() { routes.Web(); routes.Api(); routes.Grpc() }).
	WithProviders(Providers).
	WithConfig(config.Boot).
	Create()
```

Go has no autoloading, so a new command, job or migration must be listed in these files. The `make:*` generators do
that for you.

## Config

Each `config/*.go` file calls `facades.Config().Add("name", map[string]any{...})`, reading `.env` with
`config.Env("KEY", default)`. Read a value anywhere with `facades.Config().GetString("queue.default")`.

<!-- foot:start -->
---
[Previous: 4 Go tooling jargon](04-go-tooling.md) | [Index](00-index.md) | [Next: 6 Routes, controllers and views](06-routes-controllers-views.md)
<!-- foot:end -->
