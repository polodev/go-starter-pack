package feature

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app-modules/auth"
	socialmodels "goravel/app-modules/social-auth/models"
	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

// fakeGoogle plays Google's token, userinfo and tokeninfo endpoints. profile is who signs in.
type fakeGoogle struct {
	server   *httptest.Server
	mu       sync.Mutex
	profile  map[string]any
	verifier string // the PKCE verifier the token request carried
}

func newFakeGoogle() *fakeGoogle {
	google := &fakeGoogle{}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		google.mu.Lock()
		google.verifier = r.PostForm.Get("code_verifier")
		google.mu.Unlock()
		if r.PostForm.Get("code") != "good-code" || r.PostForm.Get("client_secret") != "test-secret" {
			http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"access-123","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access-123" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		google.mu.Lock()
		defer google.mu.Unlock()
		_ = json.NewEncoder(w).Encode(google.profile)
	})
	mux.HandleFunc("/tokeninfo", func(w http.ResponseWriter, r *http.Request) {
		google.mu.Lock()
		defer google.mu.Unlock()
		switch r.URL.Query().Get("id_token") {
		case "good-id-token":
			claims := map[string]any{"aud": "test-client"}
			for key, value := range google.profile {
				claims[key] = value
			}
			if verified, ok := claims["email_verified"].(bool); ok {
				claims["email_verified"] = map[bool]string{true: "true", false: "false"}[verified] // tokeninfo sends a string
			}
			_ = json.NewEncoder(w).Encode(claims)
		case "other-app-token":
			_ = json.NewEncoder(w).Encode(map[string]any{"aud": "someone-else", "sub": "1", "email": "x@example.com", "email_verified": "true"})
		default:
			http.Error(w, `{"error":"invalid_token"}`, http.StatusBadRequest)
		}
	})
	google.server = httptest.NewServer(mux)
	return google
}

type SocialAuthTestSuite struct {
	suite.Suite
	tests.TestCase
	google   *fakeGoogle
	original map[string]any
}

func TestSocialAuthTestSuite(t *testing.T) {
	suite.Run(t, new(SocialAuthTestSuite))
}

var googleSettings = []string{"client_id", "client_secret", "redirect_url", "client_ids", "auto_create", "auth_url", "token_url", "userinfo_url", "tokeninfo_url"}

func (s *SocialAuthTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	auth.SendMail = func(auth.Outgoing) error { return nil }
	s.google = newFakeGoogle()
	s.google.profile = map[string]any{"sub": "google-1", "email": "Amina@Example.com", "email_verified": true, "name": "Amina Khan"}
	config := facades.Config()
	s.original = map[string]any{}
	for _, key := range googleSettings {
		s.original[key] = config.Get("services.google." + key)
	}
	for key, value := range map[string]any{
		"client_id": "test-client", "client_secret": "test-secret", "redirect_url": "http://localhost/auth/google/callback",
		"client_ids": "", "auto_create": true, "auth_url": s.google.server.URL + "/auth", "token_url": s.google.server.URL + "/token",
		"userinfo_url": s.google.server.URL + "/userinfo", "tokeninfo_url": s.google.server.URL + "/tokeninfo",
	} {
		config.Add("services.google."+key, value)
	}
}

func (s *SocialAuthTestSuite) TearDownTest() {
	s.google.server.Close()
	for key, value := range s.original {
		facades.Config().Add("services.google."+key, value)
	}
}

// startSignIn opens /auth/google/redirect and returns the state Google would send back.
func (s *SocialAuthTestSuite) startSignIn(b *browser) string {
	response, _ := b.get("/auth/google/redirect")
	response.AssertStatus(303)
	target, err := url.Parse(location(response))
	s.Require().NoError(err)
	s.Equal(s.google.server.URL+"/auth", target.Scheme+"://"+target.Host+target.Path)
	query := target.Query()
	s.Equal("test-client", query.Get("client_id"))
	s.Equal("code", query.Get("response_type"))
	s.Equal("http://localhost/auth/google/callback", query.Get("redirect_uri"))
	s.Contains(query.Get("scope"), "email")
	s.Equal("S256", query.Get("code_challenge_method"), "PKCE")
	s.NotEmpty(query.Get("code_challenge"))
	s.NotEmpty(query.Get("state"))
	return query.Get("state")
}

func (s *SocialAuthTestSuite) users() []models.User {
	var users []models.User
	s.Require().NoError(facades.Orm().Query().Find(&users))
	return users
}

func (s *SocialAuthTestSuite) TestANewPersonSignsInWithGoogleAndGetsAVerifiedAccount() {
	b := newBrowser(s.T(), &s.TestCase)
	_, login := b.get("/login")
	s.Contains(login, `href="/auth/google/redirect"`, "the button shows when Google is set up")

	state := s.startSignIn(b)
	response, _ := b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/account/profile", location(response))
	s.NotEmpty(s.google.verifier, "the PKCE verifier went with the code")
	profile, body := b.get("/account/profile")
	profile.AssertStatus(200)
	s.Contains(body, "amina@example.com")

	users := s.users()
	s.Require().Len(users, 1)
	s.Equal("Amina Khan", users[0].Name)
	s.NotNil(users[0].EmailVerifiedAt)
	var links []socialmodels.SocialAccount
	s.Require().NoError(facades.Orm().Query().Find(&links))
	s.Require().Len(links, 1)
	s.Equal("google:google-1", links[0].Identity)

	// The second time, the link finds the same person, even after they changed their Google email.
	s.google.profile["email"] = "amina.new@example.com"
	again := newBrowser(s.T(), &s.TestCase)
	state = s.startSignIn(again)
	response, _ = again.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/account/profile", location(response))
	s.Len(s.users(), 1, "no second account")
	logged, _ := facades.DB().Table("activity_log").Where("event", "login_with_google").Count()
	s.Equal(int64(2), logged)
}

