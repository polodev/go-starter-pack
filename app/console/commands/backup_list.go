package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"

	"goravel/app/services/backup"
)

// BackupList is `backup:list`: the stored backups, oldest first.
type BackupList struct {
}

// Signature The name and signature of the console command.
func (r *BackupList) Signature() string {
	return "backup:list"
}

// Description The console command description.
func (r *BackupList) Description() string {
	return "List the database backups on the backup disk"
}

// Extend The console command extend.
func (r *BackupList) Extend() command.Extend {
	return command.Extend{Category: "backup"}
}

// Handle Execute the console command.
func (r *BackupList) Handle(ctx console.Context) error {
	backups, err := backup.List()
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	if len(backups) == 0 {
		ctx.Info("no backups yet")
	}
	for _, path := range backups {
		ctx.Line(path)
	}
	return nil
}
