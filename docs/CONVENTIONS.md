# Conventions of this starter pack

How code in this repository is organised and written. When a rule and the code disagree, fix the code or update this
file in the same change. The vocabulary is Laravel's, because Goravel keeps Laravel's names.

## The workflow for any change

1. Generate, then edit: `go run . artisan make:module|controller|model|migration|job|command|middleware|test ...`.
   The generators register what they make (`bootstrap/` or `bootstrap/modules.go`).
2. Use Goravel's batteries first: ORM, queue, scheduler, storage, cache, mail, translation, sessions, CSRF, rate
   limiter, Gate, validation, hashing, AI SDK. Build only what Goravel lacks, and say in the code why.
   A new dependency needs a reason you can state in one sentence (`chromedp`: PDFs; `bluemonday`: cleaning HTML).
3. Write the test with the feature.
4. Before you call it done: `gofmt -l .` (prints nothing), `go vet ./...`, `go test ./... -count=1`,
   and `go run . artisan role-permission:audit` when you touched admin routes.

## Names: long, full and specific

Name things by exactly what they are: `backup:run`, `r2_backups`, `ai_api_logs`, `activity_logs.view`, never `run`,
`disk2`, `logs`. Commands are `<module>:<action>` (`role-permission:sync`, `ai-service:purge-ai-api-logs`). Tables are
Laravel's plural `snake_case`. The history of AI calls is "AI API logs", never "API logs".

## Modules

```text
app-modules/<name>/
  module.go               Module: Name() plus the interfaces it implements (see app/modular)
  README.md               what it does, its routes and settings
  migrations/ models/     its tables (one migration file per table group) and models
  resources/views/        templates, each named "<module>::<view>" with {{ define "auth::login" }}
  lang/en.json, bn.json   its translations, read as "<module>.<key>"
```

- `go run . artisan make:module <kebab-name>` creates one and registers it in `bootstrap/modules.go`.
- A module offers routes (`HasRoutes`), migrations, commands, jobs, schedules, seeders, views and admin menu items.
- Dependency direction, no cycles: `site-layout`, `auth` < `role-permission` < `admin-layout` < modules with admin
  screens (`activity-log`, `media`, `pages`, `ai-service`). Shared writers that low modules need live in `app/`
  (`app/activity`), never in a module.
- App-wide code: `app/modular` (module system), `app/web` (web group, page data, templates, translation helpers),
  `app/activity`, `app/storage`, `app/services/backup`, `app/support`.

## Routes and URLs

| Kind | How | Example |
|---|---|---|
| Public page | `web.Group(func(router route.Router) {...})` (session, CSRF, language) | `/pages/about` |
| Signed-in area | `router.Middleware(&auth.Authenticate{})` inside the web group | `/account/profile` |
| Admin | `rolepermission.Admin(func(admin *rolepermission.AdminRouter) {...})` | `/admin/users/{id}/edit` |
| JSON API | `facades.Route().Prefix("/api/v1")`, bearer token | `/api/v1/auth/me` |
| Files | media module | `/media/{id}/{file_name}` |

Admin resources: `GET /admin/things` (list), `GET /admin/things/{id}` (show), `GET /admin/things/{id}/edit`,
`POST /admin/things/{id}` (save), `POST /admin/things/{id}/<action>`. HTML forms send GET and POST only.

## Permissions and roles

- Defined in code in `app-modules/role-permission/definitions.go`: permissions by area (`<area>.view`,
  `<area>.manage`, other actions), five ranked roles, each including the one below. Copy them into the tables with
  `go run . artisan role-permission:sync` (or the Sync button).
- One role per person plus optional direct exceptions; `roles.*` never as an exception; the last developer stays.
  A developer holds `*`; `area.*` covers an area; any grant opens the dashboard; `ROLE_PERMISSION_SUPER_USER_IDS`
  pass every check.
- Every admin route is declared with its permission: `admin.Get("/things", "things.view", Index)`. An undeclared
  admin route answers 403 to everyone; `role-permission:audit` lists them.
- In Go: `rolepermission.HasPermission(user.ID, "things.manage")`; in templates: `{{ if call .can "things.manage" }}`;
  Goravel's `facades.Gate().WithContext(ctx).Allows("things.view", nil)` answers the same.

## Admin screens

- Render with `adminlayout.View(ctx, "module::view", title, crumbs, values)`; templates use
  `{{ template "admin::header" . }}` and `{{ template "admin::footer" . }}`.
- Breadcrumb always starts at Dashboard. Lists: ID first, as a link (`data-testid="id-link"`), Actions on the right with
  Show and, if permitted, Edit. Show is read-only; Edit is a separate page.
- Feedback through `?saved=` / `?error=` (one banner, `data-testid="status"`); never free text in the URL.
- Sidebar entries come from the module's `AdminMenu()`; a link shows only to people who hold its permission.
- Every change writes an activity entry: `activity.In("users").CausedBy(id).PerformedOn("user", id).Event("updated")
  .Changes(before, after).LogQuietly("updated user")`. Never put secrets in an entry.

## Public pages, templates and translations

- Pages render with `web.View(ctx, "module::view", values)` (or `web.ViewStatus` for 404/422), wrapped in
  `{{ template "site::header" . }}` / `{{ template "site::footer" . }}`. `web.Data` adds `lang`, `csrf`, `user`, `can`.
- The site speaks English and Bangla. All visible text is a translation key in both `lang/en.json` and `lang/bn.json`
  of the module: `{{ t .lang "auth.title-login" }}`, `{{ tr .lang "auth.verify-body" "email" .email }}`,
  `web.T(ctx, "auth.status-verified")`. Admin screens are English only.
