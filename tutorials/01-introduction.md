<!-- nav:start -->
**Tutorial 1 of 18** | [Index](00-index.md) | Previous: none | [Next: 2 Setup](02-setup.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
**01. [Introduction](01-introduction.md)**  
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
18. [Cheat sheet and what to build next](18-cheatsheet-and-next-steps.md)  

</details>
<!-- nav:end -->

# 1. Introduction

This starter pack gives you Laravel's way of working in Go. It is built on [Goravel](https://www.goravel.dev) v1.18,
a Go framework that copies Laravel's structure and names: `artisan`, facades, migrations, an ORM, queues, a scheduler,
validation and storage disks. It is the sister of the Rust starter pack and follows the same rules.

**Why Go:** one small binary per app, fast start-up, low memory, simple deployment (copy the binary and `.env`), and a
language that is quick to learn. Goravel keeps the Laravel vocabulary, so you think in the same terms.

## What is built

| Part | Laravel equivalent | Tutorial |
|---|---|---|
| Token API: register, login, me | Sanctum/Passport API | [09](09-auth-api.md) |
| Database queue with `failed_jobs` | `QUEUE_CONNECTION=database` | [10](10-queues-and-jobs.md) |
| Artisan commands and the scheduler | `make:command`, `routes/console.php` | [11](11-console-commands-and-scheduler.md) |
| Cloudflare R2 public and private disks | `Storage::disk('s3')` | [12](12-files-on-r2.md) |
| Database backups to the private bucket | spatie/laravel-backup `--only-db` | [13](13-database-backups.md) |
| Tests on a separate database, real R2 tests | `php artisan test`, RefreshDatabase | [14](14-testing.md) |
| Supervisor deployment | Forge/Supervisor | [15](15-deployment.md) |

Not built yet, but in the Rust starter: web login pages, roles and permissions, the admin dashboard, media library,
activity log, bilingual pages, mail queue, PDF and the AI service. [Tutorial 18](18-cheatsheet-and-next-steps.md)
lists them and how to add each one.

## The three rules

1. **Generate, then edit.** `go run . artisan make:<thing>` writes the file and registers it.
2. **Use the facades.** `facades.Orm()`, `facades.Queue()`, `facades.Storage()`: the framework already has it.
3. **Slim controllers.** Handlers parse, validate, call a service or model, and answer.

<!-- foot:start -->
---
Previous: none | [Index](00-index.md) | [Next: 2 Setup](02-setup.md)
<!-- foot:end -->
