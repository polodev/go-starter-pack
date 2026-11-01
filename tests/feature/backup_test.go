package feature

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync/atomic"
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

// TestALargeBackupStreamsWithoutHoldingTheDumpInMemory backs up about 80 MB of data that compresses badly, so the
// upload has several parts, and samples the heap: it must stay far below the dump's size.
func (s *BackupTestSuite) TestALargeBackupStreamsWithoutHoldingTheDumpInMemory() {
	for _, name := range []string{"BACKUP_ACCESS_KEY_ID", "BACKUP_SECRET_ACCESS_KEY", "BACKUP_ENDPOINT", "BACKUP_BUCKET"} {
		if env(name) == "" {
			s.T().Skipf("%s is not set in .env", name)
		}
	}
	if exec.Command("mysqldump", "--version").Run() != nil && exec.Command("mariadb-dump", "--version").Run() != nil {
		s.T().Skip("mysqldump is not installed")
	}
	if testing.Short() {
		s.T().Skip("slow: skipped with -short")
	}

	db := facades.Orm().Query()
	_, err := db.Exec("CREATE TABLE backup_load_test (id INT AUTO_INCREMENT PRIMARY KEY, payload TEXT NOT NULL)")
	s.Require().NoError(err)
	defer func() { _, _ = facades.Orm().Query().Exec("DROP TABLE IF EXISTS backup_load_test") }()
	// 640,000 rows of 128 random hex characters: about 82 MB that gzip only halves.
	// MariaDB stops recursion after 1,000 steps, so the rows are 1,000 x 640.
	_, err = db.Exec(`INSERT INTO backup_load_test (payload)
		WITH RECURSIVE a(n) AS (SELECT 1 UNION ALL SELECT n + 1 FROM a WHERE n < 1000),
		b(m) AS (SELECT 1 UNION ALL SELECT m + 1 FROM b WHERE m < 640)
		SELECT SHA2(CONCAT(RAND(), n, '-', m), 512) FROM a, b`)
	s.Require().NoError(err)

	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	var peak atomic.Uint64
	stop := make(chan struct{})
	sampled := make(chan struct{})
	go func() {
		defer close(sampled)
		var stats runtime.MemStats
		for {
			select {
			case <-stop:
				return
			case <-time.After(20 * time.Millisecond):
				runtime.ReadMemStats(&stats)
				if stats.HeapInuse > peak.Load() {
					peak.Store(stats.HeapInuse)
				}
			}
		}
	}()

	when := time.Date(2001, 2, 3, 4, 5, 7, 0, time.UTC)
	stored, err := backup.Run(when)
	close(stop)
	<-sampled
	s.Require().NoError(err)
	defer func() { _ = facades.Storage().Disk("r2_backups").Delete(stored.Path) }()

	growth := int64(peak.Load()) - int64(before.HeapInuse)
	s.T().Logf("backup %.1f MB, heap grew by at most %.1f MB", float64(stored.Size)/1e6, float64(growth)/1e6)
	s.Greater(stored.Size, int64(2*backup.PartSize), "big enough for a multipart upload of three parts")
	s.Less(growth, int64(48<<20), "the dump is streamed, never held whole")

	size, err := facades.Storage().Disk("r2_backups").Size(stored.Path)
	s.Require().NoError(err)
	s.Equal(stored.Size, size, "the whole file arrived")
	entries, _ := os.ReadDir(backup.TemporaryDirectory())
	s.Empty(entries, "the temporary file is removed")
}
