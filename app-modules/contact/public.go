package contact

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/goravel/framework/contracts/http"
	contractsmail "github.com/goravel/framework/contracts/mail"

	"goravel/app-modules/contact/models"
	"goravel/app/facades"
	"goravel/app/web"
)

// Message is what a visitor typed, trimmed.
type Message struct {
	Name, Email, Subject, Body string
}

func readMessage(ctx http.Context) Message {
	input := func(field string) string { return strings.TrimSpace(ctx.Request().Input(field)) }
	return Message{Name: input("name"), Email: input("email"), Subject: input("subject"), Body: input("message")}
}

// Validate returns the translation key of the problem with each field (empty when the message is fine). Lengths
// are counted in characters, not bytes, so Bangla text gets the same limits as English.
func (m Message) Validate() map[string]string {
	problems := map[string]string{}
	if length := utf8.RuneCountInString(m.Name); length < 2 || length > 80 {
		problems["name"] = "contact.error-name"
	}
	if !validEmail(m.Email) {
		problems["email"] = "contact.error-email"
	}
	if utf8.RuneCountInString(m.Subject) > 150 {
		problems["subject"] = "contact.error-subject"
	}
	if length := utf8.RuneCountInString(m.Body); length < 10 || length > 5000 {
		problems["message"] = "contact.error-message"
	}
	return problems
}

func validEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	_, domain, _ := strings.Cut(email, "@")
	return strings.Contains(domain, ".") && !strings.HasSuffix(domain, ".")
}

// page renders the contact page with the typed values and translated errors.
func page(ctx http.Context, code int, message Message, problems map[string]string, failure string) http.Response {
	errors := map[string]string{}
	for field, key := range problems {
		errors[field] = web.T(ctx, key)
	}
	data := map[string]any{
		"title":  web.T(ctx, "contact.title"),
		"values": map[string]string{"name": message.Name, "email": message.Email, "subject": message.Subject, "message": message.Body},
		"errors": errors, "sent": ctx.Request().Query("status") == "sent",
	}
	if failure != "" {
		data["failure"] = web.T(ctx, failure)
	}
	if code == http.StatusOK {
		return web.View(ctx, "contact::page", data)
	}
	return web.ViewStatus(ctx, code, "contact::page", data)
}

// ShowForm is GET /contact.
func ShowForm(ctx http.Context) http.Response {
	return page(ctx, http.StatusOK, Message{}, nil, "")
}

// Send is POST /contact: CSRF-checked by the web group, throttled per address (the "contact" limiter), validated on
// the server; a filled honeypot gets the same answer as a success and nothing is stored.
func Send(ctx http.Context) http.Response {
	message := readMessage(ctx)
	if strings.TrimSpace(ctx.Request().Input("website")) != "" {
		return ctx.Response().Redirect(http.StatusSeeOther, "/contact?status=sent")
	}
	if problems := message.Validate(); len(problems) > 0 {
		return page(ctx, http.StatusUnprocessableEntity, message, problems, "")
	}
	stored := models.Contact{
		Name: message.Name, Email: strings.ToLower(message.Email), Message: message.Body, Status: models.StatusNew,
		Locale: web.CurrentLocale(ctx), IPAddress: optional(ctx.Request().Ip(), 45),
		UserAgent: optional(ctx.Request().Header("User-Agent"), 255),
	}
	if message.Subject != "" {
		stored.Subject = &message.Subject
	}
	if err := facades.Orm().Query().Create(&stored); err != nil {
		facades.Log().Request(ctx.Request()).Error("contact: could not store a message: " + err.Error())
		return page(ctx, http.StatusInternalServerError, message, nil, "contact.error-server")
	}
	if err := NotifyStaff(stored); err != nil {
		facades.Log().Warning("contact: could not queue the notification: " + err.Error())
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/contact?status=sent")
}

// NotifyStaff queues a plain-text mail about a new message to CONTACT_NOTIFICATION_EMAIL (nothing when unset).
// Tests replace it.
var NotifyStaff = func(message models.Contact) error {
	to, _ := facades.Config().Env("CONTACT_NOTIFICATION_EMAIL", "").(string)
	if to == "" {
		return nil
	}
	subject := message.SubjectText()
	if subject == "" {
		subject = "from " + message.Name
	}
	body := "Name: " + message.Name + "\nEmail: " + message.Email + "\nSubject: " + message.SubjectText() +
		"\n\n" + message.Message + "\n\nInbox: " + facades.Config().GetString("http.url") + "/admin/contacts\n"
	return facades.Mail().To([]string{to}).Subject("New contact message " + subject).
		Content(contractsmail.Content{Text: body}).Queue()
}

func optional(text string, limit int) *string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if runes := []rune(text); len(runes) > limit {
		text = string(runes[:limit])
	}
	return &text
}
