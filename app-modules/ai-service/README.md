# ai-service

The one way to call AI models, on Goravel's AI SDK (providers `anthropic`, `openai`, `gemini` in `config/ai.go`).

```go
result, err := aiservice.Chat(ctx, aiservice.Request{Purpose: "summarise message", Instructions: "...", Input: text})
_, err = aiservice.JSON(ctx, aiservice.Request{Purpose: "extract person", Input: text, Sensitive: true}, &person)
```

```go
picture, err := aiservice.Image(ctx, aiservice.ImageRequest{Purpose: "product illustration", Prompt: "a red bicycle"})
// picture.Bytes, picture.MimeType; store it with media.AddBytes(...)
```

Try it with real calls (paid): `go run . artisan ai-service:try-text --prompt="..."` and
`go run . artisan ai-service:try-image --prompt="..." --output=storage/app/ai-image.png`. Image models: OpenAI
(`OPENAI_IMAGE_MODEL`, e.g. `gpt-image-2`) and Gemini.

- Every call needs a purpose and is recorded in `ai_api_logs` (a Goravel AI middleware): provider, model, status,
  kind (text or image), tokens, duration, estimated cost (`AI_PRICES`), and the prompt and reply unless `Sensitive` or `AI_LOG_CONTENT=false`.
- Admin: `GET /admin/ai-api-logs`, `/admin/ai-api-logs/{id}` (`ai_api_logs.view`), delete (`ai_api_logs.manage`).
- `ai-service:purge-ai-api-logs --days=90` runs daily. Tests use provider `fake` (`app/support/aifake`).
