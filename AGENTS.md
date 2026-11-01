# Agent guide for this Goravel app

This is a [Goravel](https://www.goravel.dev) app: **Laravel for Go**. When you are unsure how something should work
here, the answer is almost always "the way Laravel does it", in Goravel's names.

## Read this first

- `.agents/skills/goravel-development/SKILL.md`: Goravel's own skill (structure, facades, generators, testing).
  Its `CUSTOM.md` points back here.
- `docs/CONVENTIONS.md`: the rules of this starter pack. Follow them, and update them in the same change when you add
  a new convention.
- `tutorials/`: numbered walkthroughs for Laravel developers.

## Modules (like internachi/modular)

Features live in `app-modules/<name>`, one Go package each, with a `Module` (see `app/modular`) that hands the app its
routes, migrations, commands, jobs, schedules, seeders, views and admin menu. `bootstrap/modules.go` lists them;
`go run . artisan make:module <name>` creates and registers one. A module's templates are named `"<module>::<view>"`
and its translations are `lang/<locale>.json`, read as `"<module>.<key>"`. Dependency direction (no cycles):
`site-layout`, `auth` < `role-permission` < `admin-layout` < everything with admin screens.

## The rules that prevent most mistakes

1. **Generate, then edit.** `go run . artisan make:<thing>` writes the file and registers it in `bootstrap/`
   (commands, jobs, migrations). Run `go run . artisan list | grep make:` to see the generators.
2. **Use the facades.** `facades.Orm()`, `facades.Queue()`, `facades.Storage()`, `facades.Schedule()`, ... Read the
   interface in `github.com/goravel/framework/contracts/<module>` (version from `go.mod`) before guessing a name.
3. **Fat model or service, slim controller.** Domain logic lives in `app/services/<area>` or on the model.
4. **Never test against a real database.** `config/database.go` switches to `DB_TEST_DATABASE` inside `go test` and
   panics unless the name ends in `_test`.
5. **Use Goravel's batteries before building**: mail, translation, sessions, CSRF, rate limiting, Gate, AI SDK,
   queue, scheduler, storage. Build only what Goravel lacks, and say why in the module comment.
6. **Admin routes go through `rolepermission.Admin`** with a permission each; `role-permission:audit` must pass.
7. **R2 has three disks:** `r2_public` (public uploads), `r2_private` (private uploads) and `r2_backups` (backups). Never put
   private files on the public disk. Both go through `app/storage.R2`, which fixes the public URL the MinIO driver gets wrong for R2.

## Before you call it done

```sh
gofmt -l .        # prints nothing
go vet ./...
go test ./... -count=1
```

Then update `README.md`, `docs/CONVENTIONS.md` and the matching tutorial.
