package middleware

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

// SecurityHeaders adds the browser protections Laravel apps usually get from a package or the web server: a
// Content-Security-Policy, no framing, no MIME sniffing, a strict referrer policy and, in production, HSTS.
// It is global middleware (bootstrap/app.go), so it covers every page, API answer, static file and 404.
// A handler that needs something else (for example a page that may be framed) sets its own header after Next.
type SecurityHeaders struct{}

func (m *SecurityHeaders) Signature() string {
	return "security-headers"
}

func (m *SecurityHeaders) Handle(ctx http.Context) {
	config := facades.Config()
	response := ctx.Response()
	if policy := config.GetString("security.content_security_policy"); policy != "" {
		response.Header("Content-Security-Policy", policy)
	}
	response.Header("X-Content-Type-Options", "nosniff")
	response.Header("X-Frame-Options", "DENY")
	response.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	response.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
	response.Header("Cross-Origin-Opener-Policy", "same-origin")
	if config.GetBool("security.strict_transport_security") {
		response.Header("Strict-Transport-Security", "max-age="+config.GetString("security.strict_transport_security_max_age")+"; includeSubDomains")
	}
	ctx.Request().Next()
}
