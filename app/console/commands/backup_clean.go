package commands

import (
	"strconv"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/services/backup"
)

// BackupClean is `backup:clean` (Spatie's): delete the backups the retention settings no longer keep.
type BackupClean struct {
}

// Signature The name and signature of the console command.
func (r *BackupClean) Signature() string {
	return "backup:clean"
}

// Description The console command description.
func (r *BackupClean) Description() string {
	return "Delete old database backups by the retention settings (config backup.cleanup)"
}

// Extend The console command extend.
func (r *BackupClean) Extend() command.Extend {
	return command.Extend{Category: "backup"}
}

// Handle Execute the console command.
func (r *BackupClean) Handle(ctx console.Context) error {
	removed, kept, err := backup.Clean(time.Now())
	for _, item := range removed {
		ctx.Line("deleted " + item.Path)
	}
	if err != nil {
		ctx.Error(err.Error())
		_ = backup.Notify("Cleaning the backups failed", err.Error())
		return err
	}
	ctx.Success("deleted " + strconv.Itoa(len(removed)) + ", kept " + strconv.Itoa(len(kept)))
	return nil
}
