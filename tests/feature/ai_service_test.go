package feature

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	aiservice "goravel/app-modules/ai-service"
	"goravel/app-modules/ai-service/models"
	"goravel/app-modules/auth"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/support/aifake"
	"goravel/tests"
)

// AiServiceTestSuite covers the AI service with a fake provider: logging, sensitive calls, failures, JSON
// replies, and the AI API logs screens.
type AiServiceTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestAiServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AiServiceTestSuite))
}

func (s *AiServiceTestSuite) SetupTest() {
	s.RefreshDatabase()
	_ = facades.Cache().Flush()
	_, err := rolepermission.Sync()
	s.Require().NoError(err)
	auth.SendMail = func(auth.Outgoing) error { return nil }
	aifake.Reply("Hello from the fake model")
	s.T().Setenv("AI_PRICES", `{"fake-model-1": {"input": 3, "output": 15}}`)
}

func (s *AiServiceTestSuite) lastLog() models.AiApiLog {
	var row models.AiApiLog
	s.Require().NoError(facades.Orm().Query().OrderBy("id", "desc").First(&row))
	return row
}

func (s *AiServiceTestSuite) TestACallIsAnsweredAndLoggedWithTokensAndCost() {
	result, err := aiservice.Chat(context.Background(), aiservice.Request{
		Purpose: "greet the visitor", Provider: "fake", Instructions: "Be brief.", Input: "Say hello",
	})
	s.Require().NoError(err)
	s.Equal("Hello from the fake model", result.Text)
	s.Equal(1200, result.InputTokens)
	s.Equal(300, result.OutputTokens)

	row := s.lastLog()
	s.Equal(result.LogID, row.ID)
	s.Equal("text", row.Kind)
	s.Equal("fake", row.Provider)
	s.Equal("fake-model-1", row.ModelName)
	s.Equal("greet the visitor", row.Purpose)
	s.Equal(models.Success, row.Status)
	s.Equal(uint(1200), row.InputTokens)
	s.Require().NotNil(row.Cost)
	s.InDelta(0.0081, *row.Cost, 0.0000001, "1200 x $3 + 300 x $15 per million")
	s.Equal("Say hello", models.Text(row.Prompt))
	s.Equal("Hello from the fake model", models.Text(row.Response))
}

func (s *AiServiceTestSuite) TestSensitiveCallsKeepThePromptAndReplyOutOfTheLog() {
	_, err := aiservice.Chat(context.Background(), aiservice.Request{
		Purpose: "read passport", Provider: "fake", Input: "Passport number A1234567", Sensitive: true,
	})
	s.Require().NoError(err)
	row := s.lastLog()
	s.True(row.Sensitive)
	s.Nil(row.Prompt)
	s.Nil(row.Response)
	s.Equal(uint(1200), row.InputTokens, "the numbers are still kept")
}

func (s *AiServiceTestSuite) TestAFailedCallIsLoggedAndReturnsTheError() {
	_, err := aiservice.Chat(context.Background(), aiservice.Request{Purpose: "summarise", Provider: "fake", Input: "please FAIL"})
	s.Error(err)
	row := s.lastLog()
	s.Equal(models.Failed, row.Status)
	s.Contains(models.Text(row.Error), "provider is down")
	s.Nil(row.Cost)

	_, err = aiservice.Chat(context.Background(), aiservice.Request{Provider: "fake", Input: "no purpose"})
	s.Error(err, "every call needs a purpose")
}

func (s *AiServiceTestSuite) TestJSONRepliesAreDecodedAndCheckedForShape() {
	aifake.Reply("```json\n{\"name\": \"Polo\", \"age\": 30}\n```")
	var person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	_, err := aiservice.JSON(context.Background(), aiservice.Request{Purpose: "extract person", Provider: "fake", Input: "Polo is 30"}, &person)
	s.Require().NoError(err)
	s.Equal("Polo", person.Name)
	s.Equal(30, person.Age)

	aifake.Reply("Sorry, I cannot help.")
	_, err = aiservice.JSON(context.Background(), aiservice.Request{Purpose: "extract person", Provider: "fake", Input: "?"}, &person)
	s.Error(err, "prose instead of JSON is an error, not a silent empty value")
}

func (s *AiServiceTestSuite) TestAnImageIsGeneratedAndLoggedWithoutThePicture() {
	result, err := aiservice.Image(context.Background(), aiservice.ImageRequest{
		Purpose: "draw the logo", Provider: "fake", Prompt: "a friendly gopher",
	})
	s.Require().NoError(err)
	s.Equal(aifake.PNG, result.Bytes)
	s.Equal("image/png", result.MimeType)
	row := s.lastLog()
	s.Equal(result.LogID, row.ID)
	s.Equal("image", row.Kind)
	s.Equal("fake-image-1", row.ModelName)
	s.Equal("a friendly gopher", models.Text(row.Prompt), "the prompt is kept")
	s.Nil(row.Response, "the picture is never stored in the log")

	_, err = aiservice.Image(context.Background(), aiservice.ImageRequest{Purpose: "draw", Provider: "fake", Prompt: "FAIL please"})
	s.Error(err)
	s.Equal(models.Failed, s.lastLog().Status)

	_, err = aiservice.Image(context.Background(), aiservice.ImageRequest{Provider: "fake", Prompt: "no purpose"})
	s.Error(err, "every call needs a purpose")
}

func (s *AiServiceTestSuite) TestTheAiApiLogScreens() {
	_, _ = aiservice.Chat(context.Background(), aiservice.Request{Purpose: "visible call", Provider: "fake", Input: "hi"})
	_, _ = aiservice.Chat(context.Background(), aiservice.Request{Purpose: "secret call", Provider: "fake", Input: "x", Sensitive: true})
	row := s.lastLog()
	for email, role := range map[string]string{"bd@example.com": "business_development", "employee@example.com": "employee", "admin@example.com": "admin"} {
		user, _, err := auth.CreateUser("Staff", email, "secret-password")
		s.Require().NoError(err)
		s.Require().NoError(auth.MarkVerified(&user))
		s.Require().NoError(rolepermission.AssignRole(user.ID, role))
	}
	login := func(email string) *browser {
		b := newBrowser(s.T(), &s.TestCase)
		b.get("/login")
		b.post("/login", url.Values{"email": {email}, "password": {"secret-password"}})
		return b
	}

	forbidden, _ := login("bd@example.com").get("/admin/ai-api-logs")
	forbidden.AssertStatus(403)

	employee := login("employee@example.com")
	list, body := employee.get("/admin/ai-api-logs")
	list.AssertStatus(200)
	s.Contains(body, "visible call")
	s.Contains(body, "$0.0162", "the estimated spend of the two calls")
	_, failedOnly := employee.get("/admin/ai-api-logs?status=failed")
	s.NotContains(failedOnly, `data-testid="ai-log-row"`)
	_, detail := employee.get("/admin/ai-api-logs/" + strconv.Itoa(int(row.ID)))
	s.Contains(detail, `data-testid="sensitive-note"`)
	s.NotContains(detail, `data-testid="prompt"`)
	refused, _ := employee.post("/admin/ai-api-logs/"+strconv.Itoa(int(row.ID))+"/delete", nil)
	refused.AssertStatus(403)

	admin := login("admin@example.com")
	admin.get("/admin/ai-api-logs/" + strconv.Itoa(int(row.ID)))
	deleted, _ := admin.post("/admin/ai-api-logs/"+strconv.Itoa(int(row.ID))+"/delete", nil)
	s.Equal("/admin/ai-api-logs?saved=deleted", location(deleted))
}
