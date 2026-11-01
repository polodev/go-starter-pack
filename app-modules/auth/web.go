package auth

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/web"
)

// Feedback shown after a redirect: ?status=<key> maps to the text auth.status-<key>. Only these keys are allowed,
// so a link cannot put free text on the page.
var statusKeys = map[string]bool{
	"registered": true, "verified": true, "verification-sent": true, "reset-link-sent": true,
	"password-reset": true, "logged-out": true, "link-invalid": true,
}

func status(ctx http.Context) string {
	if key := ctx.Request().Query("status"); statusKeys[key] {
		return web.T(ctx, "auth.status-"+key)
	}
	return ""
}

// form renders an auth page; with errors it answers 422 and keeps what was typed.
func form(ctx http.Context, view string, values map[string]any, errors map[string]string) http.Response {
	data := map[string]any{"status": status(ctx), "errors": errors, "title": web.T(ctx, "auth.title-"+strings.TrimPrefix(view, "auth::"))}
	for key, value := range values {
		data[key] = value
	}
	if len(errors) > 0 {
		return web.ViewStatus(ctx, http.StatusUnprocessableEntity, view, data)
	}
	return web.View(ctx, view, data)
}

// fieldErrors turns the validator's errors into one translated message per field.
func fieldErrors(ctx http.Context, all map[string]map[string]string) map[string]string {
	errors := map[string]string{}
	for field := range all {
		errors[field] = web.T(ctx, "auth.error-"+field)
	}
	return errors
}

// ShowLogin is GET /login.
func ShowLogin(ctx http.Context) http.Response {
	return form(ctx, "auth::login", map[string]any{"next": web.SafeNext(ctx.Request().Query("next"), "")}, nil)
}

// Login is POST /login. Throttled (5 a minute per email and address); the same message for every failure.
func Login(ctx http.Context) http.Response {
	email, next := ctx.Request().Input("email"), web.SafeNext(ctx.Request().Input("next"), "/account/profile")
	user, ok := CheckCredentials(email, ctx.Request().Input("password"))
	if !ok {
		return form(ctx, "auth::login", map[string]any{"email": email, "next": next},
			map[string]string{"email": web.T(ctx, "auth.error-credentials")})
	}
	if _, err := facades.Auth(ctx).Guard("web").Login(&user); err != nil {
		facades.Log().Request(ctx.Request()).Error("login failed: " + err.Error())
		return form(ctx, "auth::login", map[string]any{"email": email}, map[string]string{"email": web.T(ctx, "auth.error-server")})
	}
	return ctx.Response().Redirect(http.StatusSeeOther, next)
}

// Logout is POST /logout: the session is ended and renewed.
func Logout(ctx http.Context) http.Response {
	_ = facades.Auth(ctx).Guard("web").Logout()
	return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=logged-out")
}

// ShowRegister is GET /register.
func ShowRegister(ctx http.Context) http.Response {
	return form(ctx, "auth::register", nil, nil)
}

// Register is POST /register: creates the account, signs it in, and emails a verification link.
func Register(ctx http.Context) http.Response {
	name, email := ctx.Request().Input("name"), ctx.Request().Input("email")
	values := map[string]any{"name": name, "email": email}
	validator, err := ctx.Request().Validate(registerRules())
	if err != nil {
		return form(ctx, "auth::register", values, map[string]string{"email": web.T(ctx, "auth.error-server")})
	}
	errors := map[string]string{}
	if validator.Fails() {
		errors = fieldErrors(ctx, validator.Errors().All())
	}
	if ctx.Request().Input("password") != ctx.Request().Input("password_confirmation") {
		errors["password_confirmation"] = web.T(ctx, "auth.error-password_confirmation")
	}
	if len(errors) > 0 {
		return form(ctx, "auth::register", values, errors)
	}
	user, taken, err := CreateUser(name, email, ctx.Request().Input("password"))
	if taken {
		return form(ctx, "auth::register", values, map[string]string{"email": web.T(ctx, "auth.error-email-taken")})
	}
	if err != nil {
		facades.Log().Request(ctx.Request()).Error("register failed: " + err.Error())
		return form(ctx, "auth::register", values, map[string]string{"email": web.T(ctx, "auth.error-server")})
	}
	if _, err := facades.Auth(ctx).Guard("web").Login(&user); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/login")
	}
	sendVerification(ctx, &user)
	return ctx.Response().Redirect(http.StatusSeeOther, "/email/verify?status=registered")
}