- Database content (pages): English required, other languages optional, each empty part NULL and shown in English.
- Tailwind with the brand colours; rebuild `public/css/app.css` after adding classes; dark mode for every colour.
- `template.HTML` only for HTML cleaned with bluemonday before it was stored.

## Forms and security

- Every form carries `<input type="hidden" name="_token" value="{{ .csrf }}">`; Goravel's VerifyCsrfToken answers 419
  without it. Redirect after a successful POST (303).
- Validate on the server; on errors answer 422 with the typed values kept.
- Passwords: bcrypt through `facades.Hash()` (Laravel's default), at least 8 characters, never logged or returned.
- Login: Goravel session guard (`Guard("web")`), throttled with the rate limiter; unknown email and wrong password give
  the same answer. Email links: random secret, only its SHA-256 stored, single use, expiring.
- Middleware that stops a request answers with `Redirect(...).Abort()` or an `AbortableResponse` (a bare
  `Request().Abort()` answers 400).
- Private files and other sensitive things: unknown and forbidden both answer 404.

## JSON API

- Under `/api/v1`, JSON only, guard `api` (JWT): `router.Middleware(&auth.ApiAuthenticate{})`.
- Laravel's shapes: `{"data": ...}`, `{"user": ..., "token": ...}`, 401 `{"message": "Unauthenticated."}`,
  422 `{"message": "<first error> (and N more errors)", "errors": {...}}` (`app/http/responses`).
- The default guard is `api`, because `facades.Auth(ctx)` builds the default guard first and API routes have no session.

## Database

- MariaDB/MySQL by default; PostgreSQL by `DB_CONNECTION=postgres`. Use the schema builder and models (they quote
  columns per database); no MySQL-only SQL.
- Laravel-compatible tables keep Laravel's columns: `users`, `jobs`, `failed_jobs`, `activity_log` (Spatie), `media`
  (Spatie), `roles`/`permissions`/pivots (the Laravel apps' role-permission module).
- A shipped migration is never edited; change a table with a new migration. Every `Down` drops what `Up` made.
- Tests use `DB_TEST_DATABASE` (enforced in `config/database.go`) and call `s.RefreshDatabase()` in `SetupTest`.

## Files, media and Cloudflare R2

- Disks: `local`, `public`, `r2_public` (public uploads), `r2_private` (private uploads), `r2_backups` (backups). R2 disks
  go through `app/storage.R2` (MinIO driver plus correct public URLs).
- Media library: `media.Define(ownerType, media.Single("avatar").Images().AsPublic().Max(2<<20))`, then
  `media.Add(owner, collection, file)` / `AddBytes`, `media.URL(item)`. Files sit at `<MEDIA_PREFIX>/<id>/<file_name>`.
  Never store a path or URL in your own table; ask the media module.
- `MEDIA_STORAGE=r2` puts public collections in the public bucket and private ones in the private bucket.

## Mail, PDF and AI

- Mail: Goravel's `facades.Mail()`, queued with `.Queue()`. Locally a catcher (Mailpit); `MAIL_MAILER=ses` for SES.
  Mail text uses translation keys; never put secrets that are not needed in an email.
- PDF: `pdf.FromHTML(html)`; page size and margins in CSS `@page`. Money in minor units (`pdf.Money`), never floats.
- AI: only through `aiservice.Chat` / `aiservice.JSON` / `aiservice.Image`, always with a `Purpose`; pictures are never
  stored in the AI API logs; `Sensitive: true` for anything about
  a person. Keys and models in `.env`. A model's reply is input, not truth: check the fields.

## Queue, commands and the scheduler

- Jobs: `make:job`, dispatch with `facades.Queue().Job(...).OnQueue(...).Dispatch()`; `queue:work` under Supervisor.
- Commands: `<module>:<action>`, thin, the work in a service; a failure returns an error (non-zero exit).
- Schedules: a module's `Schedule()` (or `bootstrap/schedule.go`), `OnOneServer()` for anything that must run once.

## Tests

- Feature tests in `tests/feature/<feature>_test.go`: testify suites on `tests.TestCase`, `RefreshDatabase` in
  `SetupTest`. Pure functions get unit tests next to the code.
- Drive pages with the `browser` helper (`tests/feature/browser_test.go`): it keeps the session cookie and sends the
  CSRF token; `upload` sends files. Replace outside effects with seams (`auth.SendMail`, provider `fake`).
- Cover the happy path, validation (422), refusals (401, 403, 404, 419), and "nothing stored after a refusal".
- Assert on `data-testid` values and visible text, not on classes.

## Documentation and git

- Update in the same change: the module's `README.md`, `README.md`, this file and the matching tutorial.
- Commit messages: a short imperative subject, then a bullet list of what changed and why. No attribution lines for
  tools or assistants. Never commit `.env`, keys or tokens; scan the diff before pushing.

## Checklists

**A new module:** `make:module`, README, entry in the root README module table, no dependency cycle, migrations,
routes, views with `module::` names, translations in both languages, permissions in `definitions.go`, tests, tutorial.

**A new admin resource:** permissions `<area>.view` / `.manage` (and roles that get them), `role-permission:sync`,
routes through `rolepermission.Admin`, `AdminMenu()` entry, ID-first list, show, edit, activity entries, tests for
list, show, edit, validation, 403 and "nothing changed".

**A new public form:** `_token`, server validation with 422 and kept values, both languages, rate limit if it sends
mail or checks credentials, tests including the 419 without a token.
