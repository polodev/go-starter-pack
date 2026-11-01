// Package media is a Spatie laravel-medialibrary compatible media library: files belong to an owner (type and
// id) in named collections, single or multiple, with accepted types and a size limit; public files are served
// from their public address, private ones only to the owner and staff with media.view. Files live on this
// server or in Cloudflare R2 (MEDIA_STORAGE=r2: public and private buckets, never mixed).
package media

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/route"

	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/auth"
	"goravel/app-modules/media/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/modular"
	"goravel/app/web"
)

// Module registers the table, the collections, file serving, the avatar page and the admin screen.
type Module struct{}

func (m *Module) Name() string      { return "media" }
func (m *Module) ViewsPath() string { return "app-modules/media/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261101170000CreateMediaTable{}}
}

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "Content", Label: "Media", Href: "/admin/media", Permission: "media.view", Icon: "▣", Order: 2}}
}

func (m *Module) Routes() {
	// A person's avatar: one public image, up to 2 MB. The shared library: any file, private or public.
	Define("user", Single("avatar").Images().AsPublic().Max(2<<20))
	Define("library", Multiple("files"), Multiple("public-files").AsPublic())

	web.Compose(func(ctx http.Context, data map[string]any) {
		if user := auth.CurrentUser(ctx); user != nil {
			if avatar := First(UserOwner(user.ID), "avatar"); avatar != nil {
				data["avatarURL"] = URL(*avatar)
			}
		}
	})

	web.Group(func(router route.Router) {
		router.Get("/media/{id}/{file_name}", Serve)
		signedIn := router.Middleware(&auth.Authenticate{})
		signedIn.Get("/account/avatar", AvatarPage)
		signedIn.Post("/account/avatar", UploadAvatar)
		signedIn.Post("/account/avatar/delete", RemoveAvatar)
	})
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/media", "media.view", AdminIndex)
		admin.Post("/media", "media.manage", AdminUpload)
		admin.Post("/media/{id}/delete", "media.manage", AdminDelete)
	})
}

func adminView(ctx http.Context, name, title string, values map[string]any) http.Response {
	return adminlayout.View(ctx, name, title, []adminlayout.Crumb{{Label: title}}, values)
}
