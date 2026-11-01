package auth

import (
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/responses"
)

// ApiAuthenticate lets a request through only with a valid `Authorization: Bearer <token>` (Laravel's auth:api).
type ApiAuthenticate struct{}

func (r *ApiAuthenticate) Signature() string { return "api-authenticate" }

func (r *ApiAuthenticate) Handle(ctx http.Context) {
	token, found := strings.CutPrefix(ctx.Request().Header("Authorization"), "Bearer ")
	if !found || token == "" {
		ctx.Request().AbortWithStatusJson(http.StatusUnauthorized, responses.UnauthenticatedBody)
		return
	}
	if _, err := facades.Auth(ctx).Guard("api").Parse(token); err != nil {
		ctx.Request().AbortWithStatusJson(http.StatusUnauthorized, responses.UnauthenticatedBody)
		return
	}
	ctx.Request().Next()
}
