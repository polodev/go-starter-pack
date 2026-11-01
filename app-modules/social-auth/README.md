# social-auth

Sign in with Google for the website and the API (the Laravel Socialite idea; Goravel has no Socialite, so the
OAuth part uses `golang.org/x/oauth2`). Google accounts are linked to users in `social_accounts`.

## Setup

In the Google Cloud console create an OAuth client ("Web application"), add `{APP_URL}/auth/google/callback` as an
authorised redirect URI, then set in `.env` (read by `config/services.go`, Laravel's `config/services.php`):

```dotenv
GOOGLE_CLIENT_ID=...apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=...
# GOOGLE_REDIRECT_URL=https://example.com/auth/google/callback   # default {APP_URL}/auth/google/callback
# GOOGLE_CLIENT_IDS=ios-id.apps.googleusercontent.com,android-id.apps.googleusercontent.com
# GOOGLE_AUTO_CREATE=false   # only people who already have an account may sign in with Google
```

Without `GOOGLE_CLIENT_ID` the login page hides the button, the web routes answer "not set up" and the API 404.

## Routes

| Method | Path | What |
|---|---|---|
| GET | `/auth/google/redirect` | to Google, with a random `state` and a PKCE challenge kept in the session |
| GET | `/auth/google/callback` | checks `state`, trades the code (with the PKCE verifier), signs in |
| POST | `/api/v1/auth/google` | `{"id_token": "..."}` from a mobile app or single page app; answers `{user, token}` like `/api/v1/auth/login`; 10 a minute per address |

## Which account

1. The account already linked to this Google account (`identity` = `google:<sub>`), even if the Google email changed.
2. Else the account with the same email, which is linked now. If that account was never verified, its password is
   replaced: someone may have registered the address before its owner did and must not keep a way in.
3. Else a new account, verified, with a password nobody knows (unless `GOOGLE_AUTO_CREATE=false`).

Google must have verified the email (`email_verified`), otherwise the sign-in is refused. Each sign-in is in the
activity log (`login_with_google`). Messages on the login page come through `auth.AddLoginStatus`.
