package socialauth

import (
	"errors"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/activity"
	"goravel/app/facades"
	"goravel/app/http/responses"
)

// ApiGoogle is POST /api/v1/auth/google {"id_token": "..."}: a mobile app or single page app signs in with Google
// itself and sends the ID token; the answer is the same {user, token} as /api/v1/auth/login.
func ApiGoogle(ctx http.Context) http.Response {
	if !Enabled() {
		return ctx.Response().Json(http.StatusNotFound, http.Json{"message": "Google sign-in is not set up."})
	}
	idToken := strings.TrimSpace(ctx.Request().Input("id_token"))
	if idToken == "" {
		return responses.ValidationFailed(ctx, map[string][]string{"id_token": {"The id token field is required."}})
	}
	profile, err := verifyIDToken(ctx.Context(), idToken)
	if err != nil {
		return responses.ValidationFailed(ctx, map[string][]string{"id_token": {"The Google token is invalid."}})
	}
	user, err := FindOrCreateUser(profile, AutoCreate())
	switch {
	case errors.Is(err, ErrEmailNotVerified):
		return responses.ValidationFailed(ctx, map[string][]string{"id_token": {"Google has not verified this email address."}})
	case errors.Is(err, ErrNoAccount):
		return responses.ValidationFailed(ctx, map[string][]string{"email": {"No account uses this email address."}})
	case err != nil:
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	token, err := facades.Auth(ctx).Guard("api").Login(&user)
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"message": "Server Error"})
	}
	activity.In("auth").CausedBy(user.ID).PerformedOn("user", user.ID).Event("login_with_google").LogQuietly("logged in with Google (API)")
	return ctx.Response().Json(http.StatusOK, http.Json{"user": user, "token": token})
}
