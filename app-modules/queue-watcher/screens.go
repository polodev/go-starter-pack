package queuewatcher

import (
	"strconv"
	"strings"
	"time"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/queue"

	adminlayout "goravel/app-modules/admin-layout"
	"goravel/app-modules/auth"
	"goravel/app-modules/queue-watcher/models"
	"goravel/app/activity"
	"goravel/app/facades"
)

// Base is where the queue monitor lives in the admin.
const Base = "/admin/queue-watcher"

const perPage = 50

var (
	crumbQueue   = adminlayout.Crumb{Label: "Queue monitor", Href: Base}
	crumbHistory = adminlayout.Crumb{Label: "Job history", Href: Base + "/jobs"}
	crumbPending = adminlayout.Crumb{Label: "Pending jobs", Href: Base + "/pending"}
	crumbFailed  = adminlayout.Crumb{Label: "Failed jobs", Href: Base + "/failed"}
)

type queueCount struct {
	Queue string `gorm:"column:queue"`
	Total int64  `gorm:"column:total"`
}

// Overview is GET /admin/queue-watcher: what waits, what runs, what failed, and the newest runs.
func Overview(ctx http.Context) http.Response {
	now := time.Now().UTC()
	jobs := func() contractsorm.Query { return facades.Orm().Query().Model(&models.PendingJob{}) }
	counts := map[string]int64{}
	counts["waiting"], _ = jobs().WhereNull("reserved_at").Where("available_at <= ?", now).Count()
	counts["delayed"], _ = jobs().WhereNull("reserved_at").Where("available_at > ?", now).Count()
	counts["running"], _ = jobs().WhereNotNull("reserved_at").Count()
	counts["failed"], _ = facades.Orm().Query().Model(&models.FailedJob{}).Count()
	history := map[string]int64{}
	for _, status := range models.Statuses {
		history[status], _ = facades.Orm().Query().Model(&models.QueueJob{}).Where("status", status).Count()
	}
	var queues []queueCount
	_ = facades.DB().Table("jobs").Select("queue, count(*) as total").GroupBy("queue").OrderBy("queue").Get(&queues)
	var recent []models.QueueJob
	_ = facades.Orm().Query().OrderBy("id", "desc").Limit(10).Find(&recent)
	return adminlayout.View(ctx, "queue-watcher::overview", "Queue monitor", []adminlayout.Crumb{{Label: "Queue monitor"}},
		map[string]any{"counts": counts, "history": history, "statuses": models.Statuses, "queues": queues, "recent": recent,
			"connection": facades.Config().GetString("queue.default")})
}

// History is GET /admin/queue-watcher/jobs: every tracked run, newest first, by ?status= and ?q= (job name).
func History(ctx http.Context) http.Response {
	filters := map[string]string{"status": strings.TrimSpace(ctx.Request().Query("status")), "q": strings.TrimSpace(ctx.Request().Query("q"))}
	query := facades.Orm().Query().Model(&models.QueueJob{}).OrderBy("id", "desc")
	if filters["status"] != "" {
		query = query.Where("status", filters["status"])
	}
	if filters["q"] != "" {
		query = query.Where("LOWER(job_class) LIKE ?", "%"+strings.ToLower(filters["q"])+"%")
	}
	var runs []models.QueueJob
	var total int64
	page := adminlayout.PageParam(ctx)
	_ = query.Paginate(page, perPage, &runs, &total)
	return adminlayout.View(ctx, "queue-watcher::history", "Job history", []adminlayout.Crumb{crumbQueue, {Label: "Job history"}},
		withPages(map[string]any{"runs": runs, "filters": filters, "statuses": models.Statuses, "total": total,
			"retentionDays": RetentionDays()}, page, total))
}

// HistoryShow is GET /admin/queue-watcher/jobs/{id}.
func HistoryShow(ctx http.Context) http.Response {
	var run models.QueueJob
	if !load(ctx, &run) {
		return adminlayout.NotFound(ctx)
	}
	id := strconv.Itoa(int(run.ID))
	return adminlayout.View(ctx, "queue-watcher::history-show", "Job run "+id, []adminlayout.Crumb{crumbQueue, crumbHistory, {Label: id}},
		map[string]any{"run": run})
}

// HistoryDelete is POST /admin/queue-watcher/jobs/{id}/delete (queue.manage).
func HistoryDelete(ctx http.Context) http.Response {
	var run models.QueueJob
	if !load(ctx, &run) {
		return adminlayout.NotFound(ctx)
	}
	if _, err := facades.Orm().Query().Delete(&run); err != nil {
		return back(ctx, "/jobs", "error=delete")
	}
	logChange(ctx, "queue_job", run.ID, "deleted", "deleted a job run from the history", map[string]any{"job": valueOf(run.JobClass)})
	return back(ctx, "/jobs", "saved=deleted")
}

