// Package sitelayout is the public layout: header (navigation, language and theme switch) and footer, the home
// page, the language switch route and the error page. Pages use {{ template "site::header" . }} and
// {{ template "site::footer" . }} around their own content.
package sitelayout

import (
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	"goravel/app/web"
)

// Module registers the home page, the language switch, views and translations.
type Module struct{}

func (m *Module) Name() string      { return "site-layout" }
func (m *Module) ViewsPath() string { return "app-modules/site-layout/resources/views" }

func (m *Module) Routes() {
	web.Group(func(router route.Router) {
		router.Get("/", Home)
		router.Get("/language/{locale}", SwitchLanguage)
	})
}

// Home is the landing page.
func Home(ctx http.Context) http.Response {
	return web.View(ctx, "site::home", map[string]any{"title": web.T(ctx, "site-layout.home-title")})
}

// SwitchLanguage remembers the language for a year and goes back to ?next= (a path on this site only).
func SwitchLanguage(ctx http.Context) http.Response {
	locale := ctx.Request().Route("locale")
	if !web.Supported(locale) {
		locale = web.Locales[0]
	}
	ctx.Response().Cookie(http.Cookie{
		Name: web.LocaleCookie, Value: locale, Path: "/", MaxAge: int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true, SameSite: "lax",
	})
	return ctx.Response().Redirect(http.StatusSeeOther, web.SafeNext(ctx.Request().Query("next"), "/"))
}

// Error renders the error page with a status code (404, 403 ...).
func Error(ctx http.Context, code int, messageKey string) http.Response {
	return web.ViewStatus(ctx, code, "site::error", map[string]any{
		"code": code, "message": web.T(ctx, messageKey), "title": web.T(ctx, messageKey),
	})
}
