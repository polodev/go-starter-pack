// Package pages is bilingual website pages: English is required, every other language optional, and each empty
// part (title, description, text) falls back to English on its own. Staff edit them in the admin; their HTML is
// cleaned with bluemonday before it is stored.
package pages

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/contracts/route"

	"goravel/app-modules/pages/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/modular"
	"goravel/app/web"
)

// Module registers the tables, the public page route, the admin screens and the seeder.
type Module struct{}

func (m *Module) Name() string      { return "pages" }
func (m *Module) ViewsPath() string { return "app-modules/pages/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261101190000CreatePagesTables{}}
}

func (m *Module) Seeders() []seeder.Seeder { return []seeder.Seeder{&PagesSeeder{}} }

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "Content", Label: "Pages", Href: "/admin/pages", Permission: "pages.view", Icon: "❏", Order: 1}}
}

func (m *Module) Routes() {
	web.Group(func(router route.Router) {
		router.Get("/pages/{slug}", Show)
	})
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/pages", "pages.view", AdminIndex)
		admin.Get("/pages/create", "pages.manage", AdminCreate)
		admin.Post("/pages", "pages.manage", AdminStore)
		admin.Get("/pages/{id}", "pages.view", AdminShow)
		admin.Get("/pages/{id}/edit", "pages.manage", AdminEdit)
		admin.Post("/pages/{id}", "pages.manage", AdminUpdate)
	})
}
