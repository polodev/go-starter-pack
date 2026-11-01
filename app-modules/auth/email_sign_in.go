package auth

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/web"
)

// Sign-in links by email (passwordless login, Laravel's "magic link" packages): the person asks for a link, the
// mail holds a one-time link valid for 15 minutes, and the link opens a page with a button. The button, not the
// link itself, signs in: mail scanners open links ahead of the person and would otherwise use the link up.

// ShowEmailLinkForm is GET /login/email-link.
func ShowEmailLinkForm(ctx http.Context) http.Response {
	return form(ctx, "auth::email-link", nil, nil)
}

// SendEmailLink is POST /login/email-link (throttled like the other email links). It answers the same whether or
// not the address has an account, so it never reveals who is registered.
func SendEmailLink(ctx http.Context) http.Response {
	email := NormaliseEmail(ctx.Request().Input("email"))
	var user models.User
	if email != "" && facades.Orm().Query().Where("email", email).First(&user) == nil && user.ID != 0 {
		secret, err := IssueToken(user.ID, PurposeSignIn)
		if err == nil {
			err = sendLink(web.CurrentLocale(ctx), user.Email, PurposeSignIn, linkURL("/login/email-link/"+secret))
		}
		if err != nil {
			facades.Log().Request(ctx.Request()).Error("sign-in link not sent: " + err.Error())
		}
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/login/email-link?status=sign-in-link-sent")
}

// ShowEmailLinkConfirm is GET /login/email-link/{token}: a page with a button; the link is not used yet.
func ShowEmailLinkConfirm(ctx http.Context) http.Response {
	token := ctx.Request().Route("token")
	if _, ok := PeekToken(token, PurposeSignIn); !ok {
		return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=link-invalid")
	}
	return form(ctx, "auth::email-link-confirm", map[string]any{"token": token}, nil)
}

// UseEmailLink is POST /login/email-link/{token} (CSRF-checked): uses the link once and signs the person in. The
// address counts as verified, since the person read the mail.
func UseEmailLink(ctx http.Context) http.Response {
	userID, ok := ConsumeToken(ctx.Request().Route("token"), PurposeSignIn)
	if !ok {
		return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=link-invalid")
	}
	var user models.User
	if err := facades.Orm().Query().Find(&user, userID); err != nil || user.ID == 0 {
		return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=link-invalid")
	}
	if user.EmailVerifiedAt == nil {
		_ = MarkVerified(&user)
	}
	if _, err := facades.Auth(ctx).Guard("web").Login(&user); err != nil {
		facades.Log().Request(ctx.Request()).Error("sign-in by link failed: " + err.Error())
		return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=link-invalid")
	}
	activity.In("auth").CausedBy(user.ID).PerformedOn("user", user.ID).Event("login_by_email_link").
		LogQuietly("logged in with an email link")
	return ctx.Response().Redirect(http.StatusSeeOther, "/account/profile")
}
