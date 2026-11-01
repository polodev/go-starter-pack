# Go starter pack

A Laravel-style starter for Go, built on [Goravel](https://www.goravel.dev) v1.18 (Laravel's structure and names in Go:
artisan, facades, migrations, ORM, queues, scheduler, mail, AI). Features are modules in `app-modules/`, like
internachi/modular. The sister project of the Rust starter pack, for people who come from Laravel.

## What is integrated

- [x] **Modules** (`app-modules/<name>`) that register their own routes, migrations, commands, jobs, schedules,
      seeders, views, translations and admin menu; `make:module` creates and registers one
- [x] **Bilingual site** (English, Bangla): Goravel's translator, each module's `lang/<locale>.json`, a language switch
- [x] **Web login** with Goravel's session guard: register, login, logout, email verification, password reset,
      profile; CSRF on every form; login throttling; Laravel's `auth`, `verified` and `guest` middleware
- [x] **Token API** under `/api/v1/auth` (JWT guard) with Laravel's JSON shapes
- [x] **Roles and permissions** (the Laravel apps' model): permissions and five ranked roles defined in code and synced,
      one role per person plus direct exceptions, wildcards, admin routes that fail closed, Goravel's Gate wired in
- [x] **Admin area**: sidebar from every module shown by permission, dashboard, users, the six-part roles panel
- [x] **Activity log** (Spatie laravel-activitylog compatible) with admin screens
- [x] **Media library** (Spatie laravel-medialibrary compatible): collections, avatars, private files, Cloudflare R2
- [x] **Bilingual pages** with per-part English fallback and cleaned staff HTML
- [x] **Mail** through Goravel's mail, queued; Amazon SES through its SMTP interface
- [x] **PDF** from HTML and CSS with headless Chrome
- [x] **AI service** on Goravel's AI SDK (Claude, OpenAI, Gemini) with AI API logs and cost estimates
- [x] **Database queue** (Laravel's `jobs`, `failed_jobs`) and `queue:work`; **backups** to a private R2 bucket
- [x] Tests on a separate `*_test` database, including real R2 round trips; conventions, tutorials, Supervisor config

## Modules

| Module | Owns |
|---|---|
| `site-layout` | public layout (header, footer, language and theme switch), home page, error page |
| `auth` | web login and registration, email verification, password reset, profile, token API |
| `role-permission` | roles, permissions, assignments, fail-closed admin routes, `role-permission:sync` |
| `admin-layout` | admin shell, dashboard, user screens, roles and permissions panel |
| `activity-log` | Spatie `activity_log` screens and cleaning (writer: `app/activity`) |
| `media` | Spatie `media` table, collections, file serving, avatar page, admin media screen |
| `pdf` | HTML to PDF with headless Chrome, example invoice |
| `ai-service` | `aiservice.Chat` / `JSON`, AI API logs |
| `pages` | bilingual website pages |

App-wide pieces outside the modules: `app/modular` (the module system), `app/web` (web middleware group, page data,
templates), `app/activity` (activity writer), `app/storage` (R2 driver), `app/services/backup`, `app/support`.

## Requirements

Go 1.25 or newer, MariaDB/MySQL, `mysqldump` for backups, Chrome for PDFs, and locally a mail catcher such as
Mailpit. R2 keys and AI keys are optional (the tests that need them skip themselves).

## Run it

```sh
cp .env.example .env
go run . artisan key:generate
go run . artisan jwt:secret
mysql -uroot -e "CREATE DATABASE go_starter CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
mysql -uroot -e "CREATE DATABASE go_starter_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
go run . artisan migrate
go run . artisan role-permission:sync
go run . artisan db:seed --force       # About, Privacy policy, Terms in both languages
go run .                               # http://127.0.0.1:3200, plus a queue worker and the scheduler
```

Make yourself the first developer after registering on the site:

```sh
go run . artisan role-permission:assign --email=you@example.com --role=developer
```

Live reload: install [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`) and run
`air`. After adding Tailwind classes, rebuild the CSS:
`./tools/tailwindcss -c tailwind.config.js -i resources/css/input.css -o public/css/app.css --minify`
(Tailwind standalone CLI v3.4 in `tools/`).

## Laravel to Go, at a glance

| Laravel | Here |
|---|---|
| `php artisan ...` | `go run . artisan ...` (or `./artisan ...`) |
| `php artisan make:controller/model/migration/job/command` | `go run . artisan make:...` |
| `php artisan module:make` (internachi/modular) | `go run . artisan make:module <name>` |
| `routes/web.php`, `routes/api.php` | each module's `Routes()`; `web.Group(...)` is the `web` middleware group |
| `auth`, `verified`, `guest` middleware | `&auth.Authenticate{}`, `&auth.Verified{}`, `&auth.Guest{}` |
| `@can('users.view')` | `{{ if call .can "users.view" }}` |
| `__('auth.login')` | `{{ t .lang "auth.title-login" }}` in templates, `web.T(ctx, ...)` in Go |
| `Storage::disk('s3')` | `facades.Storage().Disk("r2_public")` |
| `Mail::to(...)->queue(...)` | `facades.Mail().To(...).Content(...).Queue()` |
| `php artisan queue:work --queue=emails` | `go run . artisan queue:work --queue=emails` |
| `php artisan test` | `go test ./...` |

## Tests

```sh
gofmt -l .
go vet ./...
go test ./... -count=1
```

Tests use `DB_TEST_DATABASE` (default `<DB_DATABASE>_test`) and refuse any other database. Web tests drive pages like
a browser (session cookie, CSRF token). The R2 tests write to the real buckets and delete what they write. AI tests
use provider `fake`, so no paid API is called.

## Docs

- `AGENTS.md` and `CLAUDE.md`: how to work in this repository (also for AI assistants)
- `docs/CONVENTIONS.md`: the rules
- `tutorials/`: numbered walkthroughs, start at `tutorials/00-index.md`