func sendVerification(ctx http.Context, user *models.User) {
	secret, err := IssueToken(user.ID, PurposeVerifyEmail)
	if err == nil {
		err = sendLink(web.CurrentLocale(ctx), user.Email, PurposeVerifyEmail, linkURL("/email/verify/"+secret))
	}
	if err != nil {
		facades.Log().Request(ctx.Request()).Error("verification mail not sent: " + err.Error())
	}
}

// VerifyNotice is GET /email/verify: "check your inbox", with a button to send the link again.
func VerifyNotice(ctx http.Context) http.Response {
	user := CurrentUser(ctx)
	if user.EmailVerifiedAt != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/account/profile")
	}
	return form(ctx, "auth::verify-notice", map[string]any{"email": user.Email}, nil)
}

// ResendVerification is POST /email/verification-notification (throttled).
func ResendVerification(ctx http.Context) http.Response {
	user := CurrentUser(ctx)
	if user.EmailVerifiedAt == nil {
		sendVerification(ctx, user)
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/email/verify?status=verification-sent")
}

// VerifyEmail is GET /email/verify/{token}: a one-time link from the email.
func VerifyEmail(ctx http.Context) http.Response {
	userID, ok := ConsumeToken(ctx.Request().Route("token"), PurposeVerifyEmail)
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
	return ctx.Response().Redirect(http.StatusSeeOther, "/account/profile?status=verified")
}

// ShowForgotPassword is GET /forgot-password.
func ShowForgotPassword(ctx http.Context) http.Response {
	return form(ctx, "auth::forgot-password", nil, nil)
}

// ForgotPassword is POST /forgot-password (throttled). It answers the same whether or not the address has an
// account, so it never reveals who is registered.
func ForgotPassword(ctx http.Context) http.Response {
	email := NormaliseEmail(ctx.Request().Input("email"))
	var user models.User
	if facades.Orm().Query().Where("email", email).First(&user) == nil && user.ID != 0 {
		secret, err := IssueToken(user.ID, PurposeResetPassword)
		if err == nil {
			err = sendLink(web.CurrentLocale(ctx), user.Email, PurposeResetPassword, linkURL("/reset-password/"+secret))
		}
		if err != nil {
			facades.Log().Request(ctx.Request()).Error("reset mail not sent: " + err.Error())
		}
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/forgot-password?status=reset-link-sent")
}

// ShowResetPassword is GET /reset-password/{token}.
func ShowResetPassword(ctx http.Context) http.Response {
	token := ctx.Request().Route("token")
	if _, ok := PeekToken(token, PurposeResetPassword); !ok {
		return ctx.Response().Redirect(http.StatusSeeOther, "/forgot-password?status=link-invalid")
	}
	return form(ctx, "auth::reset-password", map[string]any{"token": token}, nil)
}

// ResetPassword is POST /reset-password/{token}: a new password, then the reader signs in with it.
func ResetPassword(ctx http.Context) http.Response {
	token := ctx.Request().Route("token")
	password := ctx.Request().Input("password")
	if len(password) < 8 || password != ctx.Request().Input("password_confirmation") {
		key := "auth.error-password"
		if len(password) >= 8 {
			key = "auth.error-password_confirmation"
		}
		return form(ctx, "auth::reset-password", map[string]any{"token": token}, map[string]string{"password": web.T(ctx, key)})
	}
	userID, ok := ConsumeToken(token, PurposeResetPassword)
	if !ok {
		return ctx.Response().Redirect(http.StatusSeeOther, "/forgot-password?status=link-invalid")
	}
	var user models.User
	if err := facades.Orm().Query().Find(&user, userID); err != nil || user.ID == 0 {
		return ctx.Response().Redirect(http.StatusSeeOther, "/forgot-password?status=link-invalid")
	}
	if err := SetPassword(&user, password); err != nil {
		return form(ctx, "auth::reset-password", map[string]any{"token": token}, map[string]string{"password": web.T(ctx, "auth.error-server")})
	}
	if user.EmailVerifiedAt == nil {
		_ = MarkVerified(&user) // the reader proved they own the address
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/login?status=password-reset")
}

// Profile is GET /account/profile: the signed-in person's details and whether the email is verified.
func Profile(ctx http.Context) http.Response {
	return web.View(ctx, "auth::profile", map[string]any{
		"title": web.T(ctx, "auth.title-profile"), "status": status(ctx), "account": CurrentUser(ctx),
	})
}
