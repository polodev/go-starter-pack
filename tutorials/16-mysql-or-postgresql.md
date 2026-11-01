<!-- nav:start -->
**Tutorial 16 of 32** | [Index](00-index.md) | [Previous: 15 Deployment with Supervisor](15-deployment.md) | [Next: 17 Conventions](17-conventions.md)

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
**16. [MySQL or PostgreSQL](16-mysql-or-postgresql.md)**  
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
32. [Contact form](32-contact-form.md)  

</details>
<!-- nav:end -->

# 16. MySQL or PostgreSQL

MariaDB/MySQL is the default. The PostgreSQL driver is installed too, so an app can switch with `.env`:

```sh
DB_CONNECTION=postgres
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=go_starter
DB_TEST_DATABASE=go_starter_test
DB_USERNAME=postgres
```

Unlike the Rust starter, Go needs no separate build: Goravel's ORM handles both at run time, and the migrations use
the schema builder, which writes the right SQL for each database. Backups switch from `mysqldump` to `pg_dump` by the
connection.

**Status:** the test suite has been run on MariaDB. It has not yet been run on PostgreSQL; do that before you ship a
PostgreSQL app:

```sh
createdb go_starter_test
go test ./... -count=1        # with DB_CONNECTION=postgres in .env
```

**Things that differ:** `LIKE` ignores case on MySQL but not on PostgreSQL (use `ILIKE` there for search), booleans are
real `true`/`false`, and PostgreSQL has no unsigned integers. For `psql` and `pg_dump` next to `mysql` and `mysqldump`,
see the Rust starter's tutorial 27.

<!-- foot:start -->
---
[Previous: 15 Deployment with Supervisor](15-deployment.md) | [Index](00-index.md) | [Next: 17 Conventions](17-conventions.md)
<!-- foot:end -->
