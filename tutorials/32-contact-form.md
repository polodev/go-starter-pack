<!-- nav:start -->
**Tutorial 32 of 32** | [Index](00-index.md) | [Previous: 31 Queue monitor](31-queue-monitor.md) | Next: none

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
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  
**32. [Contact form](32-contact-form.md)**  

</details>
<!-- nav:end -->

# 32. Contact form

`app-modules/contact` is the example of a public form done properly. Copy its shape for any form anyone can post:
sign-ups for a newsletter, quote requests, feedback.

## The route

```go
facades.RateLimiter().For("contact", func(ctx http.Context) http.Limit {
	return limit.PerMinutes(10, 5).By(ctx.Request().Ip()).Response(/* the form again, 429, values kept */)
})

web.Group(func(router route.Router) {          // session, CSRF, language
	router.Get("/contact", ShowForm)
	router.Middleware(middleware.Throttle("contact")).Post("/contact", Send)
})
```

Laravel: `Route::post('/contact', ...)->middleware('throttle:5,10')` inside the `web` group.

## The handler, in order

```go
func Send(ctx http.Context) http.Response {
	message := readMessage(ctx)                                   // trimmed input
	if ctx.Request().Input("website") != "" {                     // honeypot filled: a bot
		return redirect("/contact?status=sent")                     // looks like success, stores nothing
	}
	if problems := message.Validate(); len(problems) > 0 {
		return page(ctx, 422, message, problems, "")               // errors, typed values kept
	}
	// store, tell the staff (queued mail), then Post/Redirect/Get
	return redirect("/contact?status=sent")
}
```

1. **CSRF** is checked before the handler by the web group: a post without `_token` gets 419.
2. **Rate limit** runs before the handler too: the sixth post in ten minutes from one address gets 429.
3. **Honeypot**: the `website` field sits off-screen (`absolute -left-[10000px] h-px w-px overflow-hidden`,
   `aria-hidden`, `tabindex="-1"`). People never fill it; form-filling bots do. Answer exactly like a success so the
   bot learns nothing.
4. **Validate on the server** and count characters, not bytes: `utf8.RuneCountInString`, so "রহিম" is 4 characters,
   not 12. Lengths: name 2 to 80, a real email (`net/mail`, and the domain needs a dot), subject up to 150, message 10
   to 5000.
5. **Redirect after success** (303), so a reload does not send it twice.

## The inbox

`/admin/contacts` (sidebar: Content › Contact messages): newest first, filter by status (`new`, `replied`), search
name, email and subject (case-insensitive on MySQL and PostgreSQL alike: `LOWER(...) LIKE`). The show page has the
whole message, a "Reply by email" `mailto:` link with `Re: <subject>`, the status form and Delete.

| Permission | Who | What |
|---|---|---|
| `contacts.view` | business development and above | read |
| `contacts.manage` | employee and above | set the status, delete |

## Settings

```dotenv
CONTACT_NOTIFICATION_EMAIL=hello@example.com   # a queued plain-text mail per new message; unset: none
```

## What the tests prove

`tests/feature/contact_test.go`: a message is stored (email lower-cased, language kept) and the staff is told; Bangla
text and errors, with Bangla names measured in characters; invalid input answers 422 with what was typed; the
honeypot looks like success and stores nothing; the sixth post in ten minutes gets 429 with the message kept; no
CSRF token, no message; customers and new employees get 403 on the inbox; business development reads but cannot
change; an employee sets "replied", cannot set an unknown status, searches regardless of case, deletes; both changes
are in the activity log.

<!-- foot:start -->
---
[Previous: 31 Queue monitor](31-queue-monitor.md) | [Index](00-index.md) | Next: none
<!-- foot:end -->
