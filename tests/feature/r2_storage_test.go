package feature

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// R2StorageTestSuite talks to the real Cloudflare R2 buckets from .env. It is skipped when the keys are missing,
// and it removes every file it writes.
type R2StorageTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestR2StorageTestSuite(t *testing.T) {
	suite.Run(t, new(R2StorageTestSuite))
}

func env(name string) string {
	value, _ := facades.Config().Env(name, "").(string)
	return value
}

func (s *R2StorageTestSuite) requireKeys(names ...string) {
	for _, name := range names {
		if env(name) == "" {
			s.T().Skipf("%s is not set in .env", name)
		}
	}
}

func testObjectPath(kind string) string {
	root := env("R2_ROOT")
	if root == "" {
		root = "go-starter-pack"
	}
	return fmt.Sprintf("%s/tests/%s-%d.txt", root, kind, time.Now().UnixNano())
}

func (s *R2StorageTestSuite) TestAPublicFileIsStoredReadServedByItsPublicUrlAndDeleted() {
	s.requireKeys("R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_ENDPOINT", "R2_PUBLIC_BUCKET")
	disk := facades.Storage().Disk("r2_public")
	name := testObjectPath("public")
	content := "hello from the Go starter pack"

	s.Require().NoError(disk.Put(name, content))
	s.True(disk.Exists(name))
	stored, err := disk.Get(name)
	s.Require().NoError(err)
	s.Equal(content, stored)

	if publicURL := env("R2_PUBLIC_URL"); publicURL != "" {
		url := disk.Url(name)
		s.T().Logf("public address: %s", url)
		s.True(strings.HasPrefix(url, strings.TrimSuffix(publicURL, "/")), "public address %s", url)
		response, err := http.Get(url)
		s.Require().NoError(err)
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		s.Equal(http.StatusOK, response.StatusCode, "anyone can read a public file")
		s.Equal(content, string(body))
	}

	s.Require().NoError(disk.Delete(name))
	s.False(disk.Exists(name))
}

func (s *R2StorageTestSuite) TestABackupGoesToThePrivateBucketWhichIsNotPubliclyReadable() {
	s.requireKeys("BACKUP_ACCESS_KEY_ID", "BACKUP_SECRET_ACCESS_KEY", "BACKUP_ENDPOINT", "BACKUP_BUCKET")
	disk := facades.Storage().Disk("r2_backups")
	name := testObjectPath("private")

	s.Require().NoError(disk.Put(name, "secret backup bytes"))
	stored, err := disk.Get(name)
	s.Require().NoError(err)
	s.Equal("secret backup bytes", stored)

	// The private bucket must not answer an anonymous request.
	endpoint := strings.TrimSuffix(env("BACKUP_ENDPOINT"), "/")
	anonymous := fmt.Sprintf("%s/%s/%s", endpoint, env("BACKUP_BUCKET"), name)
	response, err := http.Get(anonymous)
	s.Require().NoError(err)
	_ = response.Body.Close()
	s.NotEqual(http.StatusOK, response.StatusCode, "an anonymous request is refused")

	s.Require().NoError(disk.Delete(name))
	s.False(disk.Exists(name))
}
