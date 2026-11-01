# Project rules (go-starter-pack)

Read `AGENTS.md` and `docs/CONVENTIONS.md` at the repository root and follow them. In short: generate with
`go run . artisan make:*` then edit; JSON API under `/api/v1` with Laravel's shapes; tests run only against a
`*_test` database; R2 disks `r2_public` and `r2_backups` through `app/storage.R2`; run gofmt, go vet and go test
before finishing.
