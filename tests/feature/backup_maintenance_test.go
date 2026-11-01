package feature

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/services/backup"
	"goravel/tests"
)

// BackupMaintenanceTestSuite covers backup:clean and backup:monitor on the local disk (no R2 needed): the retention
// settings, the newest backup that always stays, and the failure notices.
type BackupMaintenanceTestSuite struct {
	suite.Suite
	tests.TestCase
	notices  []string
	original map[string]any
	notify   func(subject, body string) error
}

// changedSettings are restored after each test: config is global, and the real backup test runs in this process.
var changedSettings = []string{"backup.disk", "backup.prefix", "backup.name", "backup.dump_binary",
	"backup.monitor.maximum_age_in_days", "backup.monitor.maximum_storage_in_megabytes"}

func TestBackupMaintenanceTestSuite(t *testing.T) {
	suite.Run(t, new(BackupMaintenanceTestSuite))
}

func (s *BackupMaintenanceTestSuite) SetupTest() {
	config := facades.Config()
	s.original = map[string]any{}
	for _, key := range changedSettings {
		s.original[key] = config.Get(key)
	}
	s.notify = backup.Notify
	config.Add("backup.disk", "local")
	config.Add("backup.prefix", "backup-maintenance-tests")
	config.Add("backup.name", "app-"+strconv.FormatInt(time.Now().UnixNano(), 10))
	config.Add("backup.monitor.maximum_age_in_days", 1)
	config.Add("backup.monitor.maximum_storage_in_megabytes", 5000)
	s.notices = nil
	backup.Notify = func(subject, body string) error {
		s.notices = append(s.notices, subject+": "+body)
		return nil
	}
}

func (s *BackupMaintenanceTestSuite) TearDownTest() {
	_ = facades.Storage().Disk("local").DeleteDirectory("backup-maintenance-tests")
	for key, value := range s.original {
		facades.Config().Add(key, value)
	}
	backup.Notify = s.notify
}

// store puts a fake backup taken `age` ago on the backup disk and returns its path.
func (s *BackupMaintenanceTestSuite) store(age time.Duration) string {
	path := backup.Folder() + "/" + time.Now().Add(-age).UTC().Format("2006-01-02-15-04-05") + ".sql.gz"
	s.Require().NoError(facades.Storage().Disk("local").Put(path, "fake dump"))
	return path
}

func (s *BackupMaintenanceTestSuite) TestCleanKeepsTheRecentOnesAndOneOfEachOlderDay() {
	day := 24 * time.Hour
	recent := []string{s.store(time.Hour), s.store(2 * time.Hour), s.store(3 * day)}
	newerOfTheDay := s.store(10*day + time.Hour)
	olderOfTheDay := s.store(10*day + 2*time.Hour)
	ancient := s.store(5 * 365 * day)
	notABackup := backup.Folder() + "/notes.sql.gz"
	s.Require().NoError(facades.Storage().Disk("local").Put(notABackup, "kept: its name is not a backup time"))

	s.Require().NoError(facades.Artisan().Call("backup:clean"))

	disk := facades.Storage().Disk("local")
	for _, path := range append(recent, newerOfTheDay, notABackup) {
		s.True(disk.Exists(path), "%s is kept", path)
	}
	s.False(disk.Exists(olderOfTheDay), "only the newest of an older day stays")
	s.False(disk.Exists(ancient), "past the yearly period it goes")
	s.Empty(s.notices)
}

func (s *BackupMaintenanceTestSuite) TestCleanFollowsTheRetentionSettings() {
	settings := map[string]int{"keep_all_backups_for_days": 0, "keep_daily_backups_for_days": 0, "keep_weekly_backups_for_weeks": 0,
		"keep_monthly_backups_for_months": 0, "keep_yearly_backups_for_years": 0}
	for key, value := range settings {
		facades.Config().Add("backup.cleanup."+key, value)
	}
	defer func() {
		defaults := backup.DefaultRetention
		facades.Config().Add("backup.cleanup.keep_all_backups_for_days", defaults.KeepAllBackupsForDays)
		facades.Config().Add("backup.cleanup.keep_daily_backups_for_days", defaults.KeepDailyBackupsForDays)
		facades.Config().Add("backup.cleanup.keep_weekly_backups_for_weeks", defaults.KeepWeeklyBackupsForWeeks)
		facades.Config().Add("backup.cleanup.keep_monthly_backups_for_months", defaults.KeepMonthlyBackupsForMonths)
		facades.Config().Add("backup.cleanup.keep_yearly_backups_for_years", defaults.KeepYearlyBackupsForYears)
	}()
	newest := s.store(time.Hour)
	older := s.store(2 * time.Hour)

	removed, kept, err := backup.Clean(time.Now())
	s.Require().NoError(err)
	s.Len(removed, 1)
	s.Len(kept, 1)
	s.True(facades.Storage().Disk("local").Exists(newest), "with nothing to keep, the newest still stays")
	s.False(facades.Storage().Disk("local").Exists(older))
}

func (s *BackupMaintenanceTestSuite) TestMonitorIsHealthyWithARecentBackup() {
	s.store(2 * time.Hour)
	s.Require().NoError(facades.Artisan().Call("backup:monitor"))
	health, err := backup.CheckHealth(time.Now())
	s.Require().NoError(err)
	s.True(health.Healthy)
	s.Equal(1, health.Count)
	s.Empty(s.notices)
}

func (s *BackupMaintenanceTestSuite) TestMonitorComplainsWithoutBackupsOrWithAnOldOne() {
	s.Error(facades.Artisan().Call("backup:monitor"), "no backups is unhealthy")
	s.Require().Len(s.notices, 1)
	s.Contains(s.notices[0], "there are no backups")

	s.store(30 * time.Hour)
	s.Error(facades.Artisan().Call("backup:monitor"))
	s.Require().Len(s.notices, 2)
	s.Contains(s.notices[1], "the newest backup is 30h0m0s old")
}

func (s *BackupMaintenanceTestSuite) TestMonitorComplainsWhenTheBackupsUseTooMuchSpace() {
	facades.Config().Add("backup.monitor.maximum_storage_in_megabytes", 1)
	path := backup.Folder() + "/" + time.Now().UTC().Format("2006-01-02-15-04-05") + ".sql.gz"
	s.Require().NoError(facades.Storage().Disk("local").Put(path, string(make([]byte, 2<<20))))
	health, _ := backup.CheckHealth(time.Now())
	s.False(health.Healthy)
	s.Contains(health.Problems[0], "the backups use 2 MB (allowed: 1 MB)")
}

func (s *BackupMaintenanceTestSuite) TestAFailedBackupRunSendsANotice() {
	facades.Config().Add("backup.dump_binary", "/no/such/dump-tool")
	s.Error(facades.Artisan().Call("backup:run"))
	s.Require().Len(s.notices, 1)
	s.Contains(s.notices[0], "The database backup failed")
}
