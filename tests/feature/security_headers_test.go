package feature

import (
	"net/url"
	"os"
	"regexp"
	"strconv"
	"testing"

	contractstesting "github.com/goravel/framework/contracts/testing/http"
	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/config"
	"goravel/tests"
)

// SecurityHeadersTestSuite covers the global browser security headers and the rule that makes the strict
// Content-Security-Policy possible: no page carries an inline script or an inline event handler.
type SecurityHeadersTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestSecurityHeadersTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityHeadersTestSuite))
}

func (s *SecurityHeadersTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	auth.SendMail = func(auth.Outgoing) error { return nil }
	facades.Config().Add("security.content_security_policy", config.DefaultContentSecurityPolicy)
	facades.Config().Add("security.strict_transport_security", false)
}

func (s *SecurityHeadersTestSuite) assertSecurityHeaders(response contractstesting.Response) {
	response.AssertHeader("Content-Security-Policy", config.DefaultContentSecurityPolicy)
	response.AssertHeader("X-Content-Type-Options", "nosniff")
	response.AssertHeader("X-Frame-Options", "DENY")
	response.AssertHeader("Referrer-Policy", "strict-origin-when-cross-origin")
	response.AssertHeader("Cross-Origin-Opener-Policy", "same-origin")
}

func (s *SecurityHeadersTestSuite) TestPagesApiAnswersStaticFilesAndNotFoundAllCarryTheHeaders() {
	for _, uri := range []string{"/", "/login", "/api/v1/auth/me", "/public/js/theme.js", "/public/css/app.css", "/no-such-page"} {
		response, err := s.Http(s.T()).Get(uri)
		s.Require().NoError(err, uri)
		s.assertSecurityHeaders(response)
		response.AssertHeaderMissing("Strict-Transport-Security")
	}
	// Tests run in tests/feature, where ./public does not resolve, so check the layouts' script on disk.
	_, err := os.Stat("../../public/js/theme.js")
	s.NoError(err, "the theme switch script exists")
}

func (s *SecurityHeadersTestSuite) TestHstsIsSentOnlyWhenTurnedOn() {
	facades.Config().Add("security.strict_transport_security", true)
	facades.Config().Add("security.strict_transport_security_max_age", 31536000)
	response, _ := s.Http(s.T()).Get("/")
	response.AssertHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
}

func (s *SecurityHeadersTestSuite) TestAnEmptyPolicySendsNoContentSecurityPolicy() {
	facades.Config().Add("security.content_security_policy", "")
	response, _ := s.Http(s.T()).Get("/")
	response.AssertHeaderMissing("Content-Security-Policy")
	response.AssertHeader("X-Frame-Options", "DENY")
}

var (
	inlineScript       = regexp.MustCompile(`<script(\s[^>]*)?>\s*[^<\s]`)
	inlineEventHandler = regexp.MustCompile(`\son[a-z]+\s*=`)
	scriptWithoutSrc   = regexp.MustCompile(`<script(?:\s+(?:type|nonce)="[^"]*")*\s*>`)
	inlineStyle        = regexp.MustCompile(`\sstyle\s*=|<style[\s>]`)
)

// The policy allows scripts and styles from this site only, so an inline script, onclick or style="..." would
// silently stop working.
func (s *SecurityHeadersTestSuite) TestNoPageHasInlineScriptsOrInlineEventHandlers() {
	user, _, err := auth.CreateUser("Developer", "developer@example.com", "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	_, err = rolepermission.Sync()
	s.Require().NoError(err)
	s.Require().NoError(rolepermission.AssignRole(user.ID, "developer"))

	guest := newBrowser(s.T(), &s.TestCase)
	pages := []string{"/", "/?lang=bn", "/login", "/register", "/forgot-password", "/contact"}
	for _, uri := range pages {
		s.assertNoInlineCode(guest, uri)
	}

	staff := newBrowser(s.T(), &s.TestCase)
	staff.get("/login")
	response, _ := staff.post("/login", url.Values{"email": {"developer@example.com"}, "password": {"secret-password"}})
	response.AssertStatus(303)
	for _, uri := range []string{"/", "/account/profile", "/account/avatar", "/admin", "/admin/users", "/admin/users/" + strconv.Itoa(int(user.ID)) + "/edit", "/admin/role-permissions", "/admin/role-permissions/matrix", "/admin/activity-logs", "/admin/media", "/admin/pages", "/admin/pages/create", "/admin/ai-api-logs",
		"/admin/queue-watcher", "/admin/queue-watcher/jobs", "/admin/queue-watcher/pending", "/admin/queue-watcher/failed",
		"/admin/contacts"} {
		s.assertNoInlineCode(staff, uri)
	}
}

func (s *SecurityHeadersTestSuite) assertNoInlineCode(b *browser, uri string) {
	response, body := b.get(uri)
	response.AssertStatus(200)
	s.NotRegexp(inlineScript, body, "%s has an inline script", uri)
	s.NotRegexp(scriptWithoutSrc, body, "%s has a script without src", uri)
	s.NotRegexp(inlineEventHandler, body, "%s has an inline event handler", uri)
	s.NotRegexp(inlineStyle, body, "%s has an inline style (the policy allows styles from this site only)", uri)
	s.Contains(body, `<script src="/public/js/theme.js"></script>`, uri)
}
