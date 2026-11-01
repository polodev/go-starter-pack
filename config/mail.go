package config

import (
	contractsconfig "github.com/goravel/framework/contracts/config"

	"goravel/app/facades"
	"goravel/app/support/ses"
)

// smtpSettings is MAIL_HOST/PORT/USERNAME/PASSWORD, or, with MAIL_MAILER=ses, the Amazon SES SMTP endpoint and
// the SMTP password derived from AWS_SES_SECRET_ACCESS_KEY (the user name is the access key id).
func smtpSettings(config contractsconfig.Config) (host, port, username, password any) {
	if mailer, _ := config.Env("MAIL_MAILER", "smtp").(string); mailer == "ses" {
		region, _ := config.Env("AWS_SES_DEFAULT_REGION", "us-east-1").(string)
		secret, _ := config.Env("AWS_SES_SECRET_ACCESS_KEY", "").(string)
		return ses.SMTPHost(region), 587, config.Env("AWS_SES_ACCESS_KEY_ID"), ses.SMTPPassword(secret, region)
	}
	return config.Env("MAIL_HOST", ""), config.Env("MAIL_PORT", 587), config.Env("MAIL_USERNAME"), config.Env("MAIL_PASSWORD")
}

func init() {
	config := facades.Config()
	host, port, username, password := smtpSettings(config)
	config.Add("mail", map[string]any{
		// SMTP Host Address
		//
		// Here you may provide the host address of the SMTP server used by your
		// applications. A default option is provided that is compatible with
		// the Mailgun mail service which will provide reliable deliveries.
		// MAIL_MAILER=smtp (locally a catcher such as Mailpit on 127.0.0.1:1025) or ses (see smtpSettings).
		"host": host,

		// SMTP Host Port
		//
		// This is the SMTP port used by your application to deliver e-mails to
		// users of the application. Like the host we have set this value to
		// stay compatible with the Mailgun e-mail application by default.
		"port": port,

		// --------------------------------------------------------------------------
		// Global "From" Address
		// --------------------------------------------------------------------------
		//
		// You may wish for all e-mails sent by your application to be sent from
		// the same address. Here, you may specify a name and address that is
		// used globally for all e-mails that are sent by your application.
		"from": map[string]any{
			"address": config.Env("MAIL_FROM_ADDRESS", "hello@example.com"),
			"name":    config.Env("MAIL_FROM_NAME", "Example"),
		},

		// SMTP Server Username
		//
		// If your SMTP server requires a username for authentication, you should
		// set it here. This will get used to authenticate with your server on
		// connection. You may also set the "password" value below this one.
		"username": username,

		"password": password,

		// Template Configuration
		//
		// This controls template rendering for email views. Template engines are cached
		// globally and support both built-in drivers and custom implementations via factories.
		//
		// Available Drivers: "html", "custom"
		"template": map[string]any{
			"default": config.Env("MAIL_TEMPLATE_ENGINE", "html"),
			"engines": map[string]any{
				"html": map[string]any{
					"driver": "html",
					"path":   config.Env("MAIL_VIEWS_PATH", "resources/views/mail"),
				},
				// Example custom template engine:
				// "blade": map[string]any{
				//     "driver": "custom",
				//     "via": func() (mail.Template, error) {
				//         return NewBladeTemplateEngine(), nil
				//     },
				// },
			},
		},
	})
}
