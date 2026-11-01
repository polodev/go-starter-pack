# Agent guide for this Goravel app

This is a [Goravel](https://www.goravel.dev) app: **Laravel for Go**. When you are unsure how something should work
here, the answer is almost always "the way Laravel does it", in Goravel's names.

## Read this first

- `.agents/skills/goravel-development/SKILL.md`: Goravel's own skill (structure, facades, generators, testing).
  Its `CUSTOM.md` points back here.
- `docs/CONVENTIONS.md`: the rules of this starter pack. Follow them, and update them in the same change when you add
  a new convention.
- `tutorials/`: numbered walkthroughs for Laravel developers.

## The rules that prevent most mistakes

1. **Generate, then edit.** `go run . artisan make:<thing>` writes the file and registers it in `bootstrap/`
   (commands, jobs, migrations). Run `go run . artisan list | grep make:` to see the generators.
2. **Use the facades.** `facades.Orm()`, `facades.Queue()`, `facades.Storage()`, `facades.Schedule()`, ... Read the
   interface in `github.com/goravel/framework/contracts/<module>` (version from `go.mod`) before guessing a name.
3. **Fat model or service, slim controller.** Domain logic lives in `app/services/<area>` or on the model.
4. **Never test against a real database.** `config/database.go` switches to `DB_TEST_DATABASE` inside `go test` and
   panics unless the name ends in `_test`.
5. **R2 has two disks:** `r2_public` (uploads, public addresses) and `r2_backups` (private). Never put private files on
   the public disk. Both go through `app/storage.R2`, which fixes the public URL the MinIO driver gets wrong for R2.

## Before you call it done

```sh
gofmt -l .        # prints nothing
go vet ./...
go test ./... -count=1
```

Then update `README.md`, `docs/CONVENTIONS.md` and the matching tutorial.
