<!-- nav:start -->
**Tutorial 1 of 31** | [Index](00-index.md) | Previous: none | [Next: 2 Setup](02-setup.md)

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
31. [Queue monitor](31-queue-monitor.md)  

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
| Modules in `app-modules/` | internachi/modular | [19](19-modules.md) |
| Layouts, bilingual site, translations | Blade layouts, `__()` | [20](20-layouts-and-translations.md) |
| Web login, email verification, password reset | Breeze/Fortify | [21](21-web-login.md) |
| Token API: register, login, me | Sanctum/Passport API | [09](09-auth-api.md) |
| Roles and permissions | your Laravel apps' role-permission module | [22](22-roles-and-permissions.md) |
| Admin area: dashboard, users, roles panel | an admin panel | [23](23-admin-area.md) |
| Activity log | spatie/laravel-activitylog | [24](24-activity-log.md) |
| Media library | spatie/laravel-medialibrary | [25](25-media-library.md) |
| Mail through the queue, Amazon SES | `Mail::queue`, SES | [26](26-mail.md) |
| PDF from HTML | Browsershot / dompdf | [27](27-pdf.md) |
| AI service and AI API logs | Laravel AI SDK / Prism | [28](28-ai-service.md) |
| Bilingual pages | a CMS pages table | [29](29-bilingual-pages.md) |
| Database queue with `failed_jobs` | `QUEUE_CONNECTION=database` | [10](10-queues-and-jobs.md) |
| Queue monitor: job history, pending and failed jobs, retry | Horizon dashboard, jobs monitor | [31](31-queue-monitor.md) |
| Artisan commands and the scheduler | `make:command`, `routes/console.php` | [11](11-console-commands-and-scheduler.md) |
| Cloudflare R2 disks | `Storage::disk('s3')` | [12](12-files-on-r2.md) |
| Database backups, retention, monitoring | spatie/laravel-backup `--only-db`, `backup:clean`, `backup:monitor` | [13](13-database-backups.md) |
| Tests on a separate database | `php artisan test`, RefreshDatabase | [14](14-testing.md) |
| Supervisor deployment | Forge/Supervisor | [15](15-deployment.md) |
| Security headers (CSP, HSTS) | a headers package or Nginx config | [30](30-security-headers.md) |

## The three rules

1. **Generate, then edit.** `go run . artisan make:<thing>` writes the file and registers it.
2. **Use Goravel's batteries.** `facades.Orm()`, `facades.Queue()`, `facades.Storage()`: the framework already has it.
3. **Slim controllers.** Handlers parse, validate, call a service or model, and answer.

<!-- foot:start -->
---
Previous: none | [Index](00-index.md) | [Next: 2 Setup](02-setup.md)
<!-- foot:end -->
