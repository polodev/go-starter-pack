package feature

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/services/backup"
	"goravel/tests"
)

// BackupTestSuite makes a real backup of the database on the real private R2 bucket, checks it and deletes it.
// Skipped without the backup keys or the dump tool.
type BackupTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestBackupTestSuite(t *testing.T) {
	suite.Run(t, new(BackupTestSuite))
}

// SetupTest gives the test database its tables, so the dump has something in it.
func (s *BackupTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *BackupTestSuite) TestABackupIsAGzippedDumpOnThePrivateDiskAndIsListed() {
	for _, name := range []string{"BACKUP_ACCESS_KEY_ID", "BACKUP_SECRET_ACCESS_KEY", "BACKUP_ENDPOINT", "BACKUP_BUCKET"} {
		if env(name) == "" {
			s.T().Skipf("%s is not set in .env", name)
		}
	}
	if exec.Command("mysqldump", "--version").Run() != nil && exec.Command("mariadb-dump", "--version").Run() != nil {
		s.T().Skip("mysqldump is not installed")
	}

	// A time far in the past keeps the name apart from real backups made today.
	when := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	stored, err := backup.Run(when)
	s.Require().NoError(err)
	s.Equal("r2_backups", stored.Disk)
	s.Equal(backup.Folder()+"/2001-02-03-04-05-06.sql.gz", stored.Path)
	s.Greater(stored.Size, int64(100))
	defer func() { _ = facades.Storage().Disk("r2_backups").Delete(stored.Path) }()

	listed, err := backup.List()
	s.Require().NoError(err)
	s.Contains(listed, stored.Path)

	sql, err := backup.Read(stored.Path)
	s.Require().NoError(err)
	s.True(strings.Contains(sql, "CREATE TABLE `jobs`") && strings.Contains(sql, "CREATE TABLE `failed_jobs`"),
		"a real dump of the app's tables")

	s.Require().NoError(facades.Storage().Disk("r2_backups").Delete(stored.Path))
	listed, err = backup.List()
	s.Require().NoError(err)
	s.NotContains(listed, stored.Path)
}
