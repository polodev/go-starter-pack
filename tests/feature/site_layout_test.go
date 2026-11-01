package feature

import (
	"testing"

	"net/http"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

// SiteLayoutTestSuite covers the public layout: the home page in both languages and the language switch.
type SiteLayoutTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestSiteLayoutTestSuite(t *testing.T) {
	suite.Run(t, new(SiteLayoutTestSuite))
}

func (s *SiteLayoutTestSuite) TestTheHomePageSpeaksEnglishOrBangla() {
	english, err := s.Http(s.T()).Get("/")
	s.Require().NoError(err)
	english.AssertStatus(200)
	body, _ := english.Content()
	s.Contains(body, `data-testid="home-title">Welcome`)
	s.Contains(body, `<html lang="en"`)

	bangla, _ := s.Http(s.T()).WithHeader("Accept-Language", "bn-BD,bn;q=0.9").Get("/")
	body, _ = bangla.Content()
	s.Contains(body, `data-testid="home-title">স্বাগতম`, "Accept-Language picks Bangla")

	byQuery, _ := s.Http(s.T()).Get("/?lang=bn")
	body, _ = byQuery.Content()
	s.Contains(body, `<html lang="bn"`)

	unknown, _ := s.Http(s.T()).Get("/?lang=xx")
	body, _ = unknown.Content()
	s.Contains(body, `<html lang="en"`, "an unknown language falls back to English")
}

func (s *SiteLayoutTestSuite) TestTheLanguageSwitchSetsACookieAndOnlyRedirectsWithinTheSite() {
	response, err := s.Http(s.T()).Get("/language/bn?next=/pages/about")
	s.Require().NoError(err)
	response.AssertStatus(303)
	response.AssertCookie("lang", "bn")
	response.AssertHeader("Location", "/pages/about")

	for _, next := range []string{"//evil.example", "https://evil.example/x", "evil"} {
		outside, _ := s.Http(s.T()).Get("/language/en?next=" + next)
		outside.AssertHeader("Location", "/")
	}

	cookie, _ := s.Http(s.T()).WithCookie(&http.Cookie{Name: "lang", Value: "bn"}).Get("/")
	body, _ := cookie.Content()
	s.Contains(body, `<html lang="bn"`, "the cookie is remembered")
}
