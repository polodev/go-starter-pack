package commands

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/services/backup"
)

// BackupMonitor is `backup:monitor` (Spatie's): is the newest backup recent and is the storage within the limit?
// An unhealthy result is logged, mailed to BACKUP_NOTIFICATION_EMAIL and ends with a non-zero exit.
type BackupMonitor struct {
}

// Signature The name and signature of the console command.
func (r *BackupMonitor) Signature() string {
	return "backup:monitor"
}

// Description The console command description.
func (r *BackupMonitor) Description() string {
	return "Check that the newest database backup is recent and the backups fit their storage limit"
}

// Extend The console command extend.
func (r *BackupMonitor) Extend() command.Extend {
	return command.Extend{Category: "backup"}
}

// Handle Execute the console command.
func (r *BackupMonitor) Handle(ctx console.Context) error {
	health, _ := backup.CheckHealth(time.Now())
	if !health.Healthy {
		problems := strings.Join(health.Problems, "; ")
		ctx.Error("unhealthy: " + problems)
		_ = backup.Notify("Backups are unhealthy", problems)
		return errors.New("backups are unhealthy: " + problems)
	}
	ctx.Success(fmt.Sprintf("healthy: %d backups, %.1f MB, newest %s", health.Count, float64(health.UsedSize)/1024/1024,
		health.Newest.Time.Format(time.RFC3339)))
	return nil
}
