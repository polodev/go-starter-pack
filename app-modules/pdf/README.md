# pdf

HTML and CSS to PDF with headless Chrome (chromedp). Page size and margins are CSS `@page`.

```go
document, err := pdf.FromHTML(html)
```

`go run . artisan pdf:sample-invoice --output=storage/app/invoice.pdf [--email=you@example.com]` renders the example
invoice (`pdf::invoice`, money in minor units) and can mail it as an attachment. `PDF_CHROME_PATH` when Chrome is not
in the usual place.
