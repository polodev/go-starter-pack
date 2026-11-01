# Go starter pack

A Laravel-style starter for Go, built on [Goravel](https://www.goravel.dev) v1.18 (Laravel's structure and names in Go:
artisan, facades, migrations, ORM, queues, scheduler). The sister project of the Rust starter pack, for people who
come from Laravel.

## What is integrated

- [x] Goravel v1.18 on **MariaDB/MySQL** (default) or PostgreSQL (`DB_CONNECTION`), with a separate test database
      and a guard that refuses to test against any database whose name does not end in `_test`
- [x] **Token API** under `/api/v1/auth`: register, login, me, with Laravel's JSON shapes (`{user, token}`,
      `{"message": "Unauthenticated."}`, 422 `{message, errors}`); bcrypt passwords, so Laravel users import as is
- [x] Laravel's `users`, `jobs` and `failed_jobs` tables
- [x] **Database queue** (Laravel's `database` driver): `dispatch`, `OnQueue`, `Delay`, retries, `failed_jobs`,
      `queue:failed`, `queue:retry`, and a `queue:work` command for Supervisor
- [x] **Cloudflare R2**: a public disk for uploads (served from `R2_PUBLIC_URL`) and a private disk for backups
- [x] **Database backups**: `backup:run` (mysqldump or pg_dump, gzip) to the private R2 bucket, `backup:list`, daily at 01:30
- [x] Tests against the real R2 buckets and the real database queue
- [x] Supervisor config (`deploy/supervisor.conf`), conventions, tutorials

Not yet (the Rust starter has them): web login pages, roles and permissions, admin screens, media library,
activity log, bilingual pages, mail queue, PDF, AI service. See `docs/CONVENTIONS.md` for how to add them.

## Requirements

Go 1.25 or newer, MariaDB/MySQL, and `mysqldump` for backups. R2 keys are optional (the R2 tests skip without them).

## Run it

```sh
cp .env.example .env
go run . artisan key:generate
go run . artisan jwt:secret
mysql -uroot -e "CREATE DATABASE go_starter CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
mysql -uroot -e "CREATE DATABASE go_starter_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
go run . artisan migrate
go run .                         # http://127.0.0.1:3200, plus a queue worker and the scheduler
```

Live reload: install [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`) and run `air`.

## Laravel to Go, at a glance

| Laravel | Here |
|---|---|
| `php artisan ...` | `go run . artisan ...` (or `./artisan ...`) |
| `php artisan make:controller/model/migration/job/command` | `go run . artisan make:...` |
| `php artisan migrate` | `go run . artisan migrate` |
| `php artisan serve` | `go run .` (or `air`) |
| `php artisan queue:work --queue=emails` | `go run . artisan queue:work --queue=emails` |
| `php artisan queue:failed` / `queue:retry all` | the same names |
| `routes/web.php`, `routes/api.php` | `routes/web.go`, `routes/api.go` |
| `app/Http/Controllers` | `app/http/controllers` |
| `Storage::disk('s3')` | `facades.Storage().Disk("r2_public")` |
| `php artisan test` | `go test ./...` |

## Tests

```sh
go vet ./...
go test ./... -count=1
```

Tests use `DB_TEST_DATABASE` (default `<DB_DATABASE>_test`). The R2 and backup tests write to the real buckets and
delete what they write; they are skipped when the keys are not in `.env`.

## Docs

- `AGENTS.md` and `CLAUDE.md`: how to work in this repository (also for AI assistants)
- `docs/CONVENTIONS.md`: the rules
- `tutorials/`: numbered walkthroughs, start at `tutorials/00-index.md`
