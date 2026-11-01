package auth

import (
	"net/url"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
)

// CurrentUser is the person signed in on the web (guard "web"), or nil.
func CurrentUser(ctx http.Context) *models.User {
	if cached, ok := ctx.Value("auth_user").(*models.User); ok {
		return cached
	}
	var user models.User
	if err := facades.Auth(ctx).Guard("web").User(&user); err != nil || user.ID == 0 {
		return nil
	}
	ctx.WithValue("auth_user", &user)
	return &user
}

// Authenticate is Laravel's `auth`: a guest is sent to /login?next=<this page>.
type Authenticate struct{}

func (r *Authenticate) Signature() string { return "auth" }

func (r *Authenticate) Handle(ctx http.Context) {
	if CurrentUser(ctx) == nil {
		_ = ctx.Response().Redirect(http.StatusSeeOther, "/login?next="+url.QueryEscape(ctx.Request().FullUrl())).Abort()
		return
	}
	ctx.Request().Next()
}

// Verified is Laravel's `verified`: a person whose email is not verified yet is sent to /email/verify.
type Verified struct{}

func (r *Verified) Signature() string { return "verified" }

func (r *Verified) Handle(ctx http.Context) {
	user := CurrentUser(ctx)
	if user == nil {
		_ = ctx.Response().Redirect(http.StatusSeeOther, "/login").Abort()
		return
	}
	if user.EmailVerifiedAt == nil {
		_ = ctx.Response().Redirect(http.StatusSeeOther, "/email/verify").Abort()
		return
	}
	ctx.Request().Next()
}

// Guest is Laravel's `guest`: a signed-in person does not see the login and register pages.
type Guest struct{}

func (r *Guest) Signature() string { return "guest" }

func (r *Guest) Handle(ctx http.Context) {
	if CurrentUser(ctx) != nil {
		_ = ctx.Response().Redirect(http.StatusSeeOther, "/account/profile").Abort()
		return
	}
	ctx.Request().Next()
}
