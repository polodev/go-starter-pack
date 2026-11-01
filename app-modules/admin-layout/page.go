package adminlayout

import (
	"strconv"

	"github.com/goravel/framework/contracts/http"

	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/modular"
	"goravel/app/web"
)

// Crumb is one breadcrumb step; the last one has no link.
type Crumb struct {
	Label string
	Href  string
}

// menuLink is a sidebar entry with its section and whether it is the current page.
type menuLink struct {
	modular.MenuItem
	Active bool
}

type menuSection struct {
	Name  string
	Links []menuLink
}

// Data is what every admin page gets: the shell (sidebar filtered by permission, breadcrumb, title) plus values.
func Data(ctx http.Context, title string, crumbs []Crumb, values map[string]any) map[string]any {
	can := rolepermission.Can(ctx)
	path := ctx.Request().Path()
	var sections []menuSection
	for _, item := range modular.Menu(modular.Registered()) {
		if !can(item.Permission) {
			continue
		}
		active := path == item.Href || (item.Href != "/admin" && len(path) > len(item.Href) && path[:len(item.Href)+1] == item.Href+"/")
		if len(sections) == 0 || sections[len(sections)-1].Name != item.Section {
			sections = append(sections, menuSection{Name: item.Section})
		}
		last := &sections[len(sections)-1]
		last.Links = append(last.Links, menuLink{MenuItem: item, Active: active})
	}
	data := map[string]any{
		"title":  title,
		"crumbs": append([]Crumb{{Label: "Dashboard", Href: "/admin"}}, crumbs...),
		"menu":   sections,
		"saved":  ctx.Request().Query("saved"),
		"error":  ctx.Request().Query("error"),
	}
	for key, value := range values {
		data[key] = value
	}
	return data
}

// View renders an admin page.
func View(ctx http.Context, name, title string, crumbs []Crumb, values map[string]any) http.Response {
	return web.View(ctx, name, Data(ctx, title, crumbs, values))
}

// ViewStatus renders an admin page with a status (422 for a form with errors).
func ViewStatus(ctx http.Context, code int, name, title string, crumbs []Crumb, values map[string]any) http.Response {
	return web.ViewStatus(ctx, code, name, Data(ctx, title, crumbs, values))
}

// NotFound is the admin 404.
func NotFound(ctx http.Context) http.Response {
	return web.ViewStatus(ctx, http.StatusNotFound, "site::error", map[string]any{
		"code": http.StatusNotFound, "message": web.T(ctx, "site-layout.not-found"), "title": web.T(ctx, "site-layout.not-found"),
	})
}

// IDParam reads a numeric {id}; 0 when it is not a number.
func IDParam(ctx http.Context) uint {
	id, _ := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	return uint(id)
}

// PageParam is ?page=, at least 1.
func PageParam(ctx http.Context) int {
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	if page < 1 {
		return 1
	}
	return page
}
