package pages

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/web"
)

// Show is GET /pages/{slug}: a published page in the visitor's language.
func Show(ctx http.Context) http.Response {
	page, ok := FindBySlug(ctx.Request().Route("slug"))
	if !ok || !page.IsPublished {
		return web.ViewStatus(ctx, http.StatusNotFound, "site::error", map[string]any{
			"code": http.StatusNotFound, "message": web.T(ctx, "site-layout.not-found"), "title": web.T(ctx, "site-layout.not-found"),
		})
	}
	content := Load(page, web.CurrentLocale(ctx))
	return web.View(ctx, "pages::show", map[string]any{"title": content.Title, "content": content})
}
