package config

import (
	"goravel/app/facades"
)

// Third-party services, like Laravel's config/services.php.
func init() {
	config := facades.Config()
	config.Add("services", map[string]any{
		// Sign in with Google (app-modules/social-auth). Without a client id the button is hidden and the routes
		// answer "not set up".
		"google": map[string]any{
			"client_id":     config.Env("GOOGLE_CLIENT_ID", ""),
			"client_secret": config.Env("GOOGLE_CLIENT_SECRET", ""),
			// Default {APP_URL}/auth/google/callback; must be an authorised redirect URI of the OAuth client.
			"redirect_url": config.Env("GOOGLE_REDIRECT_URL", ""),
			// Other client ids (Android, iOS, a single page app) whose ID tokens the API accepts, comma separated.
			"client_ids": config.Env("GOOGLE_CLIENT_IDS", ""),
			// false: only people who already have an account (matched by email) can sign in with Google.
			"auto_create": config.Env("GOOGLE_AUTO_CREATE", true),
			// Google's addresses. Tests point them at a fake Google; leave them alone otherwise.
			"auth_url":      "https://accounts.google.com/o/oauth2/v2/auth",
			"token_url":     "https://oauth2.googleapis.com/token",
			"userinfo_url":  "https://openidconnect.googleapis.com/v1/userinfo",
			"tokeninfo_url": "https://oauth2.googleapis.com/tokeninfo",
		},
	})
}
