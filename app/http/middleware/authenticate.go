package middleware

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/responses"
)

// Authenticate lets a request through only with a valid `Authorization: Bearer <token>` (Laravel's auth:api).
type Authenticate struct{}

// Signature The unique name of the middleware.
func (r *Authenticate) Signature() string {
	return "authenticate"
}

// Handle Execute the middleware.
func (r *Authenticate) Handle(ctx http.Context) {
	token, found := strings.CutPrefix(ctx.Request().Header("Authorization"), "Bearer ")
	if !found || token == "" {
		ctx.Request().AbortWithStatusJson(http.StatusUnauthorized, responses.UnauthenticatedBody)
		return
	}
	if _, err := facades.Auth(ctx).Parse(token); err != nil {
		ctx.Request().AbortWithStatusJson(http.StatusUnauthorized, responses.UnauthenticatedBody)
		return
	}
	ctx.Request().Next()
}
