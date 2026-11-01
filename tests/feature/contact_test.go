package feature

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	"goravel/app-modules/contact"
	"goravel/app-modules/contact/models"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/tests"
)

// ContactTestSuite covers the public contact form (validation, honeypot, rate limit, CSRF, both languages) and the
// admin inbox (who may read and change messages).
type ContactTestSuite struct {
	suite.Suite
	tests.TestCase
	notified []models.Contact
}

func TestContactTestSuite(t *testing.T) {
	suite.Run(t, new(ContactTestSuite))
}

func (s *ContactTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
	s.notified = nil
	contact.NotifyStaff = func(message models.Contact) error {
		s.notified = append(s.notified, message)
		return nil
	}
}

func (s *ContactTestSuite) stored() []models.Contact {
	var messages []models.Contact
	s.Require().NoError(facades.Orm().Query().OrderBy("id").Find(&messages))
	return messages
}

func validMessage() url.Values {
	return url.Values{"name": {"Rahim Uddin"}, "email": {"Rahim@Example.com"}, "subject": {"Opening hours"},
		"message": {"When are you open on Fridays?"}}
}

func (s *ContactTestSuite) TestAMessageIsStoredTheStaffIsToldAndTheVisitorThanked() {
	visitor := newBrowser(s.T(), &s.TestCase)
	response, body := visitor.get("/contact")
	response.AssertStatus(200)
	s.Contains(body, `data-testid="contact-form"`)
	s.Contains(body, `name="website"`, "the honeypot is there")
	s.NotContains(body, `style="`, "the honeypot hides with classes: the CSP blocks inline styles")

	response, _ = visitor.post("/contact", validMessage())
	s.Equal("/contact?status=sent", location(response))
	_, body = visitor.get("/contact?status=sent")
	s.Contains(body, `data-testid="contact-success"`)

	messages := s.stored()
	s.Require().Len(messages, 1)
	s.Equal("Rahim Uddin", messages[0].Name)
	s.Equal("rahim@example.com", messages[0].Email)
	s.Equal("Opening hours", messages[0].SubjectText())
	s.Equal(models.StatusNew, messages[0].Status)
	s.Equal("en", messages[0].Locale)
	s.Require().Len(s.notified, 1, "the staff is told")
}

func (s *ContactTestSuite) TestBanglaVisitorsGetBanglaTextAndBanglaLengthsCountAsCharacters() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/language/bn?next=/contact") // the language switch remembers Bangla in a cookie
	_, body := visitor.get("/contact")
	s.Contains(body, "যোগাযোগ করুন")

	form := validMessage()
	form.Set("name", "রহিম") // 4 characters, 12 bytes
	form.Set("message", "শুক্রবার কখন খোলা থাকে?")
	response, _ := visitor.post("/contact", form)
	s.Equal("/contact?status=sent", location(response))
	messages := s.stored()
	s.Require().Len(messages, 1)
	s.Equal("bn", messages[0].Locale)

	form.Set("name", "র")
	response, body = visitor.post("/contact", form)
	response.AssertStatus(422)
	s.Contains(body, "অনুগ্রহ করে আপনার নাম লিখুন", "the error is in Bangla")
}

func (s *ContactTestSuite) TestInvalidInputIsRefusedWithTheTypedValuesKept() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/contact")
	form := url.Values{"name": {"R"}, "email": {"not-an-email"}, "subject": {strings.Repeat("s", 151)}, "message": {"short"}}
	response, body := visitor.post("/contact", form)
	response.AssertStatus(422)
	for _, field := range []string{"name", "email", "subject", "message"} {
		s.Contains(body, `data-testid="error-`+field+`"`)
	}
	s.Contains(body, `value="not-an-email"`, "what was typed stays in the form")
	s.Contains(body, ">short</textarea>")

	for _, email := range []string{"a@b", "a@b.", "Name <a@b.com>", "a b@c.com"} {
		form := validMessage()
		form.Set("email", email)
		response, _ := visitor.post("/contact", form)
		response.AssertStatus(422)
	}
	s.Empty(s.stored(), "nothing was stored")
	s.Empty(s.notified)
}

