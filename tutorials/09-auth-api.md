<!-- nav:start -->
**Tutorial 9 of 18** | [Index](00-index.md) | [Previous: 8 Validation and JSON responses](08-validation-and-json.md) | [Next: 10 Queues and jobs](10-queues-and-jobs.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
06. [Routes, controllers and views](06-routes-controllers-views.md)  
07. [Database: migrations, models, queries, seeding](07-database.md)  
08. [Validation and JSON responses](08-validation-and-json.md)  
**09. [Token API: register, login, me](09-auth-api.md)**  
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

# 9. Token API: register, login, me

```sh
curl -s -X POST localhost:3200/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"name":"Polo","email":"polo@example.com","password":"secret-password"}'
# 201 {"user": {...}, "token": "..."}

curl -s -X POST localhost:3200/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"polo@example.com","password":"secret-password"}'
# 200 {"user": {...}, "token": "..."}

curl -s localhost:3200/api/v1/auth/me -H "Authorization: Bearer <token>"
# 200 {"data": {...}}         no or bad token: 401 {"message": "Unauthenticated."}
```

| Piece | File |
|---|---|
| routes | `routes/api.go` |
| handlers | `app/http/controllers/auth_controller.go` |
| bearer check | `app/http/middleware/authenticate.go` |
| user | `app/models/user.go`, table `users` (Laravel's columns) |

**Security choices:**

- Emails are stored lower-case and the database keeps them unique; two sign-ups at once cannot create twins.
- A wrong password and an unknown email get the same answer, so the API never reveals who has an account.
- Passwords are bcrypt through `facades.Hash()`, Laravel's default, so users imported from a Laravel database sign in
  with their old passwords.
- Tokens are JWTs signed with `JWT_SECRET` (`config/jwt.go` sets their lifetime).

**Protect your own route:** `router.Middleware(&middleware.Authenticate{}).Get("/orders", orders.Index)`, then
`facades.Auth(ctx).User(&user)` inside the handler.

Not yet: throttling failed logins, web login pages with sessions, email verification and password reset (the Rust
starter has all of them).

<!-- foot:start -->
---
[Previous: 8 Validation and JSON responses](08-validation-and-json.md) | [Index](00-index.md) | [Next: 10 Queues and jobs](10-queues-and-jobs.md)
<!-- foot:end -->
