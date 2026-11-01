<!-- nav:start -->
**Tutorial 21 of 29** | [Index](00-index.md) | [Previous: 20 Layouts, pages and translations](20-layouts-and-translations.md) | [Next: 22 Roles and permissions](22-roles-and-permissions.md)

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
**21. [Web login, verification and password reset](21-web-login.md)**  
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

# 21. Web login, verification and password reset

| Page | What happens |
|---|---|
| `/register` | creates the account (bcrypt password), signs in, emails a verification link, shows "check your inbox" |
| `/email/verify/{token}` | one-time link (24 h): the email is verified |
| `/login` | email and password; 5 tries a minute; the same message for an unknown email and a wrong password |
| `/forgot-password` | emails a reset link (60 min), with the same answer whether or not the address has an account |
| `/reset-password/{token}` | a new password; the reader proved they own the address, so it is verified too |
| `/logout` | ends and renews the session |
| `/account/profile` | name, email, verified or not |

## Laravel's middleware

| Laravel | Here |
|---|---|
| `auth` | `&auth.Authenticate{}`: a guest goes to `/login?next=...` |
| `verified` | `&auth.Verified{}`: an unverified email goes to `/email/verify` |
| `guest` | `&auth.Guest{}`: a signed-in person skips login and register |
| `auth()->user()` | `auth.CurrentUser(ctx)` |

## Under the hood

- Goravel's session guard: `facades.Auth(ctx).Guard("web").Login(&user)` renews the session id (no session fixation).
- CSRF: every form sends `_token`; Goravel's VerifyCsrfToken answers 419 without it.
- Throttling: Goravel's rate limiter (`facades.RateLimiter().For("login", ...)` and `middleware.Throttle("login")`).
- Email links: a random secret in the URL, only its SHA-256 in `user_security_tokens`, single use, expiring;
  `auth:prune-tokens` cleans them daily.
- Mail: Goravel's `facades.Mail().Queue()`; in tests `auth.SendMail` is replaced to capture the mail.

<!-- foot:start -->
---
[Previous: 20 Layouts, pages and translations](20-layouts-and-translations.md) | [Index](00-index.md) | [Next: 22 Roles and permissions](22-roles-and-permissions.md)
<!-- foot:end -->
