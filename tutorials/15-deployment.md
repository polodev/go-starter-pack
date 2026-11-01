<!-- nav:start -->
**Tutorial 15 of 18** | [Index](00-index.md) | [Previous: 14 Testing](14-testing.md) | [Next: 16 MySQL or PostgreSQL](16-mysql-or-postgresql.md)

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
14. [Testing](14-testing.md)  
**15. [Deployment with Supervisor](15-deployment.md)**  
16. [MySQL or PostgreSQL](16-mysql-or-postgresql.md)  
17. [Conventions](17-conventions.md)  
18. [Cheat sheet and what to build next](18-cheatsheet-and-next-steps.md)  

</details>
<!-- nav:end -->

# 15. Deployment with Supervisor

## Build

```sh
go build -o go-starter-pack .                       # on the server, or
GOOS=linux GOARCH=amd64 go build -o go-starter-pack .   # cross-compile on your Mac
```

The binary contains the app; the server needs it, `.env`, `resources/` (views), `public/` and `storage/`.

## First deploy

```sh
git clone ... /var/www/app && cd /var/www/app
cp .env.example .env              # APP_ENV=production, APP_DEBUG=false, real DB, R2 and backup keys
./go-starter-pack artisan key:generate
./go-starter-pack artisan jwt:secret
./go-starter-pack artisan migrate
sudo cp deploy/supervisor.conf /etc/supervisor/conf.d/app.conf
sudo supervisorctl reread && sudo supervisorctl update
```

`deploy/supervisor.conf` has two programs:

| Program | Runs |
|---|---|
| `app-web` | the site, one queue worker and the scheduler |
| `app-queue-worker` | an extra `queue:work` (add one per busy queue) |

Put Nginx (or Caddy) in front of `APP_HOST:APP_PORT` for HTTPS.

## Every deploy

```sh
git pull --ff-only
go build -o go-starter-pack .
./go-starter-pack artisan migrate
sudo supervisorctl restart app-web app-queue-worker
```

Keep exactly one `app-web` process, or the scheduler runs each task once per process.

<!-- foot:start -->
---
[Previous: 14 Testing](14-testing.md) | [Index](00-index.md) | [Next: 16 MySQL or PostgreSQL](16-mysql-or-postgresql.md)
<!-- foot:end -->
