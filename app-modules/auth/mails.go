package auth

import (
	"github.com/goravel/framework/contracts/mail"

	"goravel/app/facades"
	"goravel/app/web"
)

// Outgoing is one email the auth module sends.
type Outgoing struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// SendMail queues an email with Goravel's mail (facades.Mail().Queue(), so the request does not wait for SMTP).
// Tests replace it to capture the mail instead.
var SendMail = func(message Outgoing) error {
	return facades.Mail().To([]string{message.To}).Subject(message.Subject).
		Content(mail.Content{Html: message.HTML, Text: message.Text}).Queue()
}

// sendLink emails a one-time link (verify the address, reset the password) in the visitor's language.
func sendLink(lang, to, purpose, link string) error {
	data := map[string]any{
		"lang": lang, "link": link, "purpose": purpose,
		"appName": facades.Config().GetString("app.name"),
		"minutes": int(lifetimes[purpose].Minutes()),
	}
	html, err := web.Render("auth::mail-link", data)
	if err != nil {
		return err
	}
	text, err := web.Render("auth::mail-link-text", data)
	if err != nil {
		return err
	}
	subjectKey := "auth.mail-subject-verify"
	if purpose == PurposeResetPassword {
		subjectKey = "auth.mail-subject-reset"
	}
	return SendMail(Outgoing{To: to, Subject: web.Translate(lang, subjectKey), HTML: html, Text: text})
}

// linkURL is the absolute address of a link page.
func linkURL(path string) string {
	return facades.Config().GetString("http.url") + path
}
