package backup

import (
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/mail"

	"goravel/app/facades"
)

// Inventory lists the stored backups with the time from their names and their sizes, newest first. Files whose
// names are not a backup time are left out (and so never deleted by Clean).
func Inventory() ([]Backup, error) {
	paths, err := List()
	if err != nil {
		return nil, err
	}
	disk := facades.Storage().Disk(facades.Config().GetString("backup.disk"))
	backups := make([]Backup, 0, len(paths))
	for index := len(paths) - 1; index >= 0; index-- {
		taken, ok := TimeFromPath(paths[index])
		if !ok {
			continue
		}
		size, err := disk.Size(paths[index])
		if err != nil {
			return nil, err
		}
		backups = append(backups, Backup{Path: paths[index], Time: taken, Size: size})
	}
	return backups, nil
}

// ConfiguredRetention is config backup.cleanup (Spatie's names), falling back to Spatie's defaults.
func ConfiguredRetention() Retention {
	config := facades.Config()
	return Retention{
		KeepAllBackupsForDays:       config.GetInt("backup.cleanup.keep_all_backups_for_days", DefaultRetention.KeepAllBackupsForDays),
		KeepDailyBackupsForDays:     config.GetInt("backup.cleanup.keep_daily_backups_for_days", DefaultRetention.KeepDailyBackupsForDays),
		KeepWeeklyBackupsForWeeks:   config.GetInt("backup.cleanup.keep_weekly_backups_for_weeks", DefaultRetention.KeepWeeklyBackupsForWeeks),
		KeepMonthlyBackupsForMonths: config.GetInt("backup.cleanup.keep_monthly_backups_for_months", DefaultRetention.KeepMonthlyBackupsForMonths),
		KeepYearlyBackupsForYears:   config.GetInt("backup.cleanup.keep_yearly_backups_for_years", DefaultRetention.KeepYearlyBackupsForYears),
		DeleteOldestBackupsWhenUsingMoreMegabytesThan: int64(config.GetInt("backup.cleanup.delete_oldest_backups_when_using_more_megabytes_than",
			int(DefaultRetention.DeleteOldestBackupsWhenUsingMoreMegabytesThan))),
	}
}

// Clean deletes the backups the retention plan does not keep (backup:clean) and returns what it deleted.
func Clean(now time.Time) (removed []Backup, kept []Backup, err error) {
	backups, err := Inventory()
	if err != nil {
		return nil, nil, err
	}
	kept, remove := Plan(backups, now, ConfiguredRetention())
	disk := facades.Storage().Disk(facades.Config().GetString("backup.disk"))
	for _, item := range remove {
		if err := disk.Delete(item.Path); err != nil {
			return removed, kept, fmt.Errorf("could not delete %s: %w", item.Path, err)
		}
		removed = append(removed, item)
	}
	return removed, kept, nil
}

// Health is what backup:monitor checks: a recent enough newest backup, and not too much storage used.
type Health struct {
	Healthy  bool
	Problems []string
	Count    int
	Newest   *Backup
	UsedSize int64
}

// CheckHealth is Spatie's backup:monitor: unhealthy when there is no backup, the newest is older than
// backup.monitor.maximum_age_in_days, or all of them together use more than maximum_storage_in_megabytes.
func CheckHealth(now time.Time) (Health, error) {
	backups, err := Inventory()
	if err != nil {
		return Health{Problems: []string{"cannot read the backup disk: " + err.Error()}}, err
	}
	config := facades.Config()
	health := Health{Count: len(backups), UsedSize: totalSize(backups)}
	if len(backups) == 0 {
		health.Problems = append(health.Problems, "there are no backups")
	} else {
		health.Newest = &backups[0]
		maximumAge := time.Duration(config.GetInt("backup.monitor.maximum_age_in_days", 1)) * 24 * time.Hour
		if age := now.Sub(backups[0].Time); age > maximumAge {
			health.Problems = append(health.Problems, fmt.Sprintf("the newest backup is %s old (allowed: %s)", age.Round(time.Minute), maximumAge))
		}
	}
	if limit := int64(config.GetInt("backup.monitor.maximum_storage_in_megabytes", 5000)); limit > 0 && health.UsedSize > limit*1024*1024 {
		health.Problems = append(health.Problems, fmt.Sprintf("the backups use %d MB (allowed: %d MB)", health.UsedSize/1024/1024, limit))
	}
	health.Healthy = len(health.Problems) == 0
	return health, nil
}

// Notify tells the people in BACKUP_NOTIFICATION_EMAIL that something is wrong (Spatie's mail notifications) and
// writes it to the log either way. It sends at once rather than through the queue: a broken queue must not hide
// a broken backup. Tests replace it.
var Notify = func(subject, body string) error {
	facades.Log().Error("backup: " + subject + ": " + body)
	to := facades.Config().GetString("backup.notification_email")
	if to == "" {
		return nil
	}
	appName := facades.Config().GetString("app.name")
	return facades.Mail().To([]string{to}).Subject("[" + appName + "] " + subject).
		Content(mail.Content{Text: body + "\n\nBackup folder: " + Folder() + "\n"}).Send()
}