func (s *SocialAuthTestSuite) TestAnExistingAccountWithTheSameEmailIsLinkedAndAnUnverifiedOneLosesItsPassword() {
	_, _, err := auth.CreateUser("Amina", "amina@example.com", "someone-elses-password")
	s.Require().NoError(err)
	b := newBrowser(s.T(), &s.TestCase)
	state := s.startSignIn(b)
	response, _ := b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/account/profile", location(response))
	s.Len(s.users(), 1, "linked, not duplicated")
	_, ok := auth.CheckCredentials("amina@example.com", "someone-elses-password")
	s.False(ok, "whoever registered the unverified address before its owner is locked out")
}

func (s *SocialAuthTestSuite) TestAVerifiedAccountKeepsItsPassword() {
	user, _, err := auth.CreateUser("Amina", "amina@example.com", "her-own-password")
	s.Require().NoError(err)
	s.Require().NoError(auth.MarkVerified(&user))
	b := newBrowser(s.T(), &s.TestCase)
	state := s.startSignIn(b)
	b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	_, ok := auth.CheckCredentials("amina@example.com", "her-own-password")
	s.True(ok)
}

func (s *SocialAuthTestSuite) TestAWrongOrMissingStateOrABadCodeSignsNobodyIn() {
	b := newBrowser(s.T(), &s.TestCase)
	s.startSignIn(b)
	for _, query := range []string{"code=good-code&state=forged", "code=good-code", "error=access_denied"} {
		response, _ := b.get("/auth/google/callback?" + query)
		s.Equal("/login?status=google-failed", location(response), query)
	}
	stranger := newBrowser(s.T(), &s.TestCase)
	response, _ := stranger.get("/auth/google/callback?code=good-code&state=anything")
	s.Equal("/login?status=google-failed", location(response), "no state in the session")

	state := s.startSignIn(b)
	response, _ = b.get("/auth/google/callback?code=bad-code&state=" + url.QueryEscape(state))
	s.Equal("/login?status=google-failed", location(response))
	reused, _ := b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/login?status=google-failed", location(reused), "a state works once")
	s.Empty(s.users())
	_, page := b.get("/login?status=google-failed")
	s.Contains(page, "Signing in with Google did not work")
}

func (s *SocialAuthTestSuite) TestAnUnverifiedGoogleEmailOrNoAccountWithAutoCreateOffIsRefused() {
	s.google.profile["email_verified"] = false
	b := newBrowser(s.T(), &s.TestCase)
	state := s.startSignIn(b)
	response, _ := b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/login?status=google-email-unverified", location(response))

	s.google.profile["email_verified"] = true
	facades.Config().Add("services.google.auto_create", false)
	state = s.startSignIn(b)
	response, _ = b.get("/auth/google/callback?code=good-code&state=" + url.QueryEscape(state))
	s.Equal("/login?status=google-no-account", location(response))
	s.Empty(s.users())
}

func (s *SocialAuthTestSuite) TestWithoutAClientIdTheButtonIsHiddenAndTheRoutesSayNotSetUp() {
	facades.Config().Add("services.google.client_id", "")
	b := newBrowser(s.T(), &s.TestCase)
	_, login := b.get("/login")
	s.NotContains(login, "/auth/google/redirect")
	response, _ := b.get("/auth/google/redirect")
	s.Equal("/login?status=google-unavailable", location(response))
	api, _ := s.Http(s.T()).Post("/api/v1/auth/google", strings.NewReader(`{"id_token":"good-id-token"}`))
	api.AssertStatus(404)
}

func (s *SocialAuthTestSuite) TestTheApiTradesAGoogleIdTokenForOurToken() {
	post := func(status int, body string) map[string]any {
		response, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/google", strings.NewReader(body))
		s.Require().NoError(err)
		response.AssertStatus(status)
		parsed, _ := response.Json()
		return parsed
	}
	body := post(200, `{"id_token":"good-id-token"}`)
	s.NotEmpty(body["token"])
	user, _ := body["user"].(map[string]any)
	s.Equal("amina@example.com", user["email"])
	s.NotContains(user, "password")

	me, err := s.Http(s.T()).WithHeader("Authorization", "Bearer "+body["token"].(string)).Get("/api/v1/auth/me")
	s.Require().NoError(err)
	me.AssertStatus(200)

	body = post(422, `{"id_token":"other-app-token"}`) // made for another app
	s.Equal("The Google token is invalid.", body["message"])
	post(422, `{"id_token":"broken"}`)
	body = post(422, `{}`)
	s.Contains(body["message"], "id token")
	s.Len(s.users(), 1)

	facades.Config().Add("services.google.client_ids", "someone-else")
	post(200, `{"id_token":"other-app-token"}`) // GOOGLE_CLIENT_IDS adds the mobile apps' client ids
}
