package backup

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"time"
)

// Backup is one stored backup: its path on the disk, when it was taken (from its name) and its size in bytes.
type Backup struct {
	Path string
	Time time.Time
	Size int64
}

// Retention is spatie/laravel-backup's default cleanup strategy, with the same setting names and defaults.
type Retention struct {
	KeepAllBackupsForDays       int // every backup younger than this is kept
	KeepDailyBackupsForDays     int // then the newest backup of each day
	KeepWeeklyBackupsForWeeks   int // then the newest of each week
	KeepMonthlyBackupsForMonths int // then the newest of each month
	KeepYearlyBackupsForYears   int // then the newest of each year; anything older goes
	// After the periods: delete the oldest until the rest fit. 0 means no limit.
	DeleteOldestBackupsWhenUsingMoreMegabytesThan int64
}

// DefaultRetention is Spatie's config/backup.php cleanup default.
var DefaultRetention = Retention{
	KeepAllBackupsForDays: 7, KeepDailyBackupsForDays: 16, KeepWeeklyBackupsForWeeks: 8,
	KeepMonthlyBackupsForMonths: 4, KeepYearlyBackupsForYears: 2, DeleteOldestBackupsWhenUsingMoreMegabytesThan: 5000,
}

const nameLayout = "2006-01-02-15-04-05"

// TimeFromPath reads the time a backup was taken from its name (2026-11-02-01-30-00.sql.gz, UTC).
func TimeFromPath(backupPath string) (time.Time, bool) {
	name := strings.TrimSuffix(path.Base(backupPath), ".sql.gz")
	taken, err := time.ParseInLocation(nameLayout, name, time.UTC)
	return taken, err == nil
}

// Plan decides which backups to keep and which to delete, like Spatie's DefaultStrategy:
//
//   - younger than KeepAllBackupsForDays: kept;
//   - then, period after period (daily, weekly, monthly, yearly), the newest backup of each day, week, month or
//     year is kept and the others in that group are deleted;
//   - older than the yearly period: deleted;
//   - then, while the kept ones use more than the size limit, the oldest is deleted.
//
// The newest backup is never deleted, whatever the settings say.
func Plan(backups []Backup, now time.Time, retention Retention) (keep, remove []Backup) {
	sorted := append([]Backup(nil), backups...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.After(sorted[j].Time) }) // newest first
	if len(sorted) == 0 {
		return nil, nil
	}

	allUntil := now.AddDate(0, 0, -retention.KeepAllBackupsForDays)
	dailyUntil := allUntil.AddDate(0, 0, -retention.KeepDailyBackupsForDays)
	weeklyUntil := dailyUntil.AddDate(0, 0, -7*retention.KeepWeeklyBackupsForWeeks)
	monthlyUntil := weeklyUntil.AddDate(0, -retention.KeepMonthlyBackupsForMonths, 0)
	yearlyUntil := monthlyUntil.AddDate(-retention.KeepYearlyBackupsForYears, 0, 0)

	seen := map[string]bool{}
	for index, item := range sorted {
		taken := item.Time.UTC()
		var group string
		switch {
		case index == 0 || !taken.Before(allUntil):
			keep = append(keep, item)
			continue
		case !taken.Before(dailyUntil):
			group = "day " + taken.Format("2006-01-02")
		case !taken.Before(weeklyUntil):
			year, week := taken.ISOWeek()
			group = fmt.Sprintf("week %d-%02d", year, week)
		case !taken.Before(monthlyUntil):
			group = "month " + taken.Format("2006-01")
		case !taken.Before(yearlyUntil):
			group = "year " + taken.Format("2006")
		default:
			remove = append(remove, item)
			continue
		}
		if seen[group] {
			remove = append(remove, item)
			continue
		}
		seen[group] = true
		keep = append(keep, item)
	}

	if limit := retention.DeleteOldestBackupsWhenUsingMoreMegabytesThan * 1024 * 1024; limit > 0 {
		for len(keep) > 1 && totalSize(keep) > limit {
			oldest := keep[len(keep)-1]
			keep = keep[:len(keep)-1]
			remove = append(remove, oldest)
		}
	}
	return keep, remove
}

func totalSize(backups []Backup) int64 {
	var total int64
	for _, item := range backups {
		total += item.Size
	}
	return total
}
