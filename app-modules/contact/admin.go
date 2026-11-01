package contact

import (
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/auth"
	"goravel/app-modules/contact/models"
	"goravel/app/activity"
	"goravel/app/facades"
)

const perPage = 50

// AdminIndex is GET /admin/contacts: newest first, by ?status= and ?q= (name, email or subject).
func AdminIndex(ctx http.Context) http.Response {
	filters := map[string]string{"status": strings.TrimSpace(ctx.Request().Query("status")), "q": strings.TrimSpace(ctx.Request().Query("q"))}
	query := facades.Orm().Query().Model(&models.Contact{}).OrderBy("id", "desc")
	if filters["status"] != "" {
		query = query.Where("status", filters["status"])
	}
	if filters["q"] != "" {
		// LOWER on both sides: LIKE ignores case on MySQL but not on PostgreSQL.
		like := "%" + strings.ToLower(filters["q"]) + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(subject) LIKE ?)", like, like, like)
	}
	var messages []models.Contact
	var total int64
	page := adminlayout.PageParam(ctx)
	_ = query.Paginate(page, perPage, &messages, &total)
	waiting, _ := facades.Orm().Query().Model(&models.Contact{}).Where("status", models.StatusNew).Count()
	pages := int((total + perPage - 1) / perPage)
	next := 0
	if page < pages {
		next = page + 1
	}
	return adminlayout.View(ctx, "contact::admin-index", "Contact messages", []adminlayout.Crumb{{Label: "Contact messages"}},
		map[string]any{"messages": messages, "filters": filters, "statuses": models.Statuses, "total": total, "waiting": waiting,
			"page": page, "prev": page - 1, "next": next})
}

func find(ctx http.Context) (models.Contact, bool) {
	var message models.Contact
	id := adminlayout.IDParam(ctx)
	if id == 0 || facades.Orm().Query().Where("id", id).First(&message) != nil || message.ID == 0 {
		return message, false
	}
	return message, true
}

// AdminShow is GET /admin/contacts/{id}: the whole message, a reply link and the status form.
func AdminShow(ctx http.Context) http.Response {
	message, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	id := strconv.Itoa(int(message.ID))
	subject := message.SubjectText()
	if subject == "" {
		subject = facades.Config().GetString("app.name")
	}
	reply := "mailto:" + message.Email + "?subject=" + url.PathEscape("Re: "+subject)
	return adminlayout.View(ctx, "contact::admin-show", "Message "+id,
		[]adminlayout.Crumb{{Label: "Contact messages", Href: "/admin/contacts"}, {Label: id}},
		map[string]any{"message": message, "statuses": models.Statuses, "replyLink": reply})
}

// AdminStatus is POST /admin/contacts/{id}/status (contacts.manage): new or replied.
func AdminStatus(ctx http.Context) http.Response {
	message, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	status := ctx.Request().Input("status")
	back := "/admin/contacts/" + strconv.Itoa(int(message.ID))
	if !slices.Contains(models.Statuses, status) {
		return ctx.Response().Redirect(http.StatusSeeOther, back+"?error=status")
	}
	changes := map[string]any{"status": status, "updated_at": time.Now().UTC(), "replied_at": nil}
	if status == models.StatusReplied {
		changes["replied_at"] = time.Now().UTC()
	}
	if _, err := facades.Orm().Query().Model(&models.Contact{}).Where("id", message.ID).Update(changes); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, back+"?error=status")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		activity.In("contacts").CausedBy(user.ID).PerformedOn("contact", message.ID).Event("status_changed").
			Changes(map[string]any{"status": message.Status}, map[string]any{"status": status}).LogQuietly("changed a contact message's status")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, back+"?saved=status")
}

// AdminDelete is POST /admin/contacts/{id}/delete (contacts.manage).
func AdminDelete(ctx http.Context) http.Response {
	message, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	if _, err := facades.Orm().Query().Where("id", message.ID).Delete(&models.Contact{}); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/contacts?error=delete")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		// No name or email in the entry: a deletion may be someone asking to be forgotten.
		activity.In("contacts").CausedBy(user.ID).PerformedOn("contact", message.ID).Event("deleted").
			LogQuietly("deleted a contact message")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/contacts?saved=deleted")
}
