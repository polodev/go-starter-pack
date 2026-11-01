<!-- nav:start -->
**Tutorial 24 of 30** | [Index](00-index.md) | [Previous: 23 The admin area](23-admin-area.md) | [Next: 25 Media library](25-media-library.md)

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
23. [The admin area](23-admin-area.md)  
**24. [Activity log](24-activity-log.md)**  
25. [Media library](25-media-library.md)  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  

</details>
<!-- nav:end -->

# 24. Activity log

```go
// Laravel: activity()->causedBy($admin)->performedOn($user)->event('updated')->withProperties(...)->log('updated user');
activity.In("users").CausedBy(admin.ID).PerformedOn("user", user.ID).Event("updated").
	Changes(before, after).LogQuietly("updated user")
```

- The table is Spatie's `activity_log`, so a Laravel app reads the same rows. `Changes` stores
  `{"attributes": new, "old": old}` with only the fields that changed. Never put passwords or tokens in it.
- `LogQuietly` never breaks the action: a failure is written to the app log (`facades.Log()`).
- Admin: `/admin/activity-logs` with filters (log, event, causer, text) and a show page with the changes
  (`activity_logs.view`); delete and "Remove old entries" (`activity_logs.manage`).
- `go run . artisan activity-log:clean --days=365` runs daily (`ACTIVITY_LOG_RETENTION_DAYS`).

Goravel's logging (`facades.Log()`, daily files in `storage/logs`) is for application errors; the activity log is the
audit trail people read in the admin.

<!-- foot:start -->
---
[Previous: 23 The admin area](23-admin-area.md) | [Index](00-index.md) | [Next: 25 Media library](25-media-library.md)
<!-- foot:end -->
