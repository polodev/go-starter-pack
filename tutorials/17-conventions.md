<!-- nav:start -->
**Tutorial 17 of 33** | [Index](00-index.md) | [Previous: 16 MySQL or PostgreSQL](16-mysql-or-postgresql.md) | [Next: 18 Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md)

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
**17. [Conventions](17-conventions.md)**  
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
33. [Sign in with Google](33-sign-in-with-google.md)  

</details>
<!-- nav:end -->

# 17. Conventions

The full rules are in `docs/CONVENTIONS.md`. The short version:

1. **Generate, then edit**: `go run . artisan make:*`.
2. **Names are full and specific**: `backup:run`, `r2_backups`, `AuthController.Register`.
3. **Slim handlers**: parse, validate, call a service or model, answer.
4. **JSON API** under `/api/v1` with Laravel's shapes; never return secrets.
5. **One migration per table**; change a shipped table with a new migration.
6. **Tests use a `*_test` database**, enforced in `config/database.go`.
7. **R2:** `r2_public` for public files, `r2_backups` for private ones, paths start with `R2_ROOT`.
8. **Queue:** `database` driver; Supervisor runs `queue:work` for busy queues.
9. **Before you finish:** `gofmt -l .` (empty), `go vet ./...`, `go test ./... -count=1`.
10. **Docs move with code**: `README.md`, `docs/CONVENTIONS.md`, the matching tutorial.
11. **Git:** short imperative subject plus a bullet list; no tool or assistant attribution; never commit `.env`.
12. **No inline scripts:** JavaScript in `public/js/`, buttons with `data-*` attributes; the Content-Security-Policy
    blocks inline code (tutorial 30).

<!-- foot:start -->
---
[Previous: 16 MySQL or PostgreSQL](16-mysql-or-postgresql.md) | [Index](00-index.md) | [Next: 18 Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md)
<!-- foot:end -->
