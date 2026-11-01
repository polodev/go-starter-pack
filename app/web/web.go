// Package web is the "web" middleware group and the helpers every page uses: sessions, CSRF, the visitor's
// language, and the data a layout needs (Laravel's app/Http/Kernel web group plus view composers).
package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/render"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/contracts/translation"
	"github.com/goravel/framework/http/middleware"
	sessionmiddleware "github.com/goravel/framework/session/middleware"

	"github.com/goravel/gin"

	"goravel/app/facades"
)

// LocaleCookie remembers the visitor's language.
const LocaleCookie = "lang"

// Locales are the languages the site speaks; the first is the default (config app.locale and fallback_locale).
var Locales = []string{"en", "bn"}

// Supported reports whether a locale is one of Locales.
func Supported(locale string) bool {
	for _, known := range Locales {
		if known == locale {
			return true
		}
	}
	return false
}

// Group registers page routes behind sessions, CSRF protection and the language middleware.
func Group(register func(router route.Router)) {
	facades.Route().Middleware(sessionmiddleware.StartSession(), middleware.VerifyCsrfToken(), &Locale{}).Group(register)
}

// Locale picks the language: ?lang=, then the cookie, then Accept-Language, then English.
type Locale struct{}

func (r *Locale) Signature() string { return "locale" }

func (r *Locale) Handle(ctx http.Context) {
	locale := ""
	if query := ctx.Request().Query("lang"); Supported(query) {
		locale = query
	} else if cookie := ctx.Request().Cookie(LocaleCookie); Supported(cookie) {
		locale = cookie
	} else {
		for _, part := range strings.Split(ctx.Request().Header("Accept-Language"), ",") {
			code := strings.ToLower(strings.TrimSpace(strings.SplitN(strings.SplitN(part, ";", 2)[0], "-", 2)[0]))
			if Supported(code) {
				locale = code
				break
			}
		}
	}
	if locale == "" {
		locale = Locales[0]
	}
	ctx.WithValue("locale", locale)
	facades.Lang(ctx).SetLocale(locale)
	ctx.Request().Next()
}

// CurrentLocale is the language chosen by the Locale middleware.
func CurrentLocale(ctx http.Context) string {
	if locale, ok := ctx.Value("locale").(string); ok && locale != "" {
		return locale
	}
	return Locales[0]
}

// T translates for the current request with Goravel's translator: T(ctx, "auth.login-title"),
// T(ctx, "auth.welcome", "name", user.Name) replaces :name.
func T(ctx http.Context, key string, replacements ...string) string {
	return translate(CurrentLocale(ctx), key, replacements...)
}

// Translate is T for a given language, outside a request (an email in the visitor's language).
func Translate(locale, key string, replacements ...string) string {
	return translate(locale, key, replacements...)
}

func translate(locale, key string, replacements ...string) string {
	translator := facades.Lang(context.Background())
	translator.SetLocale(locale)
	if len(replacements) < 2 {
		return translator.Get(key)
	}
	values := map[string]string{}
	for index := 0; index+1 < len(replacements); index += 2 {
		values[replacements[index]] = replacements[index+1]
	}
	return translator.Get(key, translation.Option{Replace: values})
}

// Data is what every page template receives, plus the page's own values: the language, the CSRF token, the
// current path and the signed-in person (added by the auth module through Composers).
func Data(ctx http.Context, values map[string]any) map[string]any {
	data := map[string]any{
		"lang":    CurrentLocale(ctx),
		"locales": Locales,
		"csrf":    ctx.Request().Session().Token(),
		"path":    ctx.Request().Path(),
		"appName": facades.Config().GetString("app.name"),
	}
	for _, composer := range composers {
		composer(ctx, data)
	}
	for key, value := range values {
		data[key] = value
	}
	return data
}

var composers []func(ctx http.Context, data map[string]any)

// Compose adds values to every page's data (Laravel's View::composer('*', ...)); modules call it at boot.
func Compose(composer func(ctx http.Context, data map[string]any)) {
	composers = append(composers, composer)
}

// View renders a template with Data.
func View(ctx http.Context, name string, values map[string]any) http.Response {
	return ctx.Response().View().Make(name, Data(ctx, values))
}

// Templates is the HTML renderer for every page. It compiles on first use, because module view folders are
// registered while the app boots, after the router has been set up.
type Templates struct {
	once     sync.Once
	compiled *render.HTMLProduction
	err      error
}

var templates = &Templates{}

// Renderer is what config/http.go hands to Gin.
func Renderer() render.HTMLRender { return templates }

func (r *Templates) load() (*render.HTMLProduction, error) {
	r.once.Do(func() {
		r.compiled, r.err = gin.NewTemplate(gin.RenderOptions{FuncMap: FuncMap()})
		if r.err == nil && r.compiled == nil {
			r.err = errors.New("no templates found in resources/views or the modules")
		}
	})
	return r.compiled, r.err
}

// Instance is Gin's render.HTMLRender method.
func (r *Templates) Instance(name string, data any) render.Render {
	compiled, err := r.load()
	if err != nil {
		panic(err)
	}
	return compiled.Instance(name, data)
}

// ViewStatus renders a template with Data and a status code other than 200 (an error page, a 422 form).
// View().Make always answers 200, so this renders the same templates into the response itself.
func ViewStatus(ctx http.Context, code int, name string, values map[string]any) http.Response {
	compiled, err := templates.load()
	if err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).String("templates: %v", err)
	}
	var page bytes.Buffer
	if err := compiled.Template.ExecuteTemplate(&page, name, Data(ctx, values)); err != nil {
		return ctx.Response().Status(http.StatusInternalServerError).String("template %s: %v", name, err)
	}
	return ctx.Response().Status(code).Data("text/html; charset=utf-8", page.Bytes())
}

// Render fills a template outside a request (an email body). data should carry "lang" for {{ t .lang "..." }}.
func Render(name string, data map[string]any) (string, error) {
	compiled, err := templates.load()
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := compiled.Template.ExecuteTemplate(&output, name, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

// SafeNext keeps a redirect target on this site: a path that starts with one "/", else fallback.
func SafeNext(next, fallback string) string {
	parsed, err := url.Parse(next)
	if err != nil || next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || parsed.Host != "" {
		return fallback
	}
	return next
}

// FuncMap is available in every template: {{ t .lang "home.title" }}, {{ tr .lang "x" "name" .Name }}.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"t":          func(locale, key string) string { return translate(locale, key) },
		"tr":         func(locale, key string, pairs ...string) string { return translate(locale, key, pairs...) },
		"eq_str":     func(a, b any) bool { return fmt.Sprint(a) == fmt.Sprint(b) },
		"has_prefix": strings.HasPrefix,
		"list":       func(items ...string) []string { return items },
		"split":      strings.Split,
		"upper":      strings.ToUpper,
	}
}