func (s *ContactTestSuite) TestTheHoneypotLooksLikeSuccessButStoresNothing() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/contact")
	form := validMessage()
	form.Set("website", "https://spam.example")
	response, _ := visitor.post("/contact", form)
	s.Equal("/contact?status=sent", location(response), "the bot learns nothing")
	s.Empty(s.stored())
	s.Empty(s.notified)
}

func (s *ContactTestSuite) TestAnAddressMaySendFiveMessagesPerTenMinutes() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/contact")
	for range 5 {
		response, _ := visitor.post("/contact", validMessage())
		s.Equal("/contact?status=sent", location(response))
	}
	response, body := visitor.post("/contact", validMessage())
	response.AssertStatus(429)
	s.Contains(body, `data-testid="contact-failure"`)
	s.Contains(body, `value="Rahim Uddin"`, "the typed message is not lost")
	s.Len(s.stored(), 5)
}

func (s *ContactTestSuite) TestWithoutTheCsrfTokenNothingIsStored() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/contact")
	response, _ := visitor.postWithToken("/contact", validMessage(), "")
	response.AssertStatus(419)
	s.Empty(s.stored())
}

func (s *ContactTestSuite) signedInAs(role string) *browser {
	email := role + "@example.com"
	user, _, err := auth.CreateUser("Person "+role, email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	if role != "customer" {
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	response, _ := b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	response.AssertStatus(303)
	b.get("/") // the session is new after login: pick up its CSRF token
	return b
}

func (s *ContactTestSuite) TestTheInboxIsForStaffWithContactsViewAndChangesNeedContactsManage() {
	visitor := newBrowser(s.T(), &s.TestCase)
	visitor.get("/contact")
	visitor.post("/contact", validMessage())
	message := s.stored()[0]
	show := "/admin/contacts/" + strconv.Itoa(int(message.ID))

	for _, role := range []string{"customer", "new_employee"} {
		response, _ := s.signedInAs(role).get("/admin/contacts")
		response.AssertStatus(403)
	}

	reader := s.signedInAs("business_development")
	response, body := reader.get("/admin/contacts")
	response.AssertStatus(200)
	s.Contains(body, `data-testid="contact-row"`)
	s.Contains(body, "Opening hours")
	s.Contains(body, `data-testid="id-link"`)
	_, body = reader.get(show)
	s.Contains(body, "When are you open on Fridays?")
	s.Contains(body, `href="mailto:rahim@example.com?subject=Re:%20Opening%20hours"`)
	s.NotContains(body, `data-testid="status-form"`, "reading does not show the controls")
	response, _ = reader.post(show+"/status", url.Values{"status": {"replied"}})
	response.AssertStatus(403)
	s.Equal(models.StatusNew, s.stored()[0].Status, "nothing changed after the refusal")

	manager := s.signedInAs("employee")
	_, body = manager.get(show)
	s.Contains(body, `data-testid="status-form"`)
	response, _ = manager.post(show+"/status", url.Values{"status": {"replied"}})
	s.Equal(show+"?saved=status", location(response))
	s.Equal(models.StatusReplied, s.stored()[0].Status)
	s.NotNil(s.stored()[0].RepliedAt)

	response, _ = manager.post(show+"/status", url.Values{"status": {"archived"}})
	s.Equal(show+"?error=status", location(response), "only known statuses")

	_, body = manager.get("/admin/contacts?status=new")
	s.NotContains(body, `data-testid="contact-row"`)
	_, body = manager.get("/admin/contacts?q=RAHIM")
	s.Contains(body, `data-testid="contact-row"`, "search ignores case")

	response, _ = manager.post(show+"/delete", url.Values{})
	s.Equal("/admin/contacts?saved=deleted", location(response))
	s.Empty(s.stored())
	logged, _ := facades.DB().Table("activity_log").Where("log_name", "contacts").Count()
	s.Equal(int64(2), logged, "the status change and the deletion are in the activity log")

	response, _ = manager.get(show)
	response.AssertStatus(404)
}