// HistoryMarkStale is POST /admin/queue-watcher/jobs/mark-stale (queue.manage).
func HistoryMarkStale(ctx http.Context) http.Response {
	marked, err := MarkStale(StaleAfter())
	if err != nil {
		return back(ctx, "/jobs", "error=mark-stale")
	}
	logChange(ctx, "", 0, "marked-stale", "marked unfinished job runs as stale", map[string]any{"marked": marked})
	return back(ctx, "/jobs", "saved=marked-stale")
}

// HistoryPurge is POST /admin/queue-watcher/jobs/purge (queue.manage): runs older than the retention go.
func HistoryPurge(ctx http.Context) http.Response {
	removed, err := PurgeHistory(RetentionDays())
	if err != nil {
		return back(ctx, "/jobs", "error=purge")
	}
	logChange(ctx, "", 0, "purged", "purged the job history", map[string]any{"removed": removed})
	return back(ctx, "/jobs", "saved=purged")
}

// Pending is GET /admin/queue-watcher/pending: the jobs table, oldest first, by ?queue=.
func Pending(ctx http.Context) http.Response {
	queueName := strings.TrimSpace(ctx.Request().Query("queue"))
	query := facades.Orm().Query().Model(&models.PendingJob{}).OrderBy("id")
	if queueName != "" {
		query = query.Where("queue", queueName)
	}
	var jobs []models.PendingJob
	var total int64
	page := adminlayout.PageParam(ctx)
	_ = query.Paginate(page, perPage, &jobs, &total)
	now := time.Now().UTC()
	rows := make([]map[string]any, 0, len(jobs))
	for _, job := range jobs {
		rows = append(rows, map[string]any{"job": job, "payload": ReadPayload(job.Payload), "state": job.State(now)})
	}
	return adminlayout.View(ctx, "queue-watcher::pending", "Pending jobs", []adminlayout.Crumb{crumbQueue, {Label: "Pending jobs"}},
		withPages(map[string]any{"rows": rows, "queue": queueName, "total": total}, page, total))
}

// PendingShow is GET /admin/queue-watcher/pending/{id}.
func PendingShow(ctx http.Context) http.Response {
	var job models.PendingJob
	if !load(ctx, &job) {
		return adminlayout.NotFound(ctx)
	}
	id := strconv.Itoa(int(job.ID))
	return adminlayout.View(ctx, "queue-watcher::pending-show", "Pending job "+id, []adminlayout.Crumb{crumbQueue, crumbPending, {Label: id}},
		map[string]any{"job": job, "payload": ReadPayload(job.Payload), "state": job.State(time.Now().UTC())})
}

// PendingDelete is POST /admin/queue-watcher/pending/{id}/delete (queue.manage). A job a worker is running is
// left alone: the worker owns it until it finishes.
func PendingDelete(ctx http.Context) http.Response {
	var job models.PendingJob
	if !load(ctx, &job) {
		return adminlayout.NotFound(ctx)
	}
	if job.ReservedAt != nil {
		return back(ctx, "/pending/"+strconv.Itoa(int(job.ID)), "error=running")
	}
	if _, err := facades.Orm().Query().Where("id", job.ID).WhereNull("reserved_at").Delete(&models.PendingJob{}); err != nil {
		return back(ctx, "/pending", "error=delete")
	}
	logChange(ctx, "job", job.ID, "deleted", "deleted a pending job", map[string]any{"job": ReadPayload(job.Payload).Signature, "queue": job.Queue})
	return back(ctx, "/pending", "saved=deleted")
}

// Failed is GET /admin/queue-watcher/failed: failed_jobs, newest first, by ?queue=.
func Failed(ctx http.Context) http.Response {
	queueName := strings.TrimSpace(ctx.Request().Query("queue"))
	query := facades.Orm().Query().Model(&models.FailedJob{}).OrderBy("id", "desc")
	if queueName != "" {
		query = query.Where("queue", queueName)
	}
	var jobs []models.FailedJob
	var total int64
	page := adminlayout.PageParam(ctx)
	_ = query.Paginate(page, perPage, &jobs, &total)
	rows := make([]map[string]any, 0, len(jobs))
	for _, job := range jobs {
		rows = append(rows, map[string]any{"job": job, "payload": ReadPayload(job.Payload), "error": firstLine(job.Exception)})
	}
	return adminlayout.View(ctx, "queue-watcher::failed", "Failed jobs", []adminlayout.Crumb{crumbQueue, {Label: "Failed jobs"}},
		withPages(map[string]any{"rows": rows, "queue": queueName, "total": total}, page, total))
}

// FailedShow is GET /admin/queue-watcher/failed/{id}: the whole error and the payload.
func FailedShow(ctx http.Context) http.Response {
	var job models.FailedJob
	if !load(ctx, &job) {
		return adminlayout.NotFound(ctx)
	}
	id := strconv.Itoa(int(job.ID))
	return adminlayout.View(ctx, "queue-watcher::failed-show", "Failed job "+id, []adminlayout.Crumb{crumbQueue, crumbFailed, {Label: id}},
		map[string]any{"job": job, "payload": ReadPayload(job.Payload)})
}

