package commands

import (
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/services/backup"
)

// BackupRun is `backup:run`: like `php artisan backup:run --only-db`.
type BackupRun struct {
}

// Signature The name and signature of the console command.
func (r *BackupRun) Signature() string {
	return "backup:run"
}

// Description The console command description.
func (r *BackupRun) Description() string {
	return "Dump the database (gzip) and store it on the private backup disk"
}

// Extend The console command extend.
func (r *BackupRun) Extend() command.Extend {
	return command.Extend{Category: "backup"}
}

// Handle Execute the console command.
func (r *BackupRun) Handle(ctx console.Context) error {
	stored, err := backup.Run(time.Now())
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("stored " + stored.Path + " on " + stored.Disk)
	return nil
}
