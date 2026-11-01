// Package auth is signing in: web pages with sessions (guard "web": login, register, logout, email verification,
// password reset, profile) and the token API under /api/v1/auth (guard "api"). Goravel provides the guards,
// hashing, validation, rate limiting and mail; this module adds the pages and one-time email links.
package auth

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/http/middleware"

	"goravel/app-modules/auth/migrations"
	"goravel/app/facades"
	"goravel/app/web"
)

// Module registers the auth pages, the API, the links table and its cleaning.
type Module struct{}

func (m *Module) Name() string      { return "auth" }
func (m *Module) ViewsPath() string { return "app-modules/auth/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261101140000CreateUserSecurityTokensTable{}}
}

func (m *Module) Commands() []console.Command { return []console.Command{&PruneTokensCommand{}} }

func (m *Module) Schedule() []schedule.Event {
	return []schedule.Event{facades.Schedule().Command("auth:prune-tokens").Daily()}
}

func (m *Module) Routes() {
	// Login: 5 tries a minute per email and address; email links: 3 an hour per email and address.
	facades.RateLimiter().For("login", func(ctx http.Context) http.Limit {
		return limit.PerMinute(5).By(NormaliseEmail(ctx.Request().Input("email")) + "|" + ctx.Request().Ip()).
			Response(func(ctx http.Context) {
				page := form(ctx, "auth::login", map[string]any{"email": ctx.Request().Input("email")},
					map[string]string{"email": web.T(ctx, "auth.error-throttled")})
				if abortable, ok := page.(http.AbortableResponse); ok {
					_ = abortable.Abort()
				}
			})
	})
	facades.RateLimiter().For("email-links", func(ctx http.Context) http.Limit {
		key := NormaliseEmail(ctx.Request().Input("email"))
		if user := CurrentUser(ctx); user != nil {
			key = user.Email
		}
		return limit.PerHour(3).By(key + "|" + ctx.Request().Ip())
	})

	web.Compose(func(ctx http.Context, data map[string]any) {
		if user := CurrentUser(ctx); user != nil {
			data["user"] = user
		}
	})

	web.Group(func(router route.Router) {
		guest := router.Middleware(&Guest{})
		guest.Get("/login", ShowLogin)
		guest.Middleware(middleware.Throttle("login")).Post("/login", Login)
		guest.Get("/register", ShowRegister)
		guest.Post("/register", Register)
		guest.Get("/forgot-password", ShowForgotPassword)
		guest.Middleware(middleware.Throttle("email-links")).Post("/forgot-password", ForgotPassword)
		guest.Get("/login/email-link", ShowEmailLinkForm)
		guest.Middleware(middleware.Throttle("email-links")).Post("/login/email-link", SendEmailLink)
		router.Get("/login/email-link/{token}", ShowEmailLinkConfirm)
		router.Post("/login/email-link/{token}", UseEmailLink)
		router.Get("/reset-password/{token}", ShowResetPassword)
		router.Post("/reset-password/{token}", ResetPassword)
		router.Get("/email/verify/{token}", VerifyEmail)

		signedIn := router.Middleware(&Authenticate{})
		signedIn.Post("/logout", Logout)
		signedIn.Get("/email/verify", VerifyNotice)
		signedIn.Middleware(middleware.Throttle("email-links")).Post("/email/verification-notification", ResendVerification)
		signedIn.Get("/account/profile", Profile)
	})

	facades.Route().Prefix("/api/v1").Group(func(router route.Router) {
		router.Post("/auth/register", ApiRegister)
		router.Post("/auth/login", ApiLogin)
		router.Middleware(&ApiAuthenticate{}).Get("/auth/me", ApiMe)
	})
}
