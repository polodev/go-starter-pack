// Package contact is the public contact form (/contact) and the admin inbox (/admin/contacts).
package contact

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/http/limit"
	"github.com/goravel/framework/http/middleware"

	"goravel/app-modules/contact/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/modular"
	"goravel/app/web"
)

// Module registers the contacts table, the form, its rate limit and the inbox.
type Module struct{}

func (m *Module) Name() string      { return "contact" }
func (m *Module) ViewsPath() string { return "app-modules/contact/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261102120000CreateContactsTable{}}
}

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "Content", Label: "Contact messages", Href: "/admin/contacts", Permission: "contacts.view", Icon: "✉", Order: 3}}
}

func (m *Module) Routes() {
	// A person sends a few messages; a script sends dozens: 5 posts per 10 minutes per address.
	facades.RateLimiter().For("contact", func(ctx http.Context) http.Limit {
		return limit.PerMinutes(10, 5).By(ctx.Request().Ip()).Response(func(ctx http.Context) {
			response := page(ctx, http.StatusTooManyRequests, readMessage(ctx), nil, "contact.error-throttled")
			if abortable, ok := response.(http.AbortableResponse); ok {
				_ = abortable.Abort()
			}
		})
	})

	web.Group(func(router route.Router) {
		router.Get("/contact", ShowForm)
		router.Middleware(middleware.Throttle("contact")).Post("/contact", Send)
	})

	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/contacts", "contacts.view", AdminIndex)
		admin.Get("/contacts/{id}", "contacts.view", AdminShow)
		admin.Post("/contacts/{id}/status", "contacts.manage", AdminStatus)
		admin.Post("/contacts/{id}/delete", "contacts.manage", AdminDelete)
	})
}
