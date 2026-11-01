<!-- nav:start -->
**Tutorial 3 of 29** | [Index](00-index.md) | [Previous: 2 Setup](02-setup.md) | [Next: 4 Go tooling jargon](04-go-tooling.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
**03. [Go essentials for Laravel developers](03-go-essentials.md)**  
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

</details>
<!-- nav:end -->

# 3. Go essentials for Laravel developers

Go is small. These are the parts this project uses, next to the PHP you know.

## Packages and imports

Each folder is a package. `goravel/app/facades` is the folder `app/facades` in the module `goravel` (see `go.mod`).

```go
import (
	"strings"                          // standard library
	"github.com/goravel/framework/contracts/http"
	"goravel/app/facades"              // this project
)
```

A name that starts with a capital letter is public (`Run`, `User`); lower-case is private to the package (`dumpDatabase`).

## Variables and types

```go
name := "Polo"                 // declare and assign; the type is inferred
var count int64                // zero value 0
tags := []string{"a", "b"}     // slice (a growable list)
ages := map[string]int{"polo": 30}
```

## Structs and methods (instead of classes)

```go
type User struct {
	orm.Model                  // embedding: User gets ID, CreatedAt, UpdatedAt
	Name     string `json:"name"`
	Password string `json:"-"` // never in JSON
}

func (u *User) Greeting() string { return "Hello " + u.Name }
```

`*User` is a pointer. Methods on a pointer can change the struct, and the ORM needs pointers so it can fill them:
`facades.Orm().Query().First(&user)`.

## Errors are values

There are no exceptions. A function that can fail returns an `error` last, and you check it at once:

```go
user, err := findUser(email)
if err != nil {
	return err                 // or answer 500, or log it
}
```

## Interfaces

An interface is a set of methods. Any type with those methods satisfies it, with no `implements` keyword. A job is any
type with `Signature() string` and `Handle(args ...any) error`.

## `defer`, goroutines and channels

```go
defer file.Close()             // runs when the function returns, like finally
go worker.Run()                // run concurrently (a goroutine)
done := make(chan string)      // a channel passes values between goroutines
```

The tests use exactly these three for the queue worker.

## Formatting

`gofmt` decides how code looks. Run `gofmt -l .`: it prints the files that need formatting (`gofmt -w .` fixes them).

<!-- foot:start -->
---
[Previous: 2 Setup](02-setup.md) | [Index](00-index.md) | [Next: 4 Go tooling jargon](04-go-tooling.md)
<!-- foot:end -->
