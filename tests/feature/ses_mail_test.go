package feature

import (
	"testing"

	"github.com/goravel/framework/contracts/mail"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/support/ses"
	"goravel/tests"
)

// SesMailTestSuite sends one real mail through Amazon SES with the keys in .env, to Amazon's mailbox simulator
// (success@simulator.amazonses.com accepts mail and delivers it to nobody). Skipped without SES keys. It proves the
// derived SMTP password, the region's host and the verified from address all work.
type SesMailTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestSesMailTestSuite(t *testing.T) {
	suite.Run(t, new(SesMailTestSuite))
}

func (s *SesMailTestSuite) TestAMailGoesOutThroughAmazonSes() {
	for _, name := range []string{"AWS_SES_ACCESS_KEY_ID", "AWS_SES_SECRET_ACCESS_KEY", "AWS_SES_DEFAULT_REGION", "MAIL_FROM_ADDRESS"} {
		if env(name) == "" {
			s.T().Skipf("%s is not set in .env", name)
		}
	}
	config := facades.Config()
	original := map[string]any{}
	for _, key := range []string{"mail.host", "mail.port", "mail.username", "mail.password"} {
		original[key] = config.Get(key)
	}
	defer func() {
		for key, value := range original {
			config.Add(key, value)
		}
	}()
	// What MAIL_MAILER=ses sets at boot (config/mail.go), applied for this one test.
	region := env("AWS_SES_DEFAULT_REGION")
	config.Add("mail.host", ses.SMTPHost(region))
	config.Add("mail.port", 587)
	config.Add("mail.username", env("AWS_SES_ACCESS_KEY_ID"))
	config.Add("mail.password", ses.SMTPPassword(env("AWS_SES_SECRET_ACCESS_KEY"), region))

	err := facades.Mail().To([]string{"success@simulator.amazonses.com"}).Subject("Go starter pack: SES check").
		Content(mail.Content{Text: "This mail checks the SES settings. It goes to Amazon's simulator only."}).Send()
	s.Require().NoError(err, "SES accepted the mail")

	// The same mail with a wrong password is refused: the test really talks to SES.
	config.Add("mail.password", ses.SMTPPassword("not-the-secret", region))
	err = facades.Mail().To([]string{"success@simulator.amazonses.com"}).Subject("Go starter pack: SES check").
		Content(mail.Content{Text: "This one must be refused."}).Send()
	s.Error(err, "SES refuses a wrong password")
}
