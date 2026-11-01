<!-- nav:start -->
**Tutorial 18 of 18** | [Index](00-index.md) | [Previous: 17 Conventions](17-conventions.md) | Next: none

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
**18. [Cheat sheet and what to build next](18-cheatsheet-and-next-steps.md)**  

</details>
<!-- nav:end -->

# 18. Cheat sheet and what to build next

## Cheat sheet

| Laravel | Go starter pack |
|---|---|
| `php artisan serve` | `go run .` / `air` |
| `php artisan make:controller X` | `go run . artisan make:controller X` |
| `php artisan make:model X -m` | `make:model X` + `make:migration create_xs_table` |
| `php artisan migrate` / `migrate:fresh` | the same names |
| `php artisan db:seed` | `go run . artisan db:seed` |
| `php artisan make:job X` / `X::dispatch()` | `make:job X` / `facades.Queue().Job(&jobs.X{}, args).Dispatch()` |
| `php artisan queue:work` | `go run . artisan queue:work --queue=default` |
| `php artisan schedule:list` | the same |
| `Storage::disk('s3')->put()` | `facades.Storage().Disk("r2_public").Put()` |
| `Hash::make()` / `Hash::check()` | `facades.Hash().Make()` / `.Check()` |
| `$request->validate([...])` | `ctx.Request().Validate(map[string]any{...})` |
| `auth()->user()` | `facades.Auth(ctx).User(&user)` |
| `config('queue.default')` | `facades.Config().GetString("queue.default")` |
| `php artisan test` | `go test ./... -count=1` |
| `composer require x` | `go get x` then `go mod tidy` |

## What to build next (the Rust starter has these)

| Feature | Start with |
|---|---|
| Web login pages and sessions | `config/session.go`, a session guard in `config/auth.go`, views in `resources/views` |
| Email verification, password reset, sign-in links | a `user_security_tokens` table, `facades.Mail()`, queued mail |
| Login throttling | `facades.RateLimiter()` |
| Roles and permissions | `facades.Gate()` and policies (`make:policy`) |
| Admin dashboard | a layout template and `/admin` routes behind a permission middleware |
| Media library (Spatie-compatible) | the `media` table and the `r2_public` disk |
| Activity log | Spatie's `activity_log` table and an observer (`make:observer`) |
| Bilingual pages (English, Bangla) | `lang/` files and `facades.Lang(ctx)` |
| Mail queue and SES | `config/mail.go` and a job per mail |
| PDF | headless Chrome (chromedp) |
| AI service | Goravel's AI SDK (`config/ai.go` is already there) |
| Backup retention and monitor | `backup:clean`, `backup:monitor` next to `backup:run` |

Add each the same way: generate, write the test, update `README.md`, `docs/CONVENTIONS.md` and a tutorial.

<!-- foot:start -->
---
[Previous: 17 Conventions](17-conventions.md) | [Index](00-index.md) | Next: none
<!-- foot:end -->
