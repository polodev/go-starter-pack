package socialauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"goravel/app/facades"
)

// Profile is what Google tells us about the person.
type Profile struct {
	Subject       string // Google's stable id for the account
	Email         string
	EmailVerified bool
	Name          string
}

var googleClient = &http.Client{Timeout: 10 * time.Second}

func setting(key string) string {
	return strings.TrimSpace(facades.Config().GetString("services.google." + key))
}

// Enabled is true when GOOGLE_CLIENT_ID is set.
func Enabled() bool { return setting("client_id") != "" }

// AutoCreate is GOOGLE_AUTO_CREATE: may a Google sign-in create an account?
func AutoCreate() bool { return facades.Config().GetBool("services.google.auto_create", true) }

func oauthConfig() *oauth2.Config {
	redirect := setting("redirect_url")
	if redirect == "" {
		redirect = strings.TrimSuffix(facades.Config().GetString("http.url"), "/") + "/auth/google/callback"
	}
	return &oauth2.Config{
		ClientID: setting("client_id"), ClientSecret: setting("client_secret"), RedirectURL: redirect,
		Scopes:   []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{AuthURL: setting("auth_url"), TokenURL: setting("token_url"), AuthStyle: oauth2.AuthStyleInParams},
	}
}

// boolish reads Google's email_verified, a JSON boolean in userinfo and the string "true" in tokeninfo.
type boolish bool

func (b *boolish) UnmarshalJSON(raw []byte) error {
	text := strings.Trim(string(raw), `"`)
	*b = boolish(text == "true")
	return nil
}

type googleClaims struct {
	Subject       string  `json:"sub"`
	Email         string  `json:"email"`
	EmailVerified boolish `json:"email_verified"`
	Name          string  `json:"name"`
	Audience      string  `json:"aud"`
	Expires       string  `json:"exp"`
}

func getJSON(ctx context.Context, address, bearer string, into any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	if bearer != "" {
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	response, err := googleClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("google answered %d", response.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(into)
}

// fetchProfile asks Google's userinfo endpoint who the access token belongs to (the web sign-in).
func fetchProfile(ctx context.Context, accessToken string) (Profile, error) {
	var claims googleClaims
	if err := getJSON(ctx, setting("userinfo_url"), accessToken, &claims); err != nil {
		return Profile{}, err
	}
	return claims.profile()
}

// ErrWrongAudience means the ID token was made for another app.
var ErrWrongAudience = errors.New("the Google token was issued for another app")

// verifyIDToken checks an ID token from a mobile app or single page app with Google's tokeninfo endpoint: it must be
// valid, unexpired and made for one of our client ids (GOOGLE_CLIENT_ID or GOOGLE_CLIENT_IDS).
func verifyIDToken(ctx context.Context, idToken string) (Profile, error) {
	var claims googleClaims
	if err := getJSON(ctx, setting("tokeninfo_url")+"?id_token="+url.QueryEscape(idToken), "", &claims); err != nil {
		return Profile{}, err
	}
	if !slices.Contains(allowedAudiences(), claims.Audience) {
		return Profile{}, ErrWrongAudience
	}
	return claims.profile()
}

func allowedAudiences() []string {
	audiences := []string{setting("client_id")}
	for _, id := range strings.Split(setting("client_ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			audiences = append(audiences, id)
		}
	}
	return audiences
}

func (c googleClaims) profile() (Profile, error) {
	if c.Subject == "" || c.Email == "" {
		return Profile{}, errors.New("google did not say who this is")
	}
	return Profile{Subject: c.Subject, Email: c.Email, EmailVerified: bool(c.EmailVerified), Name: c.Name}, nil
}
