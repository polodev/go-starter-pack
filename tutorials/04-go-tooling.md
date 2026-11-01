<!-- nav:start -->
**Tutorial 4 of 18** | [Index](00-index.md) | [Previous: 3 Go essentials for Laravel developers](03-go-essentials.md) | [Next: 5 Project structure](05-project-structure.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
**04. [Go tooling jargon](04-go-tooling.md)**  
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
18. [Cheat sheet and what to build next](18-cheatsheet-and-next-steps.md)  

</details>
<!-- nav:end -->

# 4. Go tooling jargon

**What is `go`?** One command for building, testing, formatting and managing dependencies. In Laravel terms:
Composer + `php` + PHPUnit, in one tool.

| Command | What it does | Laravel feel |
|---|---|---|
| `go run .` | compile and run the program in this folder | `php artisan serve` |
| `go build -o app .` | compile a single binary called `app` | none (PHP is not compiled) |
| `go test ./...` | run every test in every package (`./...` = this folder and below) | `php artisan test` |
| `go test ./tests/feature -run TestQueue -v` | run matching tests, verbose | `--filter` |
| `go vet ./...` | find likely bugs the compiler allows | PHPStan |
| `gofmt -l .` / `gofmt -w .` | list / fix unformatted files | Pint |
| `go mod tidy` | add missing and remove unused dependencies | `composer install` + cleanup |
| `go get example.com/pkg@v1.2.3` | add or change one dependency | `composer require` |
| `go install example.com/tool@latest` | install a command-line tool into `~/go/bin` | `composer global require` |
| `go clean -cache` | delete the build cache | clearing caches |

| Word | Meaning |
|---|---|
| **module** | a project with a `go.mod`; ours is called `goravel` | 
| **package** | one folder of `.go` files with the same `package` line |
| **`go.mod` / `go.sum`** | dependencies and their exact checksums (`composer.json` / `composer.lock`) |
| **module cache** | downloaded dependencies, in `~/go/pkg/mod`, shared by all projects (like Composer's cache) |
| **build cache** | compiled pieces in `~/Library/Caches/go-build`; makes rebuilds fast |
| **`-count=1`** | run tests even if the cache says nothing changed (use it: our tests talk to R2 and the database) |

Go keeps no per-project build folder like Rust's `target/`; the binary is the only output.

<!-- foot:start -->
---
[Previous: 3 Go essentials for Laravel developers](03-go-essentials.md) | [Index](00-index.md) | [Next: 5 Project structure](05-project-structure.md)
<!-- foot:end -->
