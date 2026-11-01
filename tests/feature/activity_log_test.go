package feature

import (
	"encoding/json"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	activitylog "goravel/app-modules/activity-log"
	"goravel/app-modules/activity-log/models"
	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/activity"
	"goravel/app/facades"
	usermodels "goravel/app/models"
	"goravel/tests"
)

// ActivityLogTestSuite covers the Spatie-compatible activity log: entries written by admin changes, the screens,
// permissions, delete and clean.
type ActivityLogTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestActivityLogTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityLogTestSuite))
}

func (s *ActivityLogTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
}

func (s *ActivityLogTestSuite) person(email, role string) usermodels.User {
	user, _, err := auth.CreateUser("Person", email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	if role != "" {
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	return user
}

func (s *ActivityLogTestSuite) signedIn(email string) *browser {
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	response, _ := b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	response.AssertStatus(303)
	return b
}

func (s *ActivityLogTestSuite) TestEditingAUserWritesASpatieShapedEntryShownOnTheScreens() {
	admin := s.person("admin@example.com", "admin")
	target := s.person("target@example.com", "")
	b := s.signedIn("admin@example.com")
	b.get("/admin/users/" + strconv.Itoa(int(target.ID)) + "/edit")
	b.post("/admin/users/"+strconv.Itoa(int(target.ID)), url.Values{"name": {"New Name"}, "email": {"target@example.com"}})

	var entry models.Activity
	s.Require().NoError(facades.Orm().Query().Where("log_name", "users").First(&entry))
	s.Equal("updated user", entry.Description)
	s.Equal("updated", models.Value(entry.Event))
	s.Equal("user", models.Value(entry.SubjectType))
	s.Equal(strconv.Itoa(int(target.ID)), models.Value(entry.SubjectID))
	s.Equal(strconv.Itoa(int(admin.ID)), models.Value(entry.CauserID))
	var properties map[string]map[string]any
	s.Require().NoError(json.Unmarshal([]byte(models.Value(entry.Properties)), &properties))
	s.Equal("New Name", properties["attributes"]["name"])
	s.Equal("Person", properties["old"]["name"])
	_, emailChanged := properties["attributes"]["email"]
	s.False(emailChanged, "only what changed is recorded")

	list, body := b.get("/admin/activity-logs")
	list.AssertStatus(200)
	s.Contains(body, "updated user")
	_, filtered := b.get("/admin/activity-logs?log=roles")
	s.NotContains(filtered, `data-testid="activity-row"`, "the log filter applies")
	show, detail := b.get("/admin/activity-logs/" + strconv.Itoa(int(entry.ID)))
	show.AssertStatus(200)
	s.Contains(detail, `data-testid="changes"`)
	s.Contains(detail, "New Name")
	missing, _ := b.get("/admin/activity-logs/999999")
	missing.AssertStatus(404)
}

func (s *ActivityLogTestSuite) TestReadingAndDeletingNeedTheirPermissions() {
	s.person("bd@example.com", "business_development")
	s.person("employee@example.com", "employee")
	s.person("admin@example.com", "admin")
	activity.In("users").Event("created").LogQuietly("an entry")
	var entry models.Activity
	s.Require().NoError(facades.Orm().Query().First(&entry))
	id := strconv.Itoa(int(entry.ID))

	forbidden, _ := s.signedIn("bd@example.com").get("/admin/activity-logs")
	forbidden.AssertStatus(403)

	employee := s.signedIn("employee@example.com")
	_, page := employee.get("/admin/activity-logs/" + id)
	s.NotContains(page, `data-testid="delete-button"`, "view only")
	refused, _ := employee.post("/admin/activity-logs/"+id+"/delete", nil)
	refused.AssertStatus(403)
	count, _ := facades.Orm().Query().Model(&models.Activity{}).Count()
	s.Equal(int64(1), count, "nothing deleted after a refusal")

	admin := s.signedIn("admin@example.com")
	admin.get("/admin/activity-logs/" + id)
	deleted, _ := admin.post("/admin/activity-logs/"+id+"/delete", nil)
	s.Equal("/admin/activity-logs?saved=deleted", location(deleted))
	count, _ = facades.Orm().Query().Model(&models.Activity{}).Count()
	s.Equal(int64(0), count)
}

func (s *ActivityLogTestSuite) TestCleaningRemovesOnlyOldEntries() {
	activity.In("users").LogQuietly("recent")
	activity.In("users").LogQuietly("old")
	_, err := facades.Orm().Query().Exec("UPDATE activity_log SET created_at = ? WHERE description = ?",
		time.Now().UTC().AddDate(0, 0, -400), "old")
	s.Require().NoError(err)
	removed, err := activitylog.CleanOlderThan(365)
	s.Require().NoError(err)
	s.Equal(int64(1), removed)
	var left []models.Activity
	s.Require().NoError(facades.Orm().Query().Find(&left))
	s.Len(left, 1)
	s.Equal("recent", left[0].Description)
}
