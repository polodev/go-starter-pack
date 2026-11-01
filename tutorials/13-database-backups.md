<!-- nav:start -->
**Tutorial 13 of 31** | [Index](00-index.md) | [Previous: 12 Files on Cloudflare R2](12-files-on-r2.md) | [Next: 14 Testing](14-testing.md)

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
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  

</details>
<!-- nav:end -->

# 13. Database backups

```sh
go run . artisan backup:run     # mysqldump (or pg_dump) -> gzip -> r2_backups
go run . artisan backup:list    # db-backups/go-starter-pack/2026-11-01-01-30-00.sql.gz ...
go run . artisan backup:clean   # delete old backups by the retention settings
go run . artisan backup:monitor # is the newest backup recent? non-zero exit and a mail when not
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

**Schedule:** when backup keys are set, `bootstrap/schedule.go` runs `backup:run` at 01:30, `backup:clean` at 02:00
(after the backup, so the size limit counts it) and `backup:monitor` at 03:00, each `OnOneServer()`.

## Retention: backup:clean

spatie/laravel-backup's default strategy, with its setting names (`config/backup.go`, `backup.cleanup`):

| Setting (.env) | Default | Meaning |
|---|---|---|
| `BACKUP_KEEP_ALL_BACKUPS_FOR_DAYS` | 7 | every backup of the last 7 days stays |
| `BACKUP_KEEP_DAILY_BACKUPS_FOR_DAYS` | 16 | then the newest of each day, for 16 more days |
| `BACKUP_KEEP_WEEKLY_BACKUPS_FOR_WEEKS` | 8 | then the newest of each week, for 8 weeks |
| `BACKUP_KEEP_MONTHLY_BACKUPS_FOR_MONTHS` | 4 | then the newest of each month, for 4 months |
| `BACKUP_KEEP_YEARLY_BACKUPS_FOR_YEARS` | 2 | then the newest of each year, for 2 years; older ones go |
| `BACKUP_DELETE_OLDEST_WHEN_USING_MORE_MEGABYTES_THAN` | 5000 | then the oldest go until the rest fit (0: no limit) |

A year of nightly backups leaves about 36. **The newest backup is never deleted**, whatever the settings. Files in
the folder whose names are not a backup time are left alone.

The decision is a pure function, `backup.Plan(backups, now, retention)`, so it is unit-tested without a bucket
(`app/services/backup/retention_test.go`); `backup.Clean(now)` applies it to the disk.

## Monitoring: backup:monitor

Spatie's `backup:monitor`. Unhealthy when there is no backup, the newest is older than `BACKUP_MAXIMUM_AGE_IN_DAYS`
(1), or all of them use more than `BACKUP_MAXIMUM_STORAGE_IN_MEGABYTES` (5000). Then it:

- writes the problem to the log,
- mails `BACKUP_NOTIFICATION_EMAIL` (sent at once, not through the queue: a stuck queue must not hide a stuck backup),
- exits non-zero, so a cron wrapper or uptime checker notices too.

A failed `backup:run` or `backup:clean` sends the same kind of notice. Without `BACKUP_NOTIFICATION_EMAIL` the log is
the only place it shows up, so set it in production.

**Restore:** download the file from the private bucket, then

```sh
gunzip -c 2026-11-01-01-30-00.sql.gz | mysql -uroot go_starter
```

**Test:** `tests/feature/backup_test.go` makes a real backup of the test database on the real bucket, finds it with
`backup:list`, unzips it, checks the tables are in it, and deletes it.
`tests/feature/backup_maintenance_test.go` runs `backup:clean` and `backup:monitor` on the local disk with fake
backups of different ages: what stays, what goes, and the notices.

<!-- foot:start -->
---
[Previous: 12 Files on Cloudflare R2](12-files-on-r2.md) | [Index](00-index.md) | [Next: 14 Testing](14-testing.md)
<!-- foot:end -->
