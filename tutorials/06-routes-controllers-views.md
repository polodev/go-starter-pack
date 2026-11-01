<!-- nav:start -->
**Tutorial 6 of 18** | [Index](00-index.md) | [Previous: 5 Project structure](05-project-structure.md) | [Next: 7 Database: migrations, models, queries, seeding](07-database.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
**06. [Routes, controllers and views](06-routes-controllers-views.md)**  
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

# 6. Routes, controllers and views

## Routes

```go
// routes/web.go
facades.Route().Get("/", func(ctx http.Context) http.Response {
	return ctx.Response().View().Make("welcome.tmpl", map[string]any{"version": support.Version})
})

// routes/api.go: a group with a prefix and middleware
facades.Route().Prefix("/api/v1").Group(func(router route.Router) {
	router.Post("/auth/login", auth.Login)
	router.Middleware(&middleware.Authenticate{}).Get("/auth/me", auth.Me)
})
```

List them: the startup log of `go run .` prints every route.

## Controllers

```sh
go run . artisan make:controller OrderController
```

A handler takes the context and returns a response:

```go
func (r *OrderController) Show(ctx http.Context) http.Response {
	id := ctx.Request().Route("id")             // /orders/{id}
	page := ctx.Request().Query("page", "1")    // ?page=2
	name := ctx.Request().Input("name")         // form or JSON body
	return ctx.Response().Json(http.StatusOK, http.Json{"data": id, "page": page, "name": name})
}
```

## Middleware

A middleware is a type with `Signature()` and `Handle(ctx)`. Call `ctx.Request().Next()` to continue, or
`ctx.Request().AbortWithStatusJson(401, body)` to stop. See `app/http/middleware/authenticate.go`.

## Views

Templates are Go's `html/template` files in `resources/views/*.tmpl`, escaped by default (like Blade's `{{ }}`).
`go run . artisan make:view orders/index` creates one. Static files are served from `public/`.

<!-- foot:start -->
---
[Previous: 5 Project structure](05-project-structure.md) | [Index](00-index.md) | [Next: 7 Database: migrations, models, queries, seeding](07-database.md)
<!-- foot:end -->
