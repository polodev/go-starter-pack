<!-- nav:start -->
**Tutorial 31 of 33** | [Index](00-index.md) | [Previous: 30 Security headers](30-security-headers.md) | [Next: 32 Contact form](32-contact-form.md)

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
24. [Activity log](24-activity-log.md)  
25. [Media library](25-media-library.md)  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  
**31. [Queue monitor](31-queue-monitor.md)**  
32. [Contact form](32-contact-form.md)  
33. [Sign in with Google](33-sign-in-with-google.md)  

</details>
<!-- nav:end -->

# 31. Queue monitor

Laravel has Horizon (for Redis) and jobs-monitor packages for the database queue. This starter has the
`queue-watcher` module: open **`/admin/queue-watcher`** (sidebar: System › Queue monitor).

| Screen | Shows | Actions (`queue.manage`) |
|---|---|---|
| Overview | waiting, delayed, running and failed counts; jobs per queue; newest runs | |
| Job history | every run: status, start, duration, error; filter by status and job name | delete a run, mark unfinished runs stale, purge old ones |
| Pending jobs | the `jobs` table: waiting, delayed, running | delete a waiting job |
| Failed jobs | `failed_jobs`, with the whole error and the arguments | retry one or all, delete one or all |

Employees can look (`queue.view`); admins can act (`queue.*`). Every action is in the activity log (log `queue`).

## How every run is recorded without touching your jobs

Goravel's worker does not call your job struct directly: it reads the signature from the payload and asks the
job storer for the registered job with that signature. So at start-up the module re-registers every job wrapped in a
tracker (`app-modules/queue-watcher/tracker.go`):

```go
func (t *trackedJob) Handle(args ...any) error {
	// insert a "running" row in queue_jobs
	err := t.inner.Handle(args...)
	// mark it "finished", or "failed" with err
	return err
}
```

It runs from the module's `Boot()` (the new `modular.HasBoot`, Laravel's `ServiceProvider::boot`), which bootstrap
calls after every provider and module has registered its jobs, so Goravel's own mail job is tracked too.

- Each try is its own run: a job with `--tries=3` that fails three times shows three failed runs, then one row in
  failed jobs.
- A run that never finishes (the worker was killed) stays "running" until `queue-watcher:purge` marks it **stale**
  after `QUEUE_WATCHER_STALE_AFTER_MINUTES` (60).
- `sync` jobs are not recorded: they never go through the worker.

## Retry

The Retry button is `go run . artisan queue:retry <uuid>`: Goravel's `Failer` pushes the job back on its queue and
removes it from `failed_jobs`. A worker then picks it up again.

## Secrets in arguments

Job arguments are stored in the `jobs` and `failed_jobs` payloads, and the screens show them, except for jobs in
`queuewatcher.HiddenArgumentJobs`. Goravel's mail job is there, because a queued mail can hold a password reset or
sign-in link. When your own job carries something secret, add its signature:

```go
queuewatcher.HiddenArgumentJobs["send_invoice"] = true
```

Better still, pass an id (`invoice 42`) and load the data inside the job.

## Housekeeping

`queue-watcher:purge` runs daily at 02:30: it marks stuck runs stale, then deletes history older than
`QUEUE_WATCHER_RETENTION_DAYS` (30). It never touches waiting or failed jobs.

```sh
go run . artisan queue-watcher:purge --days=7
```

## What the tests prove

`tests/feature/queue_watcher_test.go`: a real worker's runs are recorded as finished or failed with the error; the mail
job is wrapped once; customers and new employees get 403; employees see but cannot retry; an admin retries one job (it
leaves `failed_jobs` and waits on its queue again), retries all, deletes, flushes; a missing CSRF token changes
nothing; a failed job without a uuid is not retried; a mail's reset link never appears on the page; a running job
cannot be deleted; stuck runs become stale and old history is purged.

<!-- foot:start -->
---
[Previous: 30 Security headers](30-security-headers.md) | [Index](00-index.md) | [Next: 32 Contact form](32-contact-form.md)
<!-- foot:end -->
