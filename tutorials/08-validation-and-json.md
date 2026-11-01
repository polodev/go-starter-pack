<!-- nav:start -->
**Tutorial 8 of 18** | [Index](00-index.md) | [Previous: 7 Database: migrations, models, queries, seeding](07-database.md) | [Next: 9 Token API: register, login, me](09-auth-api.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
06. [Routes, controllers and views](06-routes-controllers-views.md)  
07. [Database: migrations, models, queries, seeding](07-database.md)  
**08. [Validation and JSON responses](08-validation-and-json.md)**  
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

# 8. Validation and JSON responses

## Validate a request

```go
validator, err := ctx.Request().Validate(map[string]any{
	"name":     "required|max_len:255",
	"email":    "required|email|max_len:255",
	"password": "required|min_len:8",
})
if err != nil { /* the body could not be read */ }
if validator.Fails() {
	return responses.ValidationFailed(ctx, responses.FromValidator(validator.Errors().All()))
}
```

Rule names differ a little from Laravel: `max_len`/`min_len` for strings. For a reusable rule set, generate a form
request: `go run . artisan make:request StoreOrderRequest`.

## The JSON shapes

`app/http/responses/json.go` answers exactly like Laravel, so a client written for Laravel works unchanged:

| Case | Status | Body |
|---|---|---|
| one thing or a list | 200 | `{"data": ...}` |
| sign-in | 200/201 | `{"user": ..., "token": "..."}` |
| not signed in | 401 | `{"message": "Unauthenticated."}` |
| validation | 422 | `{"message": "The email field is required. (and 2 more errors)", "errors": {"email": ["..."]}}` |

Never return password hashes or secrets: give such fields `json:"-"`.

<!-- foot:start -->
---
[Previous: 7 Database: migrations, models, queries, seeding](07-database.md) | [Index](00-index.md) | [Next: 9 Token API: register, login, me](09-auth-api.md)
<!-- foot:end -->
