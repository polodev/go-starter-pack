<!-- nav:start -->
**Tutorial 12 of 33** | [Index](00-index.md) | [Previous: 11 Console commands and the scheduler](11-console-commands-and-scheduler.md) | [Next: 13 Database backups](13-database-backups.md)

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
**12. [Files on Cloudflare R2](12-files-on-r2.md)**  
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
32. [Contact form](32-contact-form.md)  
33. [Sign in with Google](33-sign-in-with-google.md)  

</details>
<!-- nav:end -->

# 12. Files on Cloudflare R2

| Disk | Bucket | Use |
|---|---|---|
| `local` | `storage/app` | development files |
| `r2_public` | `R2_PUBLIC_BUCKET` | uploads anyone may read; `Url()` gives the `R2_PUBLIC_URL` address |
| `r2_backups` | `BACKUP_BUCKET` (own keys) | database backups; no public address |

```go
disk := facades.Storage().Disk("r2_public")
err := disk.Put("go-starter-pack/avatars/7.jpg", content)   // start paths with R2_ROOT
url := disk.Url("go-starter-pack/avatars/7.jpg")            // https://pub-....r2.dev/go-starter-pack/avatars/7.jpg
exists := disk.Exists("go-starter-pack/avatars/7.jpg")
err = disk.Delete("go-starter-pack/avatars/7.jpg")
```

An upload from a form: `file, err := ctx.Request().File("avatar")`, then `disk.PutFileAs("go-starter-pack/avatars", file, "7.jpg")`.

## Why there is a custom driver

R2 speaks the S3 API, but Goravel's S3 driver has no endpoint setting, so both disks use the official MinIO driver.
That driver builds path-style addresses (`https://host/<bucket>/<file>`), while an R2 public address (`r2.dev` or a
custom domain) is bound to one bucket and serves `https://host/<file>`. `app/storage/r2.go` wraps the MinIO driver
and fixes only `Url()`. The MinIO driver also insists on a `url` setting, so `config/filesystems.go` puts the bucket's
private S3 address there and keeps the real public address in `public_url`.

## Settings

`R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_ENDPOINT` (`https://<account>.r2.cloudflarestorage.com`),
`R2_PUBLIC_BUCKET`, `R2_PUBLIC_URL`, `R2_ROOT`; and the `BACKUP_*` set for the private bucket.

## What the tests prove

`tests/feature/r2_storage_test.go`, on the real buckets: a public file is written, read, fetched by anyone over plain
HTTP at its public address, and deleted; a private file is written and read with the keys, an anonymous request is
refused, and it is deleted.

<!-- foot:start -->
---
[Previous: 11 Console commands and the scheduler](11-console-commands-and-scheduler.md) | [Index](00-index.md) | [Next: 13 Database backups](13-database-backups.md)
<!-- foot:end -->
