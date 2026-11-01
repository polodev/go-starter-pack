// Package queuewatcher is the queue monitor: the history of every job run (queue_jobs), the jobs waiting in
// Goravel's database queue (jobs) and the failed ones (failed_jobs), with retry and delete. Like Laravel Horizon's
// dashboard or the Laravel jobs monitor, for the database queue.
package queuewatcher

import (
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/schema"
	contractsschedule "github.com/goravel/framework/contracts/schedule"

	"goravel/app-modules/queue-watcher/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/modular"
)

// Module registers the history table, the tracking of every job, the screens and the purge command.
type Module struct{}

func (m *Module) Name() string      { return "queue-watcher" }
func (m *Module) ViewsPath() string { return "app-modules/queue-watcher/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261102110000CreateQueueJobsTable{}}
}

// Boot wraps every registered job (Goravel's mail job too) so each run is written to the history.
func (m *Module) Boot() { TrackAll() }

func (m *Module) Commands() []console.Command { return []console.Command{&PurgeCommand{}} }

func (m *Module) Schedule() []contractsschedule.Event {
	return []contractsschedule.Event{facades.Schedule().Command("queue-watcher:purge").DailyAt("02:30").OnOneServer()}
}

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "System", Label: "Queue monitor", Href: Base, Permission: "queue.view", Icon: "⇄", Order: 2}}
}

func (m *Module) Routes() {
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/queue-watcher", "queue.view", Overview)
		admin.Get("/queue-watcher/jobs", "queue.view", History)
		admin.Post("/queue-watcher/jobs/mark-stale", "queue.manage", HistoryMarkStale)
		admin.Post("/queue-watcher/jobs/purge", "queue.manage", HistoryPurge)
		admin.Get("/queue-watcher/jobs/{id}", "queue.view", HistoryShow)
		admin.Post("/queue-watcher/jobs/{id}/delete", "queue.manage", HistoryDelete)
		admin.Get("/queue-watcher/pending", "queue.view", Pending)
		admin.Get("/queue-watcher/pending/{id}", "queue.view", PendingShow)
		admin.Post("/queue-watcher/pending/{id}/delete", "queue.manage", PendingDelete)
		admin.Get("/queue-watcher/failed", "queue.view", Failed)
		admin.Post("/queue-watcher/failed/retry-all", "queue.manage", FailedRetryAll)
		admin.Post("/queue-watcher/failed/flush", "queue.manage", FailedFlush)
		admin.Get("/queue-watcher/failed/{id}", "queue.view", FailedShow)
		admin.Post("/queue-watcher/failed/{id}/retry", "queue.manage", FailedRetry)
		admin.Post("/queue-watcher/failed/{id}/delete", "queue.manage", FailedDelete)
	})
}

// PurgeCommand is `queue-watcher:purge`: marks runs that never finished as stale, then deletes history older than
// QUEUE_WATCHER_RETENTION_DAYS (or --days). Waiting and failed jobs are not touched.
type PurgeCommand struct{}

func (r *PurgeCommand) Signature() string { return "queue-watcher:purge" }
func (r *PurgeCommand) Description() string {
	return "Mark unfinished job runs as stale and delete job history older than N days (default QUEUE_WATCHER_RETENTION_DAYS or 30)"
}
func (r *PurgeCommand) Extend() command.Extend {
	return command.Extend{Category: "queue-watcher", Flags: []command.Flag{&command.IntFlag{Name: "days", Usage: "keep this many days"}}}
}

func (r *PurgeCommand) Handle(ctx console.Context) error {
	days := ctx.OptionInt("days")
	if days <= 0 {
		days = RetentionDays()
	}
	marked, err := MarkStale(StaleAfter())
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	removed, err := PurgeHistory(days)
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("marked " + strconv.FormatInt(marked, 10) + " stale, removed " + strconv.FormatInt(removed, 10) +
		" run(s) older than " + strconv.Itoa(days) + " days")
	return nil
}
