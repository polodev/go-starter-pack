<!-- nav:start -->
**Tutorial 11 of 33** | [Index](00-index.md) | [Previous: 10 Queues and jobs](10-queues-and-jobs.md) | [Next: 12 Files on Cloudflare R2](12-files-on-r2.md)

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
**11. [Console commands and the scheduler](11-console-commands-and-scheduler.md)**  
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
32. [Contact form](32-contact-form.md)  
33. [Sign in with Google](33-sign-in-with-google.md)  

</details>
<!-- nav:end -->

# 11. Console commands and the scheduler

## Make a command

```sh
go run . artisan make:command SendReport     # app/console/commands/send_report.go, registered in bootstrap/commands.go
```

```go
func (r *SendReport) Signature() string   { return "report:send" }   // <area>:<action>
func (r *SendReport) Description() string { return "Email the weekly report" }

func (r *SendReport) Extend() command.Extend {
	return command.Extend{Category: "report", Flags: []command.Flag{
		&command.StringFlag{Name: "to", Usage: "address"},
		&command.BoolFlag{Name: "dry-run"},
	}}
}

func (r *SendReport) Handle(ctx console.Context) error {
	to := ctx.Option("to")                // --to=boss@example.com
	if ctx.OptionBool("dry-run") { ctx.Info("would send to " + to); return nil }
	ctx.Success("sent")
	return nil                           // an error makes the command exit non-zero
}
```

Keep the work in a service (`app/services/...`) and the command thin, as `backup:run` does.

## Schedule it

`bootstrap/schedule.go` is Laravel's `routes/console.php`:

```go
facades.Schedule().Command("report:send --to=boss@example.com").DailyAt("07:00").OnOneServer()
```

```sh
go run . artisan schedule:list
go run . artisan schedule:run     # run what is due now (for testing)
```

The web process (`go run .` or the binary) runs the scheduler. Run one web process, or keep `OnOneServer()`, so a task
does not run twice.

<!-- foot:start -->
---
[Previous: 10 Queues and jobs](10-queues-and-jobs.md) | [Index](00-index.md) | [Next: 12 Files on Cloudflare R2](12-files-on-r2.md)
<!-- foot:end -->
