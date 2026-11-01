package bootstrap

import (
	"github.com/goravel/framework/contracts/schedule"

	"goravel/app/facades"
)

// Schedule is routes/console.php: the web process runs these (Goravel's schedule runner). Run one web process,
// or keep OnOneServer, so a task does not run twice.
func Schedule() []schedule.Event {
	events := []schedule.Event{}
	if bucket, _ := facades.Config().Env("BACKUP_BUCKET", "").(string); bucket != "" {
		events = append(events, facades.Schedule().Command("backup:run").DailyAt("01:30").OnOneServer())
	}
	return events
}
