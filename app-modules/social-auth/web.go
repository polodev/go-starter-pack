package socialauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"github.com/goravel/framework/contracts/http"
	"golang.org/x/oauth2"

	"goravel/app/activity"
	"goravel/app/facades"
)

const (
	sessionState    = "google_oauth_state"
	sessionVerifier = "google_oauth_verifier"
)

func toLogin(ctx http.Context, status string) http.Response {
	return ctx.Response().Redirect(http.StatusSeeOther, "/login?status="+status)
}

// Redirect is GET /auth/google/redirect: off to Google with a random state and a PKCE challenge, both remembered in
// the session (Socialite's redirect()).
func Redirect(ctx http.Context) http.Response {
	if !Enabled() {
		return toLogin(ctx, "google-unavailable")
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return toLogin(ctx, "google-failed")
	}
	state := base64.RawURLEncoding.EncodeToString(random)
	verifier := oauth2.GenerateVerifier()
	ctx.Request().Session().Put(sessionState, state).Put(sessionVerifier, verifier)
	address := oauthConfig().AuthCodeURL(state, oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "select_account"))
	return ctx.Response().Redirect(http.StatusSeeOther, address)
}

// Callback is GET /auth/google/callback: checks the state, trades the code (with the PKCE verifier) for a token,
// asks Google who it is, finds or creates the user and signs in (Socialite's user()).
func Callback(ctx http.Context) http.Response {
	session := ctx.Request().Session()
	expected, _ := session.Pull(sessionState, "").(string)
	verifier, _ := session.Pull(sessionVerifier, "").(string)
	if !Enabled() {
		return toLogin(ctx, "google-unavailable")
	}
	given := ctx.Request().Query("state")
	if ctx.Request().Query("error") != "" || expected == "" || verifier == "" ||
		subtle.ConstantTimeCompare([]byte(expected), []byte(given)) != 1 {
		return toLogin(ctx, "google-failed")
	}
	token, err := oauthConfig().Exchange(ctx.Context(), ctx.Request().Query("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		facades.Log().Request(ctx.Request()).Warning("google sign-in: code exchange failed: " + err.Error())
		return toLogin(ctx, "google-failed")
	}
	profile, err := fetchProfile(ctx.Context(), token.AccessToken)
	if err != nil {
		facades.Log().Request(ctx.Request()).Warning("google sign-in: no profile: " + err.Error())
		return toLogin(ctx, "google-failed")
	}
	user, err := FindOrCreateUser(profile, AutoCreate())
	switch {
	case errors.Is(err, ErrEmailNotVerified):
		return toLogin(ctx, "google-email-unverified")
	case errors.Is(err, ErrNoAccount):
		return toLogin(ctx, "google-no-account")
	case err != nil:
		facades.Log().Request(ctx.Request()).Error("google sign-in: " + err.Error())
		return toLogin(ctx, "google-failed")
	}
	if _, err := facades.Auth(ctx).Guard("web").Login(&user); err != nil {
		return toLogin(ctx, "google-failed")
	}
	activity.In("auth").CausedBy(user.ID).PerformedOn("user", user.ID).Event("login_with_google").LogQuietly("logged in with Google")
	return ctx.Response().Redirect(http.StatusSeeOther, "/account/profile")
}
