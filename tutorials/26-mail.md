<!-- nav:start -->
**Tutorial 26 of 32** | [Index](00-index.md) | [Previous: 25 Media library](25-media-library.md) | [Next: 27 PDF](27-pdf.md)

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
**26. [Mail](26-mail.md)**  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  
32. [Contact form](32-contact-form.md)  

</details>
<!-- nav:end -->

# 26. Mail

Mail is Goravel's own, so there is no mail module to learn:

```go
err := facades.Mail().To([]string{"customer@example.com"}).Subject("Your invoice").
	Content(mail.Content{Html: html, Text: text}).
	Attach([]string{"storage/app/invoice.pdf"}).
	Queue() // Mail::to(...)->queue(...): a job on the database queue; Send() sends at once
```

Build the HTML from a template with `web.Render("shop::mail-order", data)` (pass `"lang"` for translations), or use
Goravel's mailables (`go run . artisan make:mail OrderShipped`).

## Settings

| Where | `.env` |
|---|---|
| Your Mac | a catcher such as Mailpit (`brew install mailpit`, inbox on http://localhost:8025): `MAIL_HOST=127.0.0.1`, `MAIL_PORT=1025` |
| Any SMTP server | `MAIL_HOST`, `MAIL_PORT`, `MAIL_USERNAME`, `MAIL_PASSWORD` |
| Amazon SES | `MAIL_MAILER=ses`, `AWS_SES_ACCESS_KEY_ID`, `AWS_SES_SECRET_ACCESS_KEY`, `AWS_SES_DEFAULT_REGION` |

With `ses`, `config/mail.go` connects to `email-smtp.<region>.amazonaws.com:587` and derives the SMTP password from the
IAM secret key (AWS's algorithm, `app/support/ses`, tested against AWS's examples). The sender address must be verified
in SES; in the SES sandbox only verified addresses receive mail.

A queued mail is sent by the worker (`go run .` runs one; production: `queue:work`). If sending fails it lands in
`failed_jobs`: `go run . artisan queue:failed`, then `queue:retry all` once the mail server works.

<!-- foot:start -->
---
[Previous: 25 Media library](25-media-library.md) | [Index](00-index.md) | [Next: 27 PDF](27-pdf.md)
<!-- foot:end -->
