<!-- nav:start -->
**Tutorial 13 of 29** | [Index](00-index.md) | [Previous: 12 Files on Cloudflare R2](12-files-on-r2.md) | [Next: 14 Testing](14-testing.md)

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
**13. [Database backups](13-database-backups.md)**  
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

</details>
<!-- nav:end -->

# 13. Database backups

```sh
go run . artisan backup:run     # mysqldump (or pg_dump) -> gzip -> r2_backups
go run . artisan backup:list    # db-backups/go-starter-pack/2026-11-01-01-30-00.sql.gz ...
```

Each backup is `<BACKUP_PREFIX>/<BACKUP_NAME>/<YYYY-MM-DD-HH-MM-SS>.sql.gz` on the private bucket.

**How it works** (`app/services/backup/backup.go`):

- MySQL/MariaDB: `mysqldump --single-transaction --quick --skip-lock-tables --routines --triggers` in utf8mb4. A MySQL 8
  client gets `--column-statistics=0` automatically so it can dump a MariaDB server.
- PostgreSQL: `pg_dump --no-owner --no-password`.
- The password goes through the environment (`MYSQL_PWD`, `PGPASSWORD`), never the command line.
- A failing tool or an empty dump stores nothing.
- **No memory used for the dump**, whatever the database size:

  ```
  mysqldump ──pipe──▶ gzip ──▶ storage/app/backup-temp/<time>.sql.gz ──▶ R2 multipart upload, 16 MB parts
  ```

  The tool's output is copied 32 KB at a time into gzip and a temporary file in `storage/app/backup-temp` (Spatie's
  backup-temp; `BACKUP_TEMPORARY_DIRECTORY` to move it). The file goes to R2 as a multipart upload, one part at a time,
  read straight from disk. A failed dump is never uploaded, the temporary file is always removed, and leftovers of a
  crashed run are cleaned at the next run. You need free disk space for one compressed dump (often 10 to 20% of the
  database). The test backs up a 48 MB dump while the program's memory grows by under 2 MB.

**Schedule:** `bootstrap/schedule.go` runs `backup:run` daily at 01:30 when backup keys are set.

**Restore:** download the file from the private bucket, then

```sh
gunzip -c 2026-11-01-01-30-00.sql.gz | mysql -uroot go_starter
```

**Test:** `tests/feature/backup_test.go` makes a real backup of the test database on the real bucket, finds it with
`backup:list`, unzips it, checks the tables are in it, and deletes it.

Not yet: retention (`backup:clean`) and a freshness monitor, which the Rust starter has.

<!-- foot:start -->
---
[Previous: 12 Files on Cloudflare R2](12-files-on-r2.md) | [Index](00-index.md) | [Next: 14 Testing](14-testing.md)
<!-- foot:end -->
