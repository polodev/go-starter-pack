package feature

import (
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

// WebAuthTestSuite covers the web pages: register, verify, login, logout, forgot and reset password.
type WebAuthTestSuite struct {
	suite.Suite
	tests.TestCase
	mails []auth.Outgoing
}

// realSendMail is Goravel's mail, kept before SetupTest replaces it with a capture.
var realSendMail = auth.SendMail

func TestWebAuthTestSuite(t *testing.T) {
	suite.Run(t, new(WebAuthTestSuite))
}

func (s *WebAuthTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush() // rate limits
	s.mails = nil
	auth.SendMail = func(message auth.Outgoing) error {
		s.mails = append(s.mails, message)
		return nil
	}
}

var linkInMail = regexp.MustCompile(`https?://[^/\s"]+(/[^\s"<]+)`)

func (s *WebAuthTestSuite) lastLink() string {
	s.Require().NotEmpty(s.mails, "a mail was sent")
	match := linkInMail.FindStringSubmatch(s.mails[len(s.mails)-1].Text)
	s.Require().NotNil(match, "the mail has a link")
	return match[1]
}

func (s *WebAuthTestSuite) register(b *browser, email string) {
	b.get("/register")
	response, _ := b.post("/register", url.Values{"name": {"Polo Dev"}, "email": {email},
		"password": {"secret-password"}, "password_confirmation": {"secret-password"}})
	response.AssertStatus(303)
	s.Equal("/email/verify?status=registered", location(response))
}

func (s *WebAuthTestSuite) TestRegisteringSignsInEmailsAVerificationLinkAndTheLinkVerifies() {
	b := newBrowser(s.T(), &s.TestCase)
	s.register(b, "Polo@Example.com")

	_, body := b.get("/email/verify")
	s.Contains(body, `data-testid="verify-notice"`)
	s.Contains(body, "polo@example.com")
	s.Len(s.mails, 1)
	s.Equal("polo@example.com", s.mails[0].To)
	s.Equal("Verify your email address", s.mails[0].Subject)
	s.Contains(s.mails[0].HTML, "/email/verify/")

	_, profile := b.get("/account/profile")
	s.Contains(profile, `data-testid="email-not-verified"`)

	link := s.lastLink()
	verified, _ := b.get(link)
	verified.AssertStatus(303)
	s.Equal("/account/profile?status=verified", location(verified))
	_, profile = b.get("/account/profile")
	s.Contains(profile, `data-testid="email-verified"`)

	again, _ := b.get(link)
	s.Equal("/login?status=link-invalid", location(again), "a link works once")
}

func (s *WebAuthTestSuite) TestRegisterKeepsTypedValuesOnErrorsAndRefusesATakenEmail() {
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/register")
	response, body := b.post("/register", url.Values{"name": {""}, "email": {"not-an-email"},
		"password": {"short"}, "password_confirmation": {"different"}})
	response.AssertStatus(422)
	s.Contains(body, `data-testid="error-name"`)
	s.Contains(body, `data-testid="error-email"`)
	s.Contains(body, `data-testid="error-password"`)
	s.Contains(body, `data-testid="error-password-confirmation"`)
	s.Contains(body, `value="not-an-email"`, "what was typed is kept")
	count, _ := facades.Orm().Query().Model(&models.User{}).Count()
	s.Equal(int64(0), count, "nothing stored after a refusal")

	s.register(newBrowser(s.T(), &s.TestCase), "taken@example.com")
	other := newBrowser(s.T(), &s.TestCase)
	other.get("/register")
	response, body = other.post("/register", url.Values{"name": {"X"}, "email": {"TAKEN@example.com"},
		"password": {"secret-password"}, "password_confirmation": {"secret-password"}})
	response.AssertStatus(422)
	s.Contains(body, "already has an account")
}

func (s *WebAuthTestSuite) TestLoginLogoutAndGuestRedirects() {
	s.register(newBrowser(s.T(), &s.TestCase), "login@example.com")

	b := newBrowser(s.T(), &s.TestCase)
	guestProfile, _ := b.get("/account/profile")
	guestProfile.AssertStatus(303)
	s.True(strings.HasPrefix(location(guestProfile), "/login?next="), "a guest is sent to log in")

	b.get("/login")
	wrong, body := b.post("/login", url.Values{"email": {"login@example.com"}, "password": {"nope-nope"}})
	wrong.AssertStatus(422)
	s.Contains(body, "do not match our records")
	unknown, unknownBody := b.post("/login", url.Values{"email": {"nobody@example.com"}, "password": {"secret-password"}})
	unknown.AssertStatus(422)
	s.Contains(unknownBody, "do not match our records", "unknown email and wrong password look the same")

	ok, _ := b.post("/login", url.Values{"email": {"LOGIN@example.com"}, "password": {"secret-password"}, "next": {"/account/profile"}})
	ok.AssertStatus(303)
	s.Equal("/account/profile", location(ok))
	_, profile := b.get("/account/profile")
	s.Contains(profile, `data-testid="profile-email">login@example.com`)
	_, home := b.get("/")
	s.Contains(home, `data-testid="nav-logout"`, "the header knows who is signed in")

	signedInLogin, _ := b.get("/login")
	s.Equal("/account/profile", location(signedInLogin), "a signed-in person does not see the login page")

	out, _ := b.post("/logout", nil)
	s.Equal("/login?status=logged-out", location(out))
	after, _ := b.get("/account/profile")
	after.AssertStatus(303)
}

func (s *WebAuthTestSuite) TestAFormWithoutTheCsrfTokenIsRefused() {
	s.register(newBrowser(s.T(), &s.TestCase), "csrf@example.com")
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	forged, _ := b.postWithToken("/login", url.Values{"email": {"csrf@example.com"}, "password": {"secret-password"}}, "forged-token")
	forged.AssertStatus(419)
	missing, _ := b.postWithToken("/login", url.Values{"email": {"csrf@example.com"}, "password": {"secret-password"}}, "")
	missing.AssertStatus(419)
	profile, _ := b.get("/account/profile")
	profile.AssertStatus(303) // still signed out
}

func (s *WebAuthTestSuite) TestLoginIsThrottledAfterFiveTries() {
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	for try := 0; try < 5; try++ {
		response, _ := b.post("/login", url.Values{"email": {"slow@example.com"}, "password": {"wrong-password"}})
		response.AssertStatus(422)
	}
	blocked, body := b.post("/login", url.Values{"email": {"slow@example.com"}, "password": {"wrong-password"}})
	blocked.AssertStatus(422)
	s.Contains(body, "Too many attempts")
}

func (s *WebAuthTestSuite) TestForgotAndResetPassword() {
	s.register(newBrowser(s.T(), &s.TestCase), "reset@example.com")
	s.mails = nil

	b := newBrowser(s.T(), &s.TestCase)
	b.get("/forgot-password")
	unknown, _ := b.post("/forgot-password", url.Values{"email": {"nobody@example.com"}})
	s.Equal("/forgot-password?status=reset-link-sent", location(unknown), "the same answer for an unknown address")
	s.Empty(s.mails)

	known, _ := b.post("/forgot-password", url.Values{"email": {"reset@example.com"}})
	s.Equal("/forgot-password?status=reset-link-sent", location(known))
	s.Equal("Reset your password", s.mails[0].Subject)
	link := s.lastLink()

	_, form := b.get(link)
	s.Contains(form, `data-testid="reset-form"`)
	short, _ := b.post(link, url.Values{"password": {"short"}, "password_confirmation": {"short"}})
	short.AssertStatus(422)
	done, _ := b.post(link, url.Values{"password": {"brand-new-password"}, "password_confirmation": {"brand-new-password"}})
	s.Equal("/login?status=password-reset", location(done))

	_, ok := auth.CheckCredentials("reset@example.com", "brand-new-password")
	s.True(ok, "the new password works")
	_, oldOK := auth.CheckCredentials("reset@example.com", "secret-password")
	s.False(oldOK, "the old one does not")

	used, _ := b.get(link)
	s.Equal("/forgot-password?status=link-invalid", location(used), "a link works once")
}

func (s *WebAuthTestSuite) TestAnExpiredLinkIsRefused() {
	s.register(newBrowser(s.T(), &s.TestCase), "late@example.com")
	link := s.lastLink()
	_, err := facades.DB().Table("user_security_tokens").Where("id > ?", 0).Update("expires_at", time.Now().UTC().Add(-time.Minute))
	s.Require().NoError(err)
	b := newBrowser(s.T(), &s.TestCase)
	expired, _ := b.get(link)
	s.Equal("/login?status=link-invalid", location(expired))
}

func (s *WebAuthTestSuite) TestTheRealMailerPutsTheMailOnTheDatabaseQueue() {
	before, _ := facades.DB().Table("jobs").Count()
	err := realSendMail(auth.Outgoing{To: "queued@example.com", Subject: "Queued", HTML: "<p>Hi</p>", Text: "Hi"})
	s.Require().NoError(err)
	after, _ := facades.DB().Table("jobs").Count()
	s.Equal(before+1, after, "Mail().Queue() stored a job; a worker sends it")
	var payload struct{ Payload string }
	s.Require().NoError(facades.DB().Table("jobs").OrderBy("id", "desc").First(&payload))
	s.Contains(payload.Payload, "queued@example.com")
}

func (s *WebAuthTestSuite) TestSignInWithAnEmailLink() {
	_, _, err := auth.CreateUser("Link User", "link@example.com", "secret-password")
	s.Require().NoError(err)

	b := newBrowser(s.T(), &s.TestCase)
	_, login := b.get("/login")
	s.Contains(login, `data-testid="email-link-link"`, "the login page offers it")
	response, form := b.get("/login/email-link")
	response.AssertStatus(200)
	s.Contains(form, `data-testid="email-link-form"`)

	unknown, _ := b.post("/login/email-link", url.Values{"email": {"nobody@example.com"}})
	s.Equal("/login/email-link?status=sign-in-link-sent", location(unknown), "the same answer for an unknown address")
	s.Empty(s.mails)

	known, _ := b.post("/login/email-link", url.Values{"email": {" LINK@example.com "}})
	s.Equal("/login/email-link?status=sign-in-link-sent", location(known))
	s.Require().Len(s.mails, 1)
	s.Equal("Your sign-in link", s.mails[0].Subject)
	s.Contains(s.mails[0].Text, "15 minutes")
	link := s.lastLink()
	s.True(strings.HasPrefix(link, "/login/email-link/"), link)

	// Opening the link (as a mail scanner does) shows a button and uses nothing.
	visitor := newBrowser(s.T(), &s.TestCase)
	for range 2 {
		response, confirm := visitor.get(link)
		response.AssertStatus(200)
		s.Contains(confirm, `data-testid="email-link-confirm-form"`)
	}
	notYet, _ := visitor.get("/account/profile")
	s.True(strings.HasPrefix(location(notYet), "/login"), "not signed in by opening the link")

	forged, _ := visitor.postWithToken(link, url.Values{}, "")
	forged.AssertStatus(419)

	done, _ := visitor.post(link, url.Values{})
	s.Equal("/account/profile", location(done))
	profile, body := visitor.get("/account/profile")
	profile.AssertStatus(200)
	s.Contains(body, "link@example.com")

	var user models.User
	s.Require().NoError(facades.Orm().Query().Where("email", "link@example.com").First(&user))
	s.NotNil(user.EmailVerifiedAt, "reading the mail proves the address")
	logged, _ := facades.DB().Table("activity_log").Where("event", "login_by_email_link").Count()
	s.Equal(int64(1), logged)

	again, _ := newBrowser(s.T(), &s.TestCase).get(link)
	s.Equal("/login?status=link-invalid", location(again), "a link works once")
}

func (s *WebAuthTestSuite) TestAnExpiredSignInLinkIsRefused() {
	_, _, err := auth.CreateUser("Late User", "late-link@example.com", "secret-password")
	s.Require().NoError(err)
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login/email-link")
	b.post("/login/email-link", url.Values{"email": {"late-link@example.com"}})
	link := s.lastLink()
	_, err = facades.DB().Table("user_security_tokens").Where("purpose", auth.PurposeSignIn).
		Update("expires_at", time.Now().UTC().Add(-time.Minute))
	s.Require().NoError(err)

	opened, _ := b.get(link)
	s.Equal("/login?status=link-invalid", location(opened))
	used, _ := b.post(link, url.Values{})
	s.Equal("/login?status=link-invalid", location(used))
	profile, _ := b.get("/account/profile")
	s.True(strings.HasPrefix(location(profile), "/login"), "not signed in")
}

func (s *WebAuthTestSuite) TestSignInLinkRequestsAreThrottled() {
	_, _, err := auth.CreateUser("Busy User", "busy@example.com", "secret-password")
	s.Require().NoError(err)
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login/email-link")
	for range 3 {
		response, _ := b.post("/login/email-link", url.Values{"email": {"busy@example.com"}})
		s.Equal("/login/email-link?status=sign-in-link-sent", location(response))
	}
	blocked, _ := b.post("/login/email-link", url.Values{"email": {"busy@example.com"}})
	blocked.AssertStatus(429)
	s.Len(s.mails, 3, "no fourth mail")
}
