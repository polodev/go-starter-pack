<!-- nav:start -->
**Tutorial 2 of 30** | [Index](00-index.md) | [Previous: 1 Introduction](01-introduction.md) | [Next: 3 Go essentials for Laravel developers](03-go-essentials.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
**02. [Setup](02-setup.md)**  
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

</details>
<!-- nav:end -->

# 2. Setup

## Tools

| Tool | Check | Install (macOS) |
|---|---|---|
| Go 1.25+ | `go version` | `brew install go` |
| MariaDB or MySQL | `mysql -uroot -e "select version()"` | `brew install mariadb` (or Herd) |
| `mysqldump` (backups) | `mysqldump --version` | comes with MariaDB/MySQL |
| Air (live reload, optional) | `air -v` | `go install github.com/air-verse/air@latest` |

## First run

```sh
cp .env.example .env
go run . artisan key:generate      # APP_KEY
go run . artisan jwt:secret        # JWT_SECRET, signs API tokens
mysql -uroot -e "CREATE DATABASE go_starter CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
mysql -uroot -e "CREATE DATABASE go_starter_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
go run . artisan migrate
go run .                           # http://127.0.0.1:3200
```

`go run .` starts the website, a queue worker and the scheduler in one process. The first run downloads modules and
compiles, so it takes a minute; later runs are quick.

## `.env`

Lines that start with `#` are off. An empty `NAME=` line would replace a built-in default with nothing, so either
fill a setting in or leave it commented. R2 and backup keys are optional; the tests that need them skip themselves.

| Setting | Meaning |
|---|---|
| `DB_CONNECTION` | `mysql` (default) or `postgres` |
| `DB_DATABASE` / `DB_TEST_DATABASE` | the app's database / the tests' database (must end in `_test`) |
| `QUEUE_CONNECTION` | `database` (jobs table) or `sync` (run at once) |
| `FILESYSTEM_DISK` | default disk: `local` or `r2_public` |
| `R2_*`, `BACKUP_*` | Cloudflare R2 keys and buckets ([12](12-files-on-r2.md), [13](13-database-backups.md)) |

## Port

The site listens on `APP_HOST:APP_PORT`, 127.0.0.1:3200 by default. If something else already uses the port, the app
stops at start-up; change `APP_PORT`.

<!-- foot:start -->
---
[Previous: 1 Introduction](01-introduction.md) | [Index](00-index.md) | [Next: 3 Go essentials for Laravel developers](03-go-essentials.md)
<!-- foot:end -->
