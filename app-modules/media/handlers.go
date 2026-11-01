package media

import (
	"errors"
	"strconv"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app-modules/auth"
	"goravel/app-modules/media/models"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/web"
)

func notFound(ctx http.Context) http.Response {
	return web.ViewStatus(ctx, http.StatusNotFound, "site::error", map[string]any{
		"code": http.StatusNotFound, "message": web.T(ctx, "site-layout.not-found"), "title": web.T(ctx, "site-layout.not-found"),
	})
}

// mayRead: public files for everyone; private ones for the owner and staff with media.view. Unknown and
// forbidden files both answer 404, so nobody learns which private files exist.
func mayRead(ctx http.Context, item models.Media) bool {
	if item.IsPublic() {
		return true
	}
	user := auth.CurrentUser(ctx)
	if user == nil {
		return false
	}
	if item.ModelType == "user" && item.ModelID == strconv.FormatUint(uint64(user.ID), 10) {
		return true
	}
	return rolepermission.HasPermission(user.ID, "media.view")
}

// Serve is GET /media/{id}/{file_name}.
func Serve(ctx http.Context) http.Response {
	id, _ := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	var item models.Media
	if id == 0 || facades.Orm().Query().Find(&item, id) != nil || item.ID == 0 ||
		item.FileName != ctx.Request().Route("file_name") || !mayRead(ctx, item) {
		return notFound(ctx)
	}
	if item.Disk == "r2_public" {
		return ctx.Response().Redirect(http.StatusFound, URL(item))
	}
	content, err := facades.Storage().Disk(item.Disk).GetBytes(Path(item))
	if err != nil {
		return notFound(ctx)
	}
	cache := "private, max-age=0, no-store"
	if item.IsPublic() {
		cache = "public, max-age=86400"
	}
	ctx.Response().Header("Cache-Control", cache)
	ctx.Response().Header("X-Content-Type-Options", "nosniff")
	ctx.Response().Header("Content-Disposition", `inline; filename="`+item.FileName+`"`)
	return ctx.Response().Data(http.StatusOK, item.Mime(), content)
}

// AvatarPage is GET /account/avatar.
func AvatarPage(ctx http.Context) http.Response {
	user := auth.CurrentUser(ctx)
	values := map[string]any{"title": web.T(ctx, "media.avatar-title"), "status": statusText(ctx)}
	if avatar := First(UserOwner(user.ID), "avatar"); avatar != nil {
		values["avatarURL"] = URL(*avatar)
	}
	return web.View(ctx, "media::avatar", values)
}

func statusText(ctx http.Context) string {
	switch ctx.Request().Query("status") {
	case "saved":
		return web.T(ctx, "media.avatar-saved")
	case "removed":
		return web.T(ctx, "media.avatar-removed")
	}
	return ""
}

// UploadAvatar is POST /account/avatar: one image, it replaces the old one.
func UploadAvatar(ctx http.Context) http.Response {
	user := auth.CurrentUser(ctx)
	file, err := ctx.Request().File("avatar")
	if err == nil {
		_, err = Add(UserOwner(user.ID), "avatar", file)
	}
	if err != nil {
		key := "media.error-upload"
		switch {
		case errors.Is(err, ErrTooLarge):
			key = "media.error-too-large"
		case errors.Is(err, ErrWrongType):
			key = "media.error-type"
		}
		return web.ViewStatus(ctx, http.StatusUnprocessableEntity, "media::avatar", map[string]any{
			"title": web.T(ctx, "media.avatar-title"), "error": web.T(ctx, key),
		})
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/account/avatar?status=saved")
}

// RemoveAvatar is POST /account/avatar/delete.
func RemoveAvatar(ctx http.Context) http.Response {
	user := auth.CurrentUser(ctx)
	for _, item := range Get(UserOwner(user.ID), "avatar") {
		_ = Delete(item)
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/account/avatar?status=removed")
}

// AdminIndex is GET /admin/media: every file, newest first; ?collection= and ?type= filter.
func AdminIndex(ctx http.Context) http.Response {
	query := facades.Orm().Query().Model(&models.Media{}).OrderBy("id", "desc")
	if collection := strings.TrimSpace(ctx.Request().Query("collection")); collection != "" {
		query = query.Where("collection_name", collection)
	}
	if ownerType := strings.TrimSpace(ctx.Request().Query("type")); ownerType != "" {
		query = query.Where("model_type", ownerType)
	}
	var items []models.Media
	var total int64
	page := pageParam(ctx)
	_ = query.Paginate(page, 50, &items, &total)
	urls := map[uint]string{}
	for _, item := range items {
		urls[item.ID] = URL(item)
	}
	return adminView(ctx, "media::admin-index", "Media", map[string]any{
		"items": items, "urls": urls, "total": total, "collection": ctx.Request().Query("collection"), "type": ctx.Request().Query("type"),
	})
}

// AdminUpload is POST /admin/media: a file into the shared library (public or private as chosen).
func AdminUpload(ctx http.Context) http.Response {
	collection := "files"
	if ctx.Request().Input("visibility") == "public" {
		collection = "public-files"
	}
	file, err := ctx.Request().File("file")
	var item models.Media
	if err == nil {
		item, err = Add(Owner{Type: "library", ID: "0"}, collection, file)
	}
	if err != nil {
		code := "upload"
		if errors.Is(err, ErrTooLarge) {
			code = "too-large"
		}
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/media?error="+code)
	}
	if user := auth.CurrentUser(ctx); user != nil {
		activity.In("media").CausedBy(user.ID).PerformedOn("media", item.ID).Event("created").
			With("file_name", item.FileName).LogQuietly("uploaded a file")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/media?saved=uploaded")
}

// AdminDelete is POST /admin/media/{id}/delete.
func AdminDelete(ctx http.Context) http.Response {
	id, _ := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	var item models.Media
	if id == 0 || facades.Orm().Query().Find(&item, id) != nil || item.ID == 0 {
		return notFound(ctx)
	}
	if err := Delete(item); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/media?error=delete")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		activity.In("media").CausedBy(user.ID).PerformedOn("media", item.ID).Event("deleted").
			With("file_name", item.FileName).LogQuietly("deleted a file")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/media?saved=deleted")
}

func pageParam(ctx http.Context) int {
	page, _ := strconv.Atoi(ctx.Request().Query("page", "1"))
	if page < 1 {
		return 1
	}
	return page
}
