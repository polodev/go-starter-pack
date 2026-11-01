// Package socialauth is sign in with Google, for the website and the API (the Laravel Socialite idea). Google
// accounts are linked to users in social_accounts. Settings: config/services.go (services.google).
package socialauth

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/http/middleware"

	"goravel/app-modules/auth"
	"goravel/app-modules/social-auth/migrations"
	"goravel/app/facades"
	"goravel/app/web"
)

// Module registers the table, the routes, the login button and the login page's messages.
type Module struct{}

func (m *Module) Name() string { return "social-auth" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261102130000CreateSocialAccountsTable{}}
}

func (m *Module) Routes() {
	for _, key := range []string{"google-unavailable", "google-failed", "google-email-unverified", "google-no-account"} {
		auth.AddLoginStatus(key, "social-auth.status-"+key)
	}
	web.Compose(func(ctx http.Context, data map[string]any) {
		data["googleSignIn"] = Enabled()
	})

	// Each call asks Google, so the API is throttled like a login.
	facades.RateLimiter().For("google-api", func(ctx http.Context) http.Limit {
		return limit.PerMinute(10).By(ctx.Request().Ip())
	})

	web.Group(func(router route.Router) {
		router.Get("/auth/google/redirect", Redirect)
		router.Get("/auth/google/callback", Callback)
	})
	facades.Route().Prefix("/api/v1/auth").Middleware(middleware.Throttle("google-api")).Post("/google", ApiGoogle)
}
