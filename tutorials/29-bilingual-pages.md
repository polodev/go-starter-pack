<!-- nav:start -->
**Tutorial 29 of 30** | [Index](00-index.md) | [Previous: 28 AI service](28-ai-service.md) | [Next: 30 Security headers](30-security-headers.md)

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
**29. [Bilingual pages](29-bilingual-pages.md)**  
30. [Security headers](30-security-headers.md)  

</details>
<!-- nav:end -->

# 29. Bilingual pages

- **Content:** each page has a slug and one row per language. English needs a title and text; Bangla is optional.
  An empty Bangla part (title, description or text) shows the English part, each on its own, with a small note.
  Emptying a whole language deletes its row.
- **Public:** `/pages/{slug}` in the visitor's language; drafts answer 404.
- **Admin:** `/admin/pages` (`pages.view`), create and edit (`pages.manage`), every save in the activity log.
- **HTML:** staff write simple HTML (headings, lists, links). It is cleaned with bluemonday before it is stored:
  scripts, `onclick` and `javascript:` links are removed. Only cleaned HTML is shown as HTML.
- **Seed:** `go run . artisan db:seed --force` adds About, Privacy policy and Terms of service in both languages
  (Goravel's seeders; each module offers its own through `Seeders()`).

```go
page, _ := pages.FindBySlug("about")
content := pages.Load(page, "bn") // content.Title, content.Description, content.Body, content.Fallback
```

<!-- foot:start -->
---
[Previous: 28 AI service](28-ai-service.md) | [Index](00-index.md) | [Next: 30 Security headers](30-security-headers.md)
<!-- foot:end -->
