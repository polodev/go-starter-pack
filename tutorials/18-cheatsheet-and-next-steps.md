<!-- nav:start -->
**Tutorial 18 of 29** | [Index](00-index.md) | [Previous: 17 Conventions](17-conventions.md) | [Next: 19 Modules (app-modules)](19-modules.md)

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
**18. [Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md)**  
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

# 18. Cheat sheet and ideas for next

## Cheat sheet

| Laravel | Go starter pack |
|---|---|
| `php artisan serve` | `go run .` / `air` |
| `php artisan make:controller X` | `go run . artisan make:controller X` |
| `php artisan make:model X -m` | `make:model X` + `make:migration create_xs_table` |
| `php artisan migrate` / `migrate:fresh` | the same names |
| `php artisan db:seed` | `go run . artisan db:seed` |
| `php artisan make:job X` / `X::dispatch()` | `make:job X` / `facades.Queue().Job(&jobs.X{}, args).Dispatch()` |
| `php artisan queue:work` | `go run . artisan queue:work --queue=default` |
| `php artisan schedule:list` | the same |
| `Storage::disk('s3')->put()` | `facades.Storage().Disk("r2_public").Put()` |
| `Hash::make()` / `Hash::check()` | `facades.Hash().Make()` / `.Check()` |
| `$request->validate([...])` | `ctx.Request().Validate(map[string]any{...})` |
| `auth()->user()` | `facades.Auth(ctx).User(&user)` |
| `config('queue.default')` | `facades.Config().GetString("queue.default")` |
| `php artisan test` | `go test ./... -count=1` |
| `composer require x` | `go get x` then `go mod tidy` |

## Where each Laravel habit lives here

| Laravel | Go starter pack |
|---|---|
| `php artisan module:make Shop` | `go run . artisan make:module shop` |
| `Route::middleware('auth')` | `router.Middleware(&auth.Authenticate{})` |
| `@can('users.view')` / `Gate::allows` | `{{ if call .can "users.view" }}` / `rolepermission.HasPermission` |
| `__('auth.login')` | `{{ t .lang "auth.title-login" }}`, `web.T(ctx, "...")` |
| `activity()->log('...')` | `activity.In("users")...LogQuietly("...")` |
| `$user->addMedia($file)->toMediaCollection('avatar')` | `media.Add(media.UserOwner(id), "avatar", file)` |
| `Mail::to($x)->queue($mail)` | `facades.Mail().To(...).Content(...).Queue()` |
| `Pdf::loadView(...)` | `pdf.FromHTML(html)` |
| `Prism::text()->...` | `aiservice.Chat(ctx, aiservice.Request{Purpose: ...})` |

## Ideas for next

- More admin resources, made with the checklist in `docs/CONVENTIONS.md`.
- A contact form (public form with a honeypot, rate limit and an admin inbox), like the Rust starter's.
- Two-factor sign-in, sign-in links by email, and an admin screen for sent mail.
- Image conversions (thumbnails) in the media library.
- Running the test suite on PostgreSQL ([16](16-mysql-or-postgresql.md)).

<!-- foot:start -->
---
[Previous: 17 Conventions](17-conventions.md) | [Index](00-index.md) | [Next: 19 Modules (app-modules)](19-modules.md)
<!-- foot:end -->
