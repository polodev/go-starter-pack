package bootstrap

import (
	"github.com/goravel/framework/contracts/schedule"

	"goravel/app/facades"
	"goravel/app/modular"
)

// Schedule is routes/console.php: the web process runs these (Goravel's schedule runner). Run one web process,
// or keep OnOneServer, so a task does not run twice.
func Schedule() []schedule.Event {
	events := []schedule.Event{}
	if bucket, _ := facades.Config().Env("BACKUP_BUCKET", "").(string); bucket != "" {
		// Spatie's order: back up, then clean (so the size limit counts tonight's backup), then check.
		events = append(events,
			facades.Schedule().Command("backup:run").DailyAt("01:30").OnOneServer(),
			facades.Schedule().Command("backup:clean").DailyAt("02:00").OnOneServer(),
			facades.Schedule().Command("backup:monitor").DailyAt("03:00").OnOneServer(),
		)
	}
	return append(events, modular.Collect(Modules(), func(m modular.HasSchedule) []schedule.Event { return m.Schedule() })...)
}
