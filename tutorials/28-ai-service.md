<!-- nav:start -->
**Tutorial 28 of 30** | [Index](00-index.md) | [Previous: 27 PDF](27-pdf.md) | [Next: 29 Bilingual pages](29-bilingual-pages.md)

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
**28. [AI service](28-ai-service.md)**  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  

</details>
<!-- nav:end -->

# 28. AI service

```go
result, err := aiservice.Chat(ctx, aiservice.Request{
	Purpose:      "summarise contact message", // required: the AI API logs say what each call was for
	Instructions: "Summarise in two sentences.",
	Input:        message,
})

var person struct{ Name string `json:"name"`; Age int `json:"age"` }
_, err = aiservice.JSON(ctx, aiservice.Request{Purpose: "read ID card", Input: text, Sensitive: true}, &person)
```

## Pictures

```go
picture, err := aiservice.Image(ctx, aiservice.ImageRequest{
	Purpose: "product illustration", Prompt: "a red bicycle on a white background",
	Size: contractsai.ImageSizeLandscape, Quality: contractsai.ImageQualityMedium,
})
item, err := media.AddBytes(media.Owner{Type: "product", ID: "7"}, "photos", "bicycle.png", picture.MimeType, picture.Bytes)
```

The call is logged with kind `image` and its prompt; the picture itself is never stored in the log.

## Try it

```sh
go run . artisan ai-service:try-text  --prompt="Say hello in Bangla"
go run . artisan ai-service:try-image --prompt="A gopher and an elephant building a website" --output=storage/app/ai-image.png
```

Both are real calls on your key (paid). Find the model names your key may use with
`curl -s https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY"`.

## Details

- **Providers** are Goravel's packages, set in `.env`: `AI_PROVIDER` (anthropic, openai, gemini), the keys and
  `CLAUDE_MODEL`, `OPENAI_MODEL`, `GEMINI_MODEL`, and `OPENAI_IMAGE_MODEL` for pictures. `Request.Provider` and `Request.Model` override per call.
- **AI API logs:** every call is recorded by a Goravel AI middleware: provider, model, purpose, status, tokens,
  duration and estimated cost. Prices come from `AI_PRICES` (USD per million tokens); without them the cost stays empty
  rather than wrong.
- **Sensitive** calls (anything about a person) keep the prompt and reply out of the logs; the numbers stay.
- **JSON:** the reply is decoded into your struct; prose instead of JSON is an error. A model's reply is input, not
  truth: check the fields.
- **Admin:** `/admin/ai-api-logs` (`ai_api_logs.view`), delete with `ai_api_logs.manage`;
  `ai-service:purge-ai-api-logs --days=90` runs daily.
- **Tests** use provider `fake` (`app/support/aifake`, Laravel's `AI::fake()`), registered only under `go test`.

<!-- foot:start -->
---
[Previous: 27 PDF](27-pdf.md) | [Index](00-index.md) | [Next: 29 Bilingual pages](29-bilingual-pages.md)
<!-- foot:end -->
