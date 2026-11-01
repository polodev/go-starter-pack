package feature

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

// AdminTestSuite covers roles and permissions and the admin screens: who may open what, role assignment,
// direct exceptions, the roles panel and fail-closed admin routes.
type AdminTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestAdminTestSuite(t *testing.T) {
	suite.Run(t, new(AdminTestSuite))
}

func (s *AdminTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
}

// person makes a verified account with a role ("" = none) and returns it.
func (s *AdminTestSuite) person(email, role string) models.User {
	user, _, err := auth.CreateUser("Person "+role, email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	if role != "" {
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	return user
}

// signedIn is a browser signed in as email.
func (s *AdminTestSuite) signedIn(email string) *browser {
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	response, _ := b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	response.AssertStatus(303)
	return b
}

func (s *AdminTestSuite) TestTheAdminAreaNeedsASignedInVerifiedStaffMember() {
	guest := newBrowser(s.T(), &s.TestCase)
	response, _ := guest.get("/admin")
	s.True(strings.HasPrefix(location(response), "/login?next="), "a guest is sent to log in")

	user, _, _ := auth.CreateUser("Unverified", "unverified@example.com", "secret-password")
	s.Require().NoError(rolepermission.AssignRole(user.ID, "admin"))
	response, _ = s.signedIn("unverified@example.com").get("/admin")
	s.Equal("/email/verify", location(response), "an unverified email cannot open the admin")

	s.person("customer@example.com", "")
	response, body := s.signedIn("customer@example.com").get("/admin")
	response.AssertStatus(403)
	s.Contains(body, `data-testid="error-title"`)
}

func (s *AdminTestSuite) TestEachRoleSeesOnlyWhatItMay() {
	s.person("new@example.com", "new_employee")
	s.person("bd@example.com", "business_development")
	target := s.person("target@example.com", "")
	userPage := "/admin/users/" + strconv.Itoa(int(target.ID))

	newEmployee := s.signedIn("new@example.com")
	dashboard, body := newEmployee.get("/admin")
	dashboard.AssertStatus(200)
	s.Contains(body, `data-testid="stat-users"`)
	s.NotContains(body, `href="/admin/users"`, "the sidebar hides what they may not open")
	forbidden, _ := newEmployee.get("/admin/users")
	forbidden.AssertStatus(403)

	bd := s.signedIn("bd@example.com")
	list, body := bd.get("/admin/users")
	list.AssertStatus(200)
	s.Contains(body, `data-testid="id-link"`)
	s.NotContains(body, `data-testid="edit-link"`, "no Edit link without users.manage")
	show, _ := bd.get(userPage)
	show.AssertStatus(200)
	edit, _ := bd.get(userPage + "/edit")
	edit.AssertStatus(403)
	roles, _ := bd.get("/admin/role-permissions")
	roles.AssertStatus(403)
}

func (s *AdminTestSuite) TestAnAdminEditsAUserButCannotAssignRoles() {
	s.person("admin@example.com", "admin")
	target := s.person("target@example.com", "")
	editPage := "/admin/users/" + strconv.Itoa(int(target.ID)) + "/edit"

	b := s.signedIn("admin@example.com")
	_, form := b.get(editPage)
	s.NotContains(form, `data-testid="role-fieldset"`, "only roles.assign sees the role section")
	bad, body := b.post("/admin/users/"+strconv.Itoa(int(target.ID)), url.Values{"name": {""}, "email": {"not-an-email"}})
	bad.AssertStatus(422)
	s.Contains(body, `data-testid="error-name"`)
	s.Contains(body, `data-testid="error-email"`)

	_, _ = b.get(editPage)
	saved, _ := b.post("/admin/users/"+strconv.Itoa(int(target.ID)), url.Values{"name": {"Renamed"}, "email": {"target@example.com"}, "role": {"developer"}})
	s.Equal("/admin/users/"+strconv.Itoa(int(target.ID))+"?saved=user", location(saved))
	var stored models.User
	s.Require().NoError(facades.Orm().Query().Find(&stored, target.ID))
	s.Equal("Renamed", stored.Name)
	s.Equal("", rolepermission.RoleSlug(target.ID), "an admin's role field is ignored")

	panel, _ := b.get("/admin/role-permissions")
	panel.AssertStatus(403)
}

func (s *AdminTestSuite) TestADeveloperAssignsRolesAndExceptionsButNeverRolesStar() {
	developer := s.person("dev@example.com", "developer")
	target := s.person("target@example.com", "")
	id := strconv.Itoa(int(target.ID))

	b := s.signedIn("dev@example.com")
	_, form := b.get("/admin/users/" + id + "/edit")
	s.Contains(form, `data-testid="role-fieldset"`)
	saved, _ := b.post("/admin/users/"+id, url.Values{"name": {"Target"}, "email": {"target@example.com"},
		"role": {"employee"}, "exceptions[]": {"users.manage"}})
	s.Equal("/admin/users/"+id+"?saved=user", location(saved))
	s.Equal("employee", rolepermission.RoleSlug(target.ID))
	s.True(rolepermission.HasPermission(target.ID, "users.manage"), "the exception is granted")
	s.True(rolepermission.HasPermission(target.ID, "pages.manage"), "from the employee role")
	s.False(rolepermission.HasPermission(target.ID, "media.manage"))

	s.Error(rolepermission.SetExceptions(target.ID, []string{"roles.assign"}), "roles.* never comes as an exception")
	s.False(rolepermission.HasPermission(target.ID, "roles.assign"))

	_, _ = b.get("/admin/users/" + strconv.Itoa(int(developer.ID)) + "/edit")
	last, body := b.post("/admin/users/"+strconv.Itoa(int(developer.ID)), url.Values{"name": {"Dev"}, "email": {"dev@example.com"}, "role": {"admin"}})
	last.AssertStatus(422)
	s.Contains(body, `data-testid="error-role"`, "the last developer cannot be demoted")
	s.Equal("developer", rolepermission.RoleSlug(developer.ID))
}

func (s *AdminTestSuite) TestTheRolesPanelAndTheSyncButton() {
	s.person("dev@example.com", "developer")
	b := s.signedIn("dev@example.com")
	for _, page := range []string{"", "/roles", "/roles/admin", "/permissions", "/matrix", "/users", "/links"} {
		response, _ := b.get("/admin/role-permissions" + page)
		response.AssertStatus(200)
	}
	_, overview := b.get("/admin/role-permissions")
	s.Regexp(`data-testid="role-count">\s*5\s*<`, overview)
	_, links := b.get("/admin/role-permissions/links")
	s.Contains(links, "GET /admin/users/{id}/edit")
	for _, route := range rolepermission.Unmapped() {
		// The audit test registers one undeclared route on purpose (routes cannot be removed again).
		s.Equal("GET /admin/undeclared-test-route", route, "every real admin route has a permission")
	}

	_, _ = b.get("/admin/role-permissions")
	synced, _ := b.post("/admin/role-permissions/sync", nil)
	s.Equal("/admin/role-permissions?saved=sync", location(synced))
	missing, _ := b.get("/admin/role-permissions/roles/nobody")
	missing.AssertStatus(404)
}

func (s *AdminTestSuite) TestPermissionRules() {
	developer := s.person("dev@example.com", "developer")
	s.True(rolepermission.HasPermission(developer.ID, "anything.at_all"), "a developer holds *")

	admin := s.person("admin@example.com", "admin")
	s.True(rolepermission.HasPermission(admin.ID, "queue.manage"), "queue.* covers queue.manage")
	s.False(rolepermission.HasPermission(admin.ID, "roles.view"))

	plain := s.person("plain@example.com", "")
	s.False(rolepermission.HasPermission(plain.ID, "dashboard.view"))
	s.Require().NoError(rolepermission.SetExceptions(plain.ID, []string{"pages.view"}))
	s.True(rolepermission.HasPermission(plain.ID, "dashboard.view"), "any grant opens the dashboard")

	_, err := facades.Orm().Query().Exec("UPDATE roles SET is_active = ? WHERE slug = ?", false, "admin")
	s.Require().NoError(err)
	rolepermission.Forget(admin.ID)
	s.False(rolepermission.HasPermission(admin.ID, "users.manage"), "a switched-off role grants nothing")

	s.T().Setenv("ROLE_PERMISSION_SUPER_USER_IDS", strconv.Itoa(int(plain.ID)))
	s.True(rolepermission.HasPermission(plain.ID, "roles.sync"), "a super user passes every check")
}

func (s *AdminTestSuite) TestAnUndeclaredAdminRouteIsFoundByTheAudit() {
	facades.Route().Get("/admin/undeclared-test-route", func(ctx http.Context) http.Response {
		return ctx.Response().String(http.StatusOK, "should not exist")
	})
	s.Contains(rolepermission.Unmapped(), "GET /admin/undeclared-test-route")
	s.NotContains(rolepermission.Unmapped(), "GET /admin/users")
}
