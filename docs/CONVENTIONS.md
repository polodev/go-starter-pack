# Conventions of this starter pack

How code in this repository is organised and written. When a rule and the code disagree, fix the code or update this
file in the same change. The vocabulary is Laravel's, because Goravel keeps Laravel's names.

## The workflow for any change

1. Generate, then edit: `go run . artisan make:controller|model|migration|job|command|middleware|test ...`. The
   generators also register commands, jobs and migrations in `bootstrap/`.
2. Use what Goravel already has (ORM, queue, scheduler, storage, cache, mail, validation, hashing) before adding a
   module. A new dependency needs a reason you can state in one sentence.
3. Write the test with the feature.
4. Before you call it done: `gofmt -l .` (prints nothing), `go vet ./...`, `go test ./... -count=1`.

## Names: long, full and specific

Name things by exactly what they are: `backup:run`, `r2_backups`, `AuthController.Register`, not `run`, `disk2`,
`Handle`. Commands are `<area>:<action>`. Tables are Laravel's plural `snake_case`.

## Layout

| Folder | Holds |
|---|---|
| `app/http/controllers` | slim handlers: parse, validate, call a service or model, answer |
| `app/http/middleware` | middleware types (`Handle` + `Signature`) |
| `app/http/responses` | Laravel-shaped JSON answers |
| `app/models` | ORM models (`orm.Model` + fields) |
| `app/services/<area>` | domain logic (`backup` ...) |
| `app/console/commands` | artisan commands |
| `app/jobs` | queue jobs |
| `app/storage` | filesystem drivers (`R2`) |
| `config/` | one file per concern, values from `.env` |
| `database/migrations` | one migration per table, schema builder |
| `routes/` | `web.go` (pages), `api.go` (`/api/v1`) |
| `tests/feature` | testify suites on `tests.TestCase` |

## JSON API

- Everything under `/api/v1`, JSON only.
- Shapes follow Laravel: `{ "data": ... }`, `{ "user": ..., "token": ... }` for sign-in, 401
  `{ "message": "Unauthenticated." }`, 422 `{ "message": "<first error> (and N more errors)", "errors": { "field": [...] } }`.
- Protect a route with `router.Middleware(&middleware.Authenticate{})` (bearer token); read the person with
  `facades.Auth(ctx).User(&user)`.
- Sign-in treats an unknown email and a wrong password alike. Never return password hashes or tokens' secrets.

## Database

- MariaDB/MySQL by default; PostgreSQL works by `DB_CONNECTION=postgres` (driver installed).
- Laravel-compatible tables: `users`, `jobs`, `failed_jobs` keep Laravel's columns. Passwords are bcrypt (Laravel's
  default), so imported Laravel users can sign in.
- One migration file per table. Change a table you have shipped with a new migration, never by editing an old one.
- **Tests use their own database.** Inside `go test`, `config/database.go` uses `DB_TEST_DATABASE` (default
  `<DB_DATABASE>_test`) and panics unless the name ends in `_test`; `tests/test_case.go` checks again. Test suites that
  touch the database call `s.RefreshDatabase()` in `SetupTest`.

## Files and Cloudflare R2

- Disks in `config/filesystems.go`: `local`, `public`, `r2_public` (uploads anyone may read, served from
  `R2_PUBLIC_URL`) and `r2_backups` (private bucket, own keys). Never put private files on `r2_public`.
- Both R2 disks use `app/storage.R2`: the official MinIO driver plus a correct `Url()` (R2 public addresses have no
  bucket in the path). `public_url` is the public address; the driver's required `url` holds the private S3 address.
- Paths start with `R2_ROOT` (the app's folder) so several apps can share a bucket.

## Queue

- Laravel's `database` driver (`QUEUE_CONNECTION=database`). Dispatch:
  `facades.Queue().Job(&jobs.X{}, args).OnQueue("emails").Delay(t).Dispatch()`.
- `go run .` runs a worker inside the web process; production can add `queue:work --queue=... --tries=3` programs
  under Supervisor (`deploy/supervisor.conf`). Each job is taken by one worker.
- Failed jobs land in `failed_jobs`: `queue:failed`, `queue:retry all`.
- Note: Goravel's `jobs` table uses timestamps for `reserved_at`/`available_at`, where Laravel uses unix integers.
  The table name and columns match, but a Laravel worker and a Goravel worker must not share one `jobs` table.

## Backups and the scheduler

- `backup:run` dumps the database (`mysqldump`, or `pg_dump` for PostgreSQL), gzips it and stores
  `<BACKUP_PREFIX>/<BACKUP_NAME>/<time>.sql.gz` on `r2_backups`. `backup:list` lists them. The password goes through
  the environment, never the command line. The dump is held in memory while it uploads (the MinIO driver reads whole
  files); stream it for very large databases.
- Schedules live in `bootstrap/schedule.go` (Laravel's `routes/console.php`); the web process runs them. `backup:run`
  runs daily at 01:30 when backup keys are set. Keep one web process, or keep `OnOneServer()`.

## Tests

- Feature tests in `tests/feature/<feature>_test.go`, testify suites with `tests.TestCase`.
- Cover the happy path, validation errors, refusals (401, 422), and "nothing was stored after a refusal".
- Tests that talk to R2 skip themselves without keys and delete everything they write.

## Documentation and git

- Update in the same change: `README.md`, this file, and the matching tutorial.
- Commit messages: a short imperative subject, then a bullet list of what changed and why. No attribution lines for
  tools or assistants.
- Never commit `.env`, keys or tokens; scan the diff before pushing.
