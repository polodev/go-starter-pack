// Package adminlayout is the admin area: the shell (sidebar built from every module's menu and shown by
// permission, breadcrumb), the dashboard, the user screens, and the roles and permissions panel.
package adminlayout

import (
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/modular"
)

// Module registers the admin routes (each declared with its permission) and the sidebar links.
type Module struct{}

func (m *Module) Name() string      { return "admin-layout" }
func (m *Module) ViewsPath() string { return "app-modules/admin-layout/resources/views" }

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{
		{Section: "Overview", Label: "Dashboard", Href: "/admin", Permission: "dashboard.view", Icon: "▦", Order: 1},
		{Section: "People and access", Label: "Users", Href: "/admin/users", Permission: "users.view", Icon: "☺", Order: 1},
		{Section: "People and access", Label: "Roles & permissions", Href: "/admin/role-permissions", Permission: "roles.view", Icon: "⚿", Order: 2},
	}
}

func (m *Module) Routes() {
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/", "dashboard.view", Dashboard)
		admin.Get("/users", "users.view", UsersIndex)
		admin.Get("/users/{id}", "users.view", UserShow)
		admin.Get("/users/{id}/edit", "users.manage", UserEdit)
		admin.Post("/users/{id}", "users.manage", UserUpdate)

		admin.Get("/role-permissions", "roles.view", RolesOverview)
		admin.Get("/role-permissions/roles", "roles.view", RolesIndex)
		admin.Get("/role-permissions/roles/{slug}", "roles.view", RoleShow)
		admin.Get("/role-permissions/permissions", "roles.view", PermissionsIndex)
		admin.Get("/role-permissions/matrix", "roles.view", Matrix)
		admin.Get("/role-permissions/users", "roles.view", Assignments)
		admin.Get("/role-permissions/links", "roles.view", BackendLinks)
		admin.Post("/role-permissions/sync", "roles.sync", SyncRoles)
	})
}
