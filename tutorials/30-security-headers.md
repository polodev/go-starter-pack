<!-- nav:start -->
**Tutorial 30 of 32** | [Index](00-index.md) | [Previous: 29 Bilingual pages](29-bilingual-pages.md) | [Next: 31 Queue monitor](31-queue-monitor.md)

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
**30. [Security headers](30-security-headers.md)**  
31. [Queue monitor](31-queue-monitor.md)  
32. [Contact form](32-contact-form.md)  

</details>
<!-- nav:end -->

# 30. Security headers

Every answer the app sends (pages, API answers, static files, 404s) carries a set of browser security headers.
In Laravel you would add a package or set them in Nginx; here one global middleware does it:
`app/http/middleware/security_headers.go`, registered in `bootstrap/app.go`.

## What is sent

| Header | Value | What it stops |
|---|---|---|
| `Content-Security-Policy` | scripts, styles and fonts from this site only; pictures also from https; no framing | injected scripts (XSS) running in the page |
| `X-Frame-Options` | `DENY` | clickjacking (your page inside someone else's frame) |
| `X-Content-Type-Options` | `nosniff` | a browser treating an upload as a script |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | full URLs (with tokens) leaking to other sites |
| `Permissions-Policy` | camera, microphone, location and payment off | a page asking for devices it never needs |
| `Cross-Origin-Opener-Policy` | `same-origin` | other windows reaching into yours |
| `Strict-Transport-Security` | one year, sub-domains, **production only** | a first request over plain http |

## Global middleware, Laravel style

Laravel 11 registers global middleware in `bootstrap/app.go`'s `->withMiddleware(...)`. Goravel has the same:

```go
WithMiddleware(func(handler contractsconfiguration.Middleware) {
	handler.Append(&middleware.SecurityHeaders{})
}).
```

Global middleware runs before routing, so it also covers `/public/...` files and pages that do not exist.

## No inline scripts

The policy says `script-src 'self'`: a `<script>` block or an `onclick="..."` in a template is ignored by the
browser. That is the point (an injected script is ignored too), but it means:

- JavaScript goes in a file under `public/js/` and is loaded with `<script src="/public/js/....js"></script>`.
- Buttons get a `data-*` attribute and the file listens for clicks. The theme switch is the example:
  `<button data-theme-toggle>` in `site::header`, handled in `public/js/theme.js`.

`tests/feature/security_headers_test.go` opens the public pages and the admin screens and fails when one has an
inline script or event handler, so you hear about it in the test run, not from a broken button.

## Settings (.env)

```dotenv
# The whole policy; empty sends none. Default: config.DefaultContentSecurityPolicy.
# SECURITY_CONTENT_SECURITY_POLICY=
# HSTS: on when APP_ENV=production. Only turn it on when the site is served over https.
# SECURITY_STRICT_TRANSPORT_SECURITY=true
# SECURITY_STRICT_TRANSPORT_SECURITY_MAX_AGE=31536000
```

To load something from another site (say a script from a CDN), add its address to the policy in `.env` rather
than turning the policy off:

```dotenv
SECURITY_CONTENT_SECURITY_POLICY="default-src 'self'; script-src 'self' https://cdn.example.com; ..."
```

A single handler that needs something different sets its own header; the middleware runs first, so the handler's
value wins:

```go
ctx.Response().Header("X-Frame-Options", "SAMEORIGIN")
```

## Check it

```sh
go test ./tests/feature -run TestSecurityHeadersTestSuite -count=1
curl -sI http://127.0.0.1:3200/ | grep -i -E "content-security|x-frame|nosniff|referrer"
```

<!-- foot:start -->
---
[Previous: 29 Bilingual pages](29-bilingual-pages.md) | [Index](00-index.md) | [Next: 31 Queue monitor](31-queue-monitor.md)
<!-- foot:end -->
