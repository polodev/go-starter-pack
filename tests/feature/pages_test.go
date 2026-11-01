package feature

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	activitymodels "goravel/app-modules/activity-log/models"
	"goravel/app-modules/auth"
	"goravel/app-modules/pages"
	"goravel/app-modules/pages/models"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/tests"
)

// PagesTestSuite covers bilingual pages: public pages with per-part English fallback, drafts, and the admin
// screens with HTML cleaning and permissions.
type PagesTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestPagesTestSuite(t *testing.T) {
	suite.Run(t, new(PagesTestSuite))
}

func (s *PagesTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	s.Require().NoError((&pages.PagesSeeder{}).Run())
	auth.SendMail = func(auth.Outgoing) error { return nil }
}

func (s *PagesTestSuite) signedIn(email, role string) *browser {
	user, _, err := auth.CreateUser("Staff", email, "secret-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	b := newBrowser(s.T(), &s.TestCase)
	b.get("/login")
	b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
	return b
}

func (s *PagesTestSuite) TestAPublishedPageInEnglishAndBangla() {
	english, body := newBrowser(s.T(), &s.TestCase).get("/pages/about")
	english.AssertStatus(200)
	s.Contains(body, `data-testid="page-heading">About us`)
	s.Contains(body, "<h2>Our promise</h2>", "the cleaned HTML is shown as HTML")
	s.NotContains(body, `data-testid="fallback-note"`)

	bangla, err := s.Http(s.T()).WithHeader("Accept-Language", "bn").Get("/pages/about")
	s.Require().NoError(err)
	text, _ := bangla.Content()
	s.Contains(text, `data-testid="page-heading">আমাদের সম্পর্কে`)

	missing, _ := newBrowser(s.T(), &s.TestCase).get("/pages/nothing-here")
	missing.AssertStatus(404)
}

func (s *PagesTestSuite) TestAnEmptyBanglaPartFallsBackToEnglishOnItsOwn() {
	page, _ := pages.FindBySlug("terms-of-service")
	s.Require().NoError(pages.SaveTranslations(page.ID, map[string]pages.Translation{
		"en": {Title: "Terms", Description: "English description", Body: "<p>English text</p>"},
		"bn": {Title: "শর্তাবলি"},
	}))
	content := pages.Load(page, "bn")
	s.Equal("শর্তাবলি", content.Title)
	s.Equal("English description", content.Description)
	s.Equal("<p>English text</p>", string(content.Body))
	s.True(content.Fallback)

	s.Require().NoError(pages.SaveTranslations(page.ID, map[string]pages.Translation{
		"en": {Title: "Terms", Body: "<p>English text</p>"},
	}))
	count, _ := facades.Orm().Query().Model(&models.PageTranslation{}).Where("page_id", page.ID).Where("locale", "bn").Count()
	s.Equal(int64(0), count, "clearing a whole language deletes its row")
}

func (s *PagesTestSuite) TestStaffCreateAndEditPagesWithCleanHtml() {
	viewer := s.signedIn("bd@example.com", "business_development")
	list, body := viewer.get("/admin/pages")
	list.AssertStatus(200)
	s.Contains(body, "Privacy policy")
	s.NotContains(body, `data-testid="new-page"`)
	refused, _ := viewer.get("/admin/pages/create")
	refused.AssertStatus(403)

	editor := s.signedIn("employee@example.com", "employee")
	editor.get("/admin/pages/create")
	bad, page := editor.post("/admin/pages", url.Values{"slug": {"Not A Slug"}, "title_en": {""}, "body_en": {""}})
	bad.AssertStatus(422)
	s.Contains(page, `data-testid="error-slug"`)
	s.Contains(page, `data-testid="error-en"`)
	editor.get("/admin/pages/create")
	taken, _ := editor.post("/admin/pages", url.Values{"slug": {"about"}, "title_en": {"Dup"}, "body_en": {"<p>x</p>"}})
	taken.AssertStatus(422)
	before, _ := facades.Orm().Query().Model(&models.Page{}).Count()
	s.Equal(int64(3), before, "nothing stored after a refusal")

	editor.get("/admin/pages/create")
	created, _ := editor.post("/admin/pages", url.Values{
		"slug": {"faq"}, "is_published": {"1"}, "sort_order": {"4"},
		"title_en": {"FAQ"}, "body_en": {`<p onclick="steal()">Answers</p><script>alert(1)</script><a href="javascript:evil()">x</a>`},
	})
	s.Contains(location(created), "?saved=page")
	faq, ok := pages.FindBySlug("faq")
	s.Require().True(ok)
	stored := pages.Load(faq, "en")
	s.Equal("FAQ", stored.Title)
	s.NotContains(string(stored.Body), "script")
	s.NotContains(string(stored.Body), "onclick")
	s.NotContains(string(stored.Body), "javascript:")
	s.Contains(string(stored.Body), "Answers")
	bn, _ := facades.Orm().Query().Model(&models.PageTranslation{}).Where("page_id", faq.ID).Where("locale", "bn").Count()
	s.Equal(int64(0), bn, "an empty Bangla form stores no row")

	editor.get("/admin/pages/" + strconv.Itoa(int(faq.ID)) + "/edit")
	editor.post("/admin/pages/"+strconv.Itoa(int(faq.ID)), url.Values{"slug": {"faq"}, "title_en": {"FAQ"}, "body_en": {"<p>Answers</p>"}})
	draft, _ := newBrowser(s.T(), &s.TestCase).get("/pages/faq")
	draft.AssertStatus(404) // unpublished again: a draft is hidden

	var entries []activitymodels.Activity
	s.Require().NoError(facades.Orm().Query().Where("log_name", "pages").Find(&entries))
	s.Len(entries, 2, "created and updated are recorded")
}
