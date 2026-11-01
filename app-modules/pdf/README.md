# pdf

HTML and CSS to PDF with headless Chrome (chromedp). Page size and margins are CSS `@page`.

```go
document, err := pdf.FromHTML(html)
```

`go run . artisan pdf:sample-invoice --output=storage/app/invoice.pdf` renders the example invoice (`pdf::invoice`,
money in minor units). `PDF_CHROME_PATH` when Chrome is not in the usual place.

`invoice.Email(to)` renders the PDF first, then sends the mail `pdf::invoice-mail` (HTML and text) with the PDF
attached as `<invoice number>.pdf`. It sends at once, because the PDF is a temporary file.

```sh
go run . artisan pdf:sample-invoice --email=you@example.com                  # through MAIL_HOST (the catcher locally)
MAIL_MAILER=ses go run . artisan pdf:sample-invoice --email=you@example.com  # really send through Amazon SES
```
