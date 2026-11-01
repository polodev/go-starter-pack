// Package activitylog owns Spatie's activity_log table and its admin screens (list, filters, show with the
// changes, delete, clean). Any module writes entries with app/activity.
package activitylog

import (
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/schema"
	contractsschedule "github.com/goravel/framework/contracts/schedule"

	"goravel/app-modules/activity-log/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/modular"
)

// Module registers the table, the screens, the clean command and its schedule.
type Module struct{}

func (m *Module) Name() string      { return "activity-log" }
func (m *Module) ViewsPath() string { return "app-modules/activity-log/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{&migrations.M20261101160000CreateActivityLogTable{}}
}

func (m *Module) Commands() []console.Command { return []console.Command{&CleanCommand{}} }

func (m *Module) Schedule() []contractsschedule.Event {
	return []contractsschedule.Event{facades.Schedule().Command("activity-log:clean").DailyAt("02:00").OnOneServer()}
}

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "System", Label: "Activity logs", Href: "/admin/activity-logs", Permission: "activity_logs.view", Icon: "☰", Order: 1}}
}

func (m *Module) Routes() {
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/activity-logs", "activity_logs.view", Index)
		admin.Post("/activity-logs/clean", "activity_logs.manage", Clean)
		admin.Get("/activity-logs/{id}", "activity_logs.view", Show)
		admin.Post("/activity-logs/{id}/delete", "activity_logs.manage", Delete)
	})
}

// CleanCommand is `activity-log:clean --days=365` (Spatie's activitylog:clean).
type CleanCommand struct{}

func (r *CleanCommand) Signature() string { return "activity-log:clean" }
func (r *CleanCommand) Description() string {
	return "Delete activity log entries older than N days (default ACTIVITY_LOG_RETENTION_DAYS or 365)"
}
func (r *CleanCommand) Extend() command.Extend {
	return command.Extend{Category: "activity-log", Flags: []command.Flag{&command.IntFlag{Name: "days", Usage: "keep this many days"}}}
}

func (r *CleanCommand) Handle(ctx console.Context) error {
	days := ctx.OptionInt("days")
	if days <= 0 {
		days = RetentionDays()
	}
	removed, err := CleanOlderThan(days)
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("removed " + strconv.FormatInt(removed, 10) + " entr(y/ies) older than " + strconv.Itoa(days) + " days")
	return nil
}
