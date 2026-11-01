<!-- nav:start -->
**Tutorial 25 of 29** | [Index](00-index.md) | [Previous: 24 Activity log](24-activity-log.md) | [Next: 26 Mail](26-mail.md)

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
**25. [Media library](25-media-library.md)**  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  

</details>
<!-- nav:end -->

# 25. Media library

```go
// in a module's Routes(): like registerMediaCollections()
media.Define("product", media.Multiple("photos").Images().AsPublic().KeepNewest(8), media.Single("manual"))

item, err := media.Add(media.Owner{Type: "product", ID: "7"}, "photos", file) // $product->addMedia($file)->toMediaCollection('photos')
items := media.Get(media.Owner{Type: "product", ID: "7"}, "photos")          // $product->getMedia('photos')
url := media.URL(item)                                                        // $media->getUrl()
```

- **Collections:** single (a new file replaces the old) or multiple (`KeepNewest`), accepted types, size limit
  (default 10 MB). The person's avatar is `media.Single("avatar").Images().AsPublic().Max(2<<20)` on `/account/avatar`.
- **Public or private:** public files are served from their public address; private ones only through
  `/media/{id}/{file_name}` to the owner and staff with `media.view`. Unknown and forbidden files both answer 404.
- **Where:** `MEDIA_STORAGE=local` (this server) or `r2` (public collections in `R2_PUBLIC_BUCKET`, private ones in
  `R2_PRIVATE_BUCKET`, never mixed). Files sit at `<MEDIA_PREFIX>/<id>/<file_name>`, Spatie's layout.
- **Admin:** `/admin/media` lists every file; upload (private or public) and delete need `media.manage`.
- The table is Spatie's `media`, so files imported from a Laravel app are found; without a `public` custom property
  they are private.

<!-- foot:start -->
---
[Previous: 24 Activity log](24-activity-log.md) | [Index](00-index.md) | [Next: 26 Mail](26-mail.md)
<!-- foot:end -->
