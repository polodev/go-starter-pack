package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("backup", map[string]any{
		// Folder that holds this app's backups on the disk (like Spatie's backup.name), under the prefix.
		"name":   config.Env("BACKUP_NAME", "go-starter-pack"),
		"prefix": config.Env("BACKUP_PREFIX", "db-backups"),
		// Disk from config/filesystems.go. Keep backups on the private bucket, never on the public one.
		"disk": config.Env("BACKUP_DISK", "r2_backups"),
		// Where the compressed dump waits before upload (Spatie: storage/app/backup-temp). Needs free space for one
		// compressed dump; "" means storage/app/backup-temp.
		"temporary_directory": config.Env("BACKUP_TEMPORARY_DIRECTORY", ""),
		// mysqldump, mariadb-dump or pg_dump when it is not on the PATH.
		"dump_binary": config.Env("BACKUP_DUMP_BINARY", ""),
	})
}
