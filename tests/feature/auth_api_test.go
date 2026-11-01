package feature

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

// AuthApiTestSuite covers /api/v1/auth: register, login and the signed-in person, with Laravel's JSON shapes.
type AuthApiTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestAuthApiTestSuite(t *testing.T) {
	suite.Run(t, new(AuthApiTestSuite))
}

func (s *AuthApiTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *AuthApiTestSuite) TestRegisterCreatesAnAccountReturnsATokenAndNeverThePassword() {
	response, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/register", strings.NewReader(`{"name":"Polo Dev","email":"Polo@Example.com","password":"secret-password"}`))
	s.Require().NoError(err)
	response.AssertStatus(201)
	body, err := response.Json()
	s.Require().NoError(err)
	s.NotEmpty(body["token"])
	user := body["user"].(map[string]any)
	s.Equal("polo@example.com", user["email"], "stored lower-case")
	s.Equal("Polo Dev", user["name"])
	_, hasPassword := user["password"]
	s.False(hasPassword, "the password hash is never returned")

	var stored models.User
	s.Require().NoError(facades.Orm().Query().Where("email", "polo@example.com").First(&stored))
	s.True(strings.HasPrefix(stored.Password, "$2"), "bcrypt, like Laravel")
	s.True(facades.Hash().Check("secret-password", stored.Password))
}

func (s *AuthApiTestSuite) TestRegisterValidatesAndRefusesATakenEmail() {
	response, err := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/register", strings.NewReader(`{"name":"","email":"not-an-email","password":"short"}`))
	s.Require().NoError(err)
	response.AssertStatus(422)
	body, _ := response.Json()
	errors := body["errors"].(map[string]any)
	s.Contains(errors, "name")
	s.Contains(errors, "email")
	s.Contains(errors, "password")
	s.Contains(body["message"], "more error", "Laravel's message: the first error and how many more")
	count, _ := facades.Orm().Query().Model(&models.User{}).Count()
	s.Equal(int64(0), count, "nothing stored after a refusal")

	register := `{"name":"A","email":"a@example.com","password":"secret-password"}`
	first, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/register", strings.NewReader(register))
	first.AssertStatus(201)
	second, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").Post("/api/v1/auth/register", strings.NewReader(register))
	second.AssertStatus(422)
	body, _ = second.Json()
	s.Equal("The email has already been taken.", body["message"])
}

func (s *AuthApiTestSuite) TestLoginGivesATokenThatOpensMeAndTreatsWrongEmailAndPasswordAlike() {
	register, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/register", strings.NewReader(`{"name":"B","email":"b@example.com","password":"secret-password"}`))
	register.AssertStatus(201)

	wrongPassword, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/login", strings.NewReader(`{"email":"b@example.com","password":"nope-nope"}`))
	wrongPassword.AssertStatus(422)
	wrongBody, _ := wrongPassword.Json()
	unknown, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/login", strings.NewReader(`{"email":"nobody@example.com","password":"secret-password"}`))
	unknown.AssertStatus(422)
	unknownBody, _ := unknown.Json()
	s.Equal(wrongBody["message"], unknownBody["message"], "no hint whether the account exists")

	login, _ := s.Http(s.T()).WithHeader("Content-Type", "application/json").
		Post("/api/v1/auth/login", strings.NewReader(`{"email":"B@example.com","password":"secret-password"}`))
	login.AssertStatus(200)
	loginBody, _ := login.Json()
	token := loginBody["token"].(string)
	s.NotEmpty(token)

	me, _ := s.Http(s.T()).WithToken(token).Get("/api/v1/auth/me")
	me.AssertStatus(200)
	meBody, _ := me.Json()
	s.Equal("b@example.com", meBody["data"].(map[string]any)["email"])

	anonymous, _ := s.Http(s.T()).Get("/api/v1/auth/me")
	anonymous.AssertStatus(401)
	anonymous.AssertJson(map[string]any{"message": "Unauthenticated."})
	forged, _ := s.Http(s.T()).WithToken("not-a-real-token").Get("/api/v1/auth/me")
	forged.AssertStatus(401)
}