// FailedRetry is POST /admin/queue-watcher/failed/{id}/retry (queue.manage): Goravel's queue:retry for one job. The
// job goes back on its queue and leaves failed_jobs.
func FailedRetry(ctx http.Context) http.Response {
	var job models.FailedJob
	if !load(ctx, &job) {
		return adminlayout.NotFound(ctx)
	}
	if strings.TrimSpace(job.UUID) == "" {
		// Failer().Get with no uuid would return every failed job.
		return back(ctx, "/failed/"+strconv.Itoa(int(job.ID)), "error=retry")
	}
	found, err := facades.Queue().Failer().Get("", "", []string{job.UUID})
	if err != nil || len(found) != 1 || found[0].Retry() != nil {
		return back(ctx, "/failed/"+strconv.Itoa(int(job.ID)), "error=retry")
	}
	logChange(ctx, "failed_job", job.ID, "retried", "retried a failed job", map[string]any{"job": ReadPayload(job.Payload).Signature, "queue": job.Queue})
	return back(ctx, "/failed", "saved=retried")
}

// FailedRetryAll is POST /admin/queue-watcher/failed/retry-all (queue.manage).
func FailedRetryAll(ctx http.Context) http.Response {
	all, err := facades.Queue().Failer().All()
	if err != nil {
		return back(ctx, "/failed", "error=retry")
	}
	retried, failed := retryEach(all)
	logChange(ctx, "", 0, "retried", "retried all failed jobs", map[string]any{"retried": retried, "could_not_retry": failed})
	if failed > 0 {
		return back(ctx, "/failed", "error=retry")
	}
	return back(ctx, "/failed", "saved=retried")
}

func retryEach(jobs []queue.FailedJob) (retried, failed int) {
	for _, job := range jobs {
		if job.Retry() != nil {
			failed++
			continue
		}
		retried++
	}
	return retried, failed
}

// FailedDelete is POST /admin/queue-watcher/failed/{id}/delete (queue.manage).
func FailedDelete(ctx http.Context) http.Response {
	var job models.FailedJob
	if !load(ctx, &job) {
		return adminlayout.NotFound(ctx)
	}
	if _, err := facades.Orm().Query().Where("id", job.ID).Delete(&models.FailedJob{}); err != nil {
		return back(ctx, "/failed", "error=delete")
	}
	logChange(ctx, "failed_job", job.ID, "deleted", "deleted a failed job", map[string]any{"job": ReadPayload(job.Payload).Signature, "queue": job.Queue})
	return back(ctx, "/failed", "saved=deleted")
}

// FailedFlush is POST /admin/queue-watcher/failed/flush (queue.manage): Laravel's queue:flush.
func FailedFlush(ctx http.Context) http.Response {
	result, err := facades.Orm().Query().Where("1 = 1").Delete(&models.FailedJob{})
	if err != nil {
		return back(ctx, "/failed", "error=flush")
	}
	logChange(ctx, "", 0, "flushed", "deleted all failed jobs", map[string]any{"removed": result.RowsAffected})
	return back(ctx, "/failed", "saved=flushed")
}

// load finds the row of {id} into model; false for a bad or unknown id.
func load(ctx http.Context, model any) bool {
	id := adminlayout.IDParam(ctx)
	if id == 0 {
		return false
	}
	if err := facades.Orm().Query().Where("id", id).First(model); err != nil {
		return false
	}
	switch row := model.(type) {
	case *models.QueueJob:
		return row.ID != 0
	case *models.PendingJob:
		return row.ID != 0
	case *models.FailedJob:
		return row.ID != 0
	}
	return false
}

func back(ctx http.Context, path, feedback string) http.Response {
	return ctx.Response().Redirect(http.StatusSeeOther, Base+path+"?"+feedback)
}

func logChange(ctx http.Context, subjectType string, subjectID uint, event, description string, details map[string]any) {
	user := auth.CurrentUser(ctx)
	if user == nil {
		return
	}
	entry := activity.In("queue").CausedBy(user.ID).Event(event)
	if subjectType != "" {
		entry = entry.PerformedOn(subjectType, subjectID)
	}
	for key, value := range details {
		entry = entry.With(key, value)
	}
	entry.LogQuietly(description)
}

func withPages(values map[string]any, page int, total int64) map[string]any {
	pages := int((total + perPage - 1) / perPage)
	values["page"], values["prev"], values["next"] = page, page-1, 0
	if page < pages {
		values["next"] = page + 1
	}
	return values
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	if len(line) > 160 {
		return line[:160] + "…"
	}
	return line
}

func valueOf(text *string) string {
	if text == nil {
		return ""
	}
	return *text
}
