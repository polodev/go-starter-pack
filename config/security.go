package config

import (
	"goravel/app/facades"
)

// DefaultContentSecurityPolicy allows scripts, styles and fonts only from this site (no inline scripts: the theme
// switch lives in public/js/theme.js), pictures also from https (the R2 public bucket), and no framing.
const DefaultContentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; " +
	"font-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'; object-src 'none'"

func init() {
	config := facades.Config()
	config.Add("security", map[string]any{
		// Sent on every response by app/http/middleware.SecurityHeaders. An empty value sends no policy.
		"content_security_policy": config.Env("SECURITY_CONTENT_SECURITY_POLICY", DefaultContentSecurityPolicy),
		// HSTS tells browsers to use https only. On by default in production; turn it on only behind https.
		"strict_transport_security":         config.Env("SECURITY_STRICT_TRANSPORT_SECURITY", config.Env("APP_ENV", "production") == "production"),
		"strict_transport_security_max_age": config.Env("SECURITY_STRICT_TRANSPORT_SECURITY_MAX_AGE", 31536000),
	})
}
