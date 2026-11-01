<!-- nav:start -->
**Tutorial 14 of 33** | [Index](00-index.md) | [Previous: 13 Database backups](13-database-backups.md) | [Next: 15 Deployment with Supervisor](15-deployment.md)

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
09. [Token API: register, login, me](09-auth-api.md)  
10. [Queues and jobs](10-queues-and-jobs.md)  
11. [Console commands and the scheduler](11-console-commands-and-scheduler.md)  
12. [Files on Cloudflare R2](12-files-on-r2.md)  
13. [Database backups](13-database-backups.md)  
**14. [Testing](14-testing.md)**  
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
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  
32. [Contact form](32-contact-form.md)  
33. [Sign in with Google](33-sign-in-with-google.md)  

</details>
<!-- nav:end -->

# 14. Testing

```sh
go test ./... -count=1                                   # everything (php artisan test)
go test ./tests/feature -run TestQueueTestSuite -v -count=1
```

Use `-count=1`: Go caches passing test results, and our tests talk to the database and R2.

## A test

```sh
go run . artisan make:test feature/OrderTest
```

```go
type OrderTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestOrderTestSuite(t *testing.T) { suite.Run(t, new(OrderTestSuite)) }

func (s *OrderTestSuite) SetupTest() { s.RefreshDatabase() }   // fresh tables for every test

func (s *OrderTestSuite) TestAGuestCannotSeeOrders() {
	response, err := s.Http(s.T()).Get("/api/v1/orders")
	s.Require().NoError(err)
	response.AssertStatus(401)
	response.AssertJson(map[string]any{"message": "Unauthenticated."})
}
```

`s.Http(s.T())` sends requests to the app without a network: `.WithHeader(...)`, `.WithToken(token)`, `.Post(uri, body)`.

## The test database guard

`RefreshDatabase` drops every table, so tests must never touch your real database:

- inside `go test`, `config/database.go` switches to `DB_TEST_DATABASE` (default `<DB_DATABASE>_test`) and panics unless
  the name ends in `_test`;
- `tests/test_case.go` checks again after boot.

The check lives in `config/database.go` because Goravel reads the config before any test code runs.

## What to cover

The happy path, validation errors (422), refusals (401), and "nothing was stored after a refusal". Tests that need
outside services (R2) skip themselves when the keys are missing and delete everything they write.

<!-- foot:start -->
---
[Previous: 13 Database backups](13-database-backups.md) | [Index](00-index.md) | [Next: 15 Deployment with Supervisor](15-deployment.md)
<!-- foot:end -->
