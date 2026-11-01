package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/mail"

	"goravel/app/facades"
	"goravel/app/web"
)

// Outgoing is one invoice mail, its PDF already on disk at Attachment.
type Outgoing struct {
	To, Subject, HTML, Text, Attachment string
}

// SendMail sends at once through Goravel's mail. Not queued: the PDF is a temporary file, gone by the time a worker
// would run. Tests replace it.
var SendMail = func(message Outgoing) error {
	return facades.Mail().To([]string{message.To}).Subject(message.Subject).
		Content(mail.Content{Html: message.HTML, Text: message.Text}).Attach([]string{message.Attachment}).Send()
}

// Email renders the invoice to PDF and mails it, the PDF attached as <invoice number>.pdf. The PDF comes first, so
// a Chrome problem sends nothing.
func (i Invoice) Email(to string) error {
	document, err := i.PDF()
	if err != nil {
		return err
	}
	html, err := web.Render("pdf::invoice-mail", map[string]any{
		"seller": i.Seller, "customer": i.Customer, "number": i.Number, "currency": i.Currency,
		"total": i.Total().String(), "due": i.Due.Format("2 January 2006"),
	})
	if err != nil {
		return err
	}
	folder, err := os.MkdirTemp("", "invoice-mail-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(folder)
	attachment := filepath.Join(folder, i.Number+".pdf") // the mail shows the file's own name
	if err := os.WriteFile(attachment, document, 0o600); err != nil {
		return err
	}
	return SendMail(Outgoing{
		To: to, Subject: fmt.Sprintf("Invoice %s from %s", i.Number, i.Seller), HTML: html, Text: i.mailText(), Attachment: attachment,
	})
}

func (i Invoice) mailText() string {
	return fmt.Sprintf("Hello %s,\n\nYour invoice %s is attached as a PDF.\n\nAmount due: %s %s\nDue date:   %s\n\n"+
		"Thank you for your business.\nIf you have any questions, just reply to this email.\n\n%s\n",
		i.Customer, i.Number, i.Currency, i.Total(), i.Due.Format("2 January 2006"), i.Seller)
}
