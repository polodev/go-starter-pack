<!-- nav:start -->
**Tutorial 27 of 29** | [Index](00-index.md) | [Previous: 26 Mail](26-mail.md) | [Next: 28 AI service](28-ai-service.md)

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
**27. [PDF](27-pdf.md)**  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  

</details>
<!-- nav:end -->

# 27. PDF

```go
document, err := pdf.FromHTML(html) // bytes of a PDF
```

Chrome prints the page like "Save as PDF": page size and margins come from CSS:

```css
@page { size: A4; margin: 18mm; }
```

- Templates: write the document as a template (`pdf::invoice`) and render it with `web.Render`.
- Money: keep it in minor units (`pdf.Money(4500000)` prints `45,000.00`), never floats.
- Try it: `go run . artisan pdf:sample-invoice --output=storage/app/invoice.pdf --email=you@example.com`.
- Chrome must be installed on the server (`apt install chromium`); set `PDF_CHROME_PATH` if it is not found.
- To keep a generated PDF, store it with `media.AddBytes(owner, "invoices", "invoice.pdf", "application/pdf", document)`.

<!-- foot:start -->
---
[Previous: 26 Mail](26-mail.md) | [Index](00-index.md) | [Next: 28 AI service](28-ai-service.md)
<!-- foot:end -->
