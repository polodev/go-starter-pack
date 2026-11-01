package activitylog

import (
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app-modules/activity-log/models"
	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/auth"
	"goravel/app/activity"
	"goravel/app/facades"
)

const perPage = 50

// Index is GET /admin/activity-logs: newest first, filtered by ?log=, ?event=, ?causer= and searched with ?q=.
func Index(ctx http.Context) http.Response {
	query := facades.Orm().Query().Model(&models.Activity{}).OrderBy("id", "desc")
	filters := map[string]string{}
	for _, field := range []string{"log", "event", "causer", "q"} {
		filters[field] = strings.TrimSpace(ctx.Request().Query(field))
	}
	if filters["log"] != "" {
		query = query.Where("log_name", filters["log"])
	}
	if filters["event"] != "" {
		query = query.Where("event", filters["event"])
	}
	if filters["causer"] != "" {
		query = query.Where("causer_id", filters["causer"])
	}
	if filters["q"] != "" {
		query = query.Where("description LIKE ?", "%"+filters["q"]+"%")
	}
	var entries []models.Activity
	var total int64
	page := adminlayout.PageParam(ctx)
	_ = query.Paginate(page, perPage, &entries, &total)
	var logNames []string
	_ = facades.Orm().Query().Model(&models.Activity{}).Distinct("log_name").WhereNotNull("log_name").Pluck("log_name", &logNames)
	pages := int((total + perPage - 1) / perPage)
	next := 0
	if page < pages {
		next = page + 1
	}
	return adminlayout.View(ctx, "activity-log::index", "Activity logs", []adminlayout.Crumb{{Label: "Activity logs"}}, map[string]any{
		"entries": entries, "filters": filters, "logNames": logNames, "total": total, "page": page, "prev": page - 1, "next": next,
	})
}

func find(ctx http.Context) (models.Activity, bool) {
	var entry models.Activity
	id := adminlayout.IDParam(ctx)
	if id == 0 || facades.Orm().Query().Find(&entry, id) != nil || entry.ID == 0 {
		return entry, false
	}
	return entry, true
}

// Show is GET /admin/activity-logs/{id}: who, what, and the field-by-field changes.
func Show(ctx http.Context) http.Response {
	entry, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	return adminlayout.View(ctx, "activity-log::show", "Activity "+strconv.Itoa(int(entry.ID)),
		[]adminlayout.Crumb{{Label: "Activity logs", Href: "/admin/activity-logs"}, {Label: strconv.Itoa(int(entry.ID))}},
		map[string]any{"entry": entry, "changes": entry.Changes(), "properties": entry.PrettyProperties()})
}

// Delete is POST /admin/activity-logs/{id}/delete (activity_logs.manage).
func Delete(ctx http.Context) http.Response {
	entry, ok := find(ctx)
	if !ok {
		return adminlayout.NotFound(ctx)
	}
	if _, err := facades.Orm().Query().Delete(&entry); err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/activity-logs?error=delete")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/activity-logs?saved=deleted")
}

// Clean is POST /admin/activity-logs/clean (activity_logs.manage): removes entries older than the retention.
func Clean(ctx http.Context) http.Response {
	removed, err := CleanOlderThan(RetentionDays())
	if err != nil {
		return ctx.Response().Redirect(http.StatusSeeOther, "/admin/activity-logs?error=clean")
	}
	if user := auth.CurrentUser(ctx); user != nil {
		activity.In("activity_logs").CausedBy(user.ID).Event("cleaned").With("removed", removed).LogQuietly("cleaned the activity log")
	}
	return ctx.Response().Redirect(http.StatusSeeOther, "/admin/activity-logs?saved=cleaned")
}

// RetentionDays is ACTIVITY_LOG_RETENTION_DAYS (default 365), like Spatie's delete_records_older_than_days.
func RetentionDays() int {
	days := facades.Config().GetInt("activity_log.retention_days", 365)
	if text, ok := facades.Config().Env("ACTIVITY_LOG_RETENTION_DAYS", "").(string); ok && text != "" {
		if parsed, err := strconv.Atoi(text); err == nil && parsed > 0 {
			days = parsed
		}
	}
	return days
}

// CleanOlderThan deletes entries older than days; returns how many.
func CleanOlderThan(days int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	result, err := facades.Orm().Query().Where("created_at < ?", cutoff).Delete(&models.Activity{})
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}
