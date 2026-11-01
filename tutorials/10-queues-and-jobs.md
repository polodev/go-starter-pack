<!-- nav:start -->
**Tutorial 10 of 29** | [Index](00-index.md) | [Previous: 9 Token API: register, login, me](09-auth-api.md) | [Next: 11 Console commands and the scheduler](11-console-commands-and-scheduler.md)

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
**10. [Queues and jobs](10-queues-and-jobs.md)**  
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

</details>
<!-- nav:end -->

# 10. Queues and jobs

Jobs wait in the `jobs` table and failures land in `failed_jobs`, like `QUEUE_CONNECTION=database` in Laravel.

## Make a job

```sh
go run . artisan make:job SendDigest      # app/jobs/send_digest.go, registered in bootstrap/jobs.go
```

```go
func (r *SendDigest) Signature() string { return "send_digest" }

func (r *SendDigest) Handle(args ...any) error {
	userID := args[0].(int)                // arguments arrive in the order they were dispatched
	// ... the work; return an error to make this try fail
	return nil
}
```

## Dispatch it

```go
// Laravel: SendDigest::dispatch(7)->onQueue('emails')->delay(now()->addMinute())
err := facades.Queue().Job(&jobs.SendDigest{}, []queue.Arg{{Type: "int", Value: 7}}).
	OnQueue("emails").Delay(time.Now().Add(time.Minute)).Dispatch()
```

`DispatchSync()` runs it at once, without a worker.

## Workers

| Laravel | Here |
|---|---|
| `php artisan queue:work --queue=emails --tries=3` | `go run . artisan queue:work --queue=emails --tries=3` |
| `php artisan queue:failed` | `go run . artisan queue:failed` |
| `php artisan queue:retry all` | `go run . artisan queue:retry all` |
| Supervisor program | `[program:app-queue-worker]` in `deploy/supervisor.conf` |

`go run .` already runs one worker for the `default` queue inside the web process. Add `queue:work` programs for
other or busy queues; each job is taken by exactly one worker.

## What the tests prove

`tests/feature/queue_test.go`, on the real test database: a dispatched job waits in `jobs` until a worker runs it with
its argument; a worker leaves other queues and delayed jobs alone; a failing job is tried twice and then moves to
`failed_jobs` with its error.

**Note:** Goravel stores `reserved_at`/`available_at` as timestamps where Laravel uses unix integers. Do not let a
Laravel worker and a Goravel worker share one `jobs` table.

<!-- foot:start -->
---
[Previous: 9 Token API: register, login, me](09-auth-api.md) | [Index](00-index.md) | [Next: 11 Console commands and the scheduler](11-console-commands-and-scheduler.md)
<!-- foot:end -->
