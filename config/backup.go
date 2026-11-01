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
		// Who hears about a failed backup or an unhealthy one (backup:run, backup:monitor); "" means the log only.
		"notification_email": config.Env("BACKUP_NOTIFICATION_EMAIL", ""),
		// backup:clean, spatie/laravel-backup's default strategy and names: keep every backup for 7 days, then the
		// newest of each day for 16 days, of each week for 8 weeks, of each month for 4 months, of each year for
		// 2 years; then delete the oldest while they use more than the limit. The newest is never deleted.
		"cleanup": map[string]any{
			"keep_all_backups_for_days":                            config.Env("BACKUP_KEEP_ALL_BACKUPS_FOR_DAYS", 7),
			"keep_daily_backups_for_days":                          config.Env("BACKUP_KEEP_DAILY_BACKUPS_FOR_DAYS", 16),
			"keep_weekly_backups_for_weeks":                        config.Env("BACKUP_KEEP_WEEKLY_BACKUPS_FOR_WEEKS", 8),
			"keep_monthly_backups_for_months":                      config.Env("BACKUP_KEEP_MONTHLY_BACKUPS_FOR_MONTHS", 4),
			"keep_yearly_backups_for_years":                        config.Env("BACKUP_KEEP_YEARLY_BACKUPS_FOR_YEARS", 2),
			"delete_oldest_backups_when_using_more_megabytes_than": config.Env("BACKUP_DELETE_OLDEST_WHEN_USING_MORE_MEGABYTES_THAN", 5000),
		},
		// backup:monitor: unhealthy when the newest backup is older than this or they use more than this.
		"monitor": map[string]any{
			"maximum_age_in_days":          config.Env("BACKUP_MAXIMUM_AGE_IN_DAYS", 1),
			"maximum_storage_in_megabytes": config.Env("BACKUP_MAXIMUM_STORAGE_IN_MEGABYTES", 5000),
		},
	})
}
