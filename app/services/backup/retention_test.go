package backup

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)

func taken(daysAgo int, hour int) Backup {
	day := now.AddDate(0, 0, -daysAgo)
	when := time.Date(day.Year(), day.Month(), day.Day(), hour, 30, 0, 0, time.UTC)
	return Backup{Path: "db-backups/app/" + when.Format(nameLayout) + ".sql.gz", Time: when, Size: 1024 * 1024}
}

func paths(backups []Backup) map[string]bool {
	found := map[string]bool{}
	for _, item := range backups {
		found[item.Path] = true
	}
	return found
}

func TestTimeFromPathReadsTheNameRunGives(t *testing.T) {
	when, ok := TimeFromPath("db-backups/app/2026-11-02-01-30-00.sql.gz")
	assert.True(t, ok)
	assert.Equal(t, time.Date(2026, 11, 2, 1, 30, 0, 0, time.UTC), when)
	_, ok = TimeFromPath("db-backups/app/notes.sql.gz")
	assert.False(t, ok, "a file that is not a backup has no time")
}

func TestEveryBackupOfTheLastWeekIsKept(t *testing.T) {
	backups := []Backup{taken(0, 1), taken(0, 13), taken(3, 1), taken(3, 13), taken(6, 1)}
	keep, remove := Plan(backups, now, DefaultRetention)
	assert.Len(t, keep, 5)
	assert.Empty(t, remove)
}

func TestOlderBackupsThinOutToOnePerDayWeekMonthAndYear(t *testing.T) {
	daily := []Backup{taken(10, 1), taken(10, 13)}    // same day, in the daily period
	weekly := []Backup{taken(30, 1), taken(31, 1)}    // 2026-10-03 and 10-02: same ISO week (40)
	monthly := []Backup{taken(100, 1), taken(105, 1)} // both in July 2026, in the monthly period
	yearly := []Backup{taken(400, 1), taken(420, 1)}  // both in 2025, in the yearly period
	tooOld := []Backup{taken(1500, 1)}                // 2022: past every period
	backups := append(append(append(append(append([]Backup{taken(0, 1)}, daily...), weekly...), monthly...), yearly...), tooOld...)

	keep, remove := Plan(backups, now, DefaultRetention)
	kept, removed := paths(keep), paths(remove)

	assert.True(t, kept[daily[1].Path] && removed[daily[0].Path], "the newest of the day stays")
	assert.True(t, kept[weekly[0].Path] && removed[weekly[1].Path], "the newest of the week stays")
	assert.True(t, kept[monthly[0].Path] && removed[monthly[1].Path], "the newest of the month stays")
	assert.True(t, kept[yearly[0].Path] && removed[yearly[1].Path], "the newest of the year stays")
	assert.True(t, removed[tooOld[0].Path], "past the yearly period it goes")
	assert.Len(t, keep, 5)
	assert.Len(t, remove, 5)
}

func TestAYearOfDailyBackupsKeepsAboutThirtyFive(t *testing.T) {
	var backups []Backup
	for day := 0; day < 365; day++ {
		backups = append(backups, taken(day, 1))
	}
	keep, remove := Plan(backups, now, DefaultRetention)
	assert.Equal(t, 365, len(keep)+len(remove), "every backup is either kept or deleted")
	// 7 (all) + 16 (daily) + 8 or 9 weekly + 4 or 5 monthly + 1 or 2 yearly.
	assert.GreaterOrEqual(t, len(keep), 36)
	assert.LessOrEqual(t, len(keep), 39)
}

func TestTheNewestBackupIsNeverDeleted(t *testing.T) {
	old := []Backup{taken(2000, 1), taken(3000, 1)}
	keep, remove := Plan(old, now, DefaultRetention)
	assert.Equal(t, []Backup{old[0]}, keep, "even years old, the newest stays")
	assert.Equal(t, []Backup{old[1]}, remove)

	big := taken(0, 1)
	big.Size = 10 << 30
	keep, remove = Plan([]Backup{big}, now, DefaultRetention)
	assert.Equal(t, []Backup{big}, keep, "even over the size limit, the newest stays")
	assert.Empty(t, remove)

	keep, remove = Plan(nil, now, DefaultRetention)
	assert.Empty(t, keep)
	assert.Empty(t, remove)
}

func TestTheOldestGoFirstWhenTheBackupsUseTooMuchSpace(t *testing.T) {
	backups := []Backup{taken(0, 1), taken(1, 1), taken(2, 1), taken(3, 1)} // 1 MB each, all in the "keep all" week
	limited := DefaultRetention
	limited.DeleteOldestBackupsWhenUsingMoreMegabytesThan = 2
	keep, remove := Plan(backups, now, limited)
	assert.Equal(t, []Backup{backups[0], backups[1]}, keep)
	assert.Equal(t, []Backup{backups[3], backups[2]}, remove, "oldest first")
}
