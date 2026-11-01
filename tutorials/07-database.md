<!-- nav:start -->
**Tutorial 7 of 31** | [Index](00-index.md) | [Previous: 6 Routes, controllers and views](06-routes-controllers-views.md) | [Next: 8 Validation and JSON responses](08-validation-and-json.md)

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
06. [Routes, controllers and views](06-routes-controllers-views.md)  
**07. [Database: migrations, models, queries, seeding](07-database.md)**  
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
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  

</details>
<!-- nav:end -->

# 7. Database: migrations, models, queries, seeding

## Migrations

```sh
go run . artisan make:migration create_orders_table
go run . artisan migrate            # migrate:status, migrate:rollback, migrate:fresh
```

```go
facades.Schema().Create("orders", func(table schema.Blueprint) {
	table.ID()
	table.UnsignedBigInteger("user_id")
	table.String("status", 20).Default("new")
	table.Decimal("total")              // money: prefer integers in minor units
	table.TimestampsTz()
	table.Index("user_id")
})
```

One migration per table. Once a migration has run on a shared or production database, change the table with a new
migration, never by editing the old one.

## Models

```sh
go run . artisan make:model Order
```

```go
type Order struct {
	orm.Model                       // id, created_at, updated_at
	UserID uint   `json:"user_id"`
	Status string `json:"status"`
}
```

Table names follow Laravel: `Order` -> `orders`.

## Queries (Eloquent next to Goravel)

| Eloquent | Goravel |
|---|---|
| `Order::where('status','new')->get()` | `facades.Orm().Query().Where("status", "new").Get(&orders)` |
| `Order::find(7)` | `facades.Orm().Query().Find(&order, 7)` |
| `Order::where(...)->first()` | `...Where(...).First(&order)` |
| `Order::create([...])` | `facades.Orm().Query().Create(&order)` |
| `->update(['status'=>'paid'])` | `...Where("id", 7).Update("status", "paid")` |
| `->count()`, `->exists()` | `.Count()`, `.Exists()` |
| `->orderBy('id','desc')` | `.OrderBy("id", "desc")` |
| `->paginate(20)` | `.Paginate(page, 20, &orders, &total)` |
| `DB::table('jobs')->count()` | `facades.DB().Table("jobs").Count()` |
| `DB::transaction(fn)` | `facades.Orm().Transaction(func(tx orm.Query) error { ... })` |

Every call that touches the database returns an `error`; check it.

## Seeding

```sh
go run . artisan make:seeder OrderSeeder
go run . artisan db:seed
```

Other useful commands: `db:show`, `db:table orders`.

<!-- foot:start -->
---
[Previous: 6 Routes, controllers and views](06-routes-controllers-views.md) | [Index](00-index.md) | [Next: 8 Validation and JSON responses](08-validation-and-json.md)
<!-- foot:end -->
