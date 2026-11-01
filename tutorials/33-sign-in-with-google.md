<!-- nav:start -->
**Tutorial 33 of 33** | [Index](00-index.md) | [Previous: 32 Contact form](32-contact-form.md) | Next: none

<details><summary>Jump to any tutorial</summary>

[00. Index](00-index.md)  
01. [Introduction](01-introduction.md)  
02. [Setup](02-setup.md)  
03. [Go essentials for Laravel developers](03-go-essentials.md)  
04. [Go tooling jargon](04-go-tooling.md)  
05. [Project structure](05-project-structure.md)  
06. [Routes, controllers and views](06-routes-controllers-views.md)  
07. [Database: migrations, models, queries, seeding](07-database.md)  
08. [Validation and JSON responses](08-validation-and-json.md)  
09. [Token API: register, login, me](09-auth-api.md)  
10. [Queues and jobs](10-queues-and-jobs.md)  
11. [Console commands and the scheduler](11-console-commands-and-scheduler.md)  
12. [Files on Cloudflare R2](12-files-on-r2.md)  
13. [Database backups](13-database-backups.md)  
14. [Testing](14-testing.md)  
15. [Deployment with Supervisor](15-deployment.md)  
16. [MySQL or PostgreSQL](16-mysql-or-postgresql.md)  
17. [Conventions](17-conventions.md)  
18. [Cheat sheet and ideas for next](18-cheatsheet-and-next-steps.md)  
19. [Modules (app-modules)](19-modules.md)  
20. [Layouts, pages and translations](20-layouts-and-translations.md)  
21. [Web login, verification and password reset](21-web-login.md)  
22. [Roles and permissions](22-roles-and-permissions.md)  
23. [The admin area](23-admin-area.md)  
24. [Activity log](24-activity-log.md)  
25. [Media library](25-media-library.md)  
26. [Mail](26-mail.md)  
27. [PDF](27-pdf.md)  
28. [AI service](28-ai-service.md)  
29. [Bilingual pages](29-bilingual-pages.md)  
30. [Security headers](30-security-headers.md)  
31. [Queue monitor](31-queue-monitor.md)  
32. [Contact form](32-contact-form.md)  
**33. [Sign in with Google](33-sign-in-with-google.md)**  

</details>
<!-- nav:end -->

# 33. Sign in with Google

`app-modules/social-auth` is Laravel Socialite's Google driver, written out. Laravel:

```php
Route::get('/auth/google/redirect', fn () => Socialite::driver('google')->redirect());
Route::get('/auth/google/callback', function () {
    $google = Socialite::driver('google')->user();
    $user = User::firstOrCreate(['email' => $google->email], [...]);
    Auth::login($user);
});
```

Here (`web.go`):

```go
func Redirect(ctx http.Context) http.Response {
	state, verifier := random(), oauth2.GenerateVerifier()
	ctx.Request().Session().Put("google_oauth_state", state).Put("google_oauth_verifier", verifier)
	return redirect(oauthConfig().AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)))
}

func Callback(ctx http.Context) http.Response {
	// state from the session must equal ?state= (constant-time), then:
	token, _ := oauthConfig().Exchange(ctx.Context(), code, oauth2.VerifierOption(verifier))
	profile, _ := fetchProfile(ctx.Context(), token.AccessToken)   // Google's userinfo
	user, _ := FindOrCreateUser(profile, AutoCreate())
	facades.Auth(ctx).Guard("web").Login(&user)
}
```

## The safety parts, and why

| Part | Stops |
|---|---|
| `state` in the session, checked once | someone tricking you into finishing *their* Google sign-in (login CSRF) |
| PKCE (`code_challenge`, `code_verifier`) | a stolen authorization code being used elsewhere |
| `email_verified` required | an address Google never checked being used to take over an account |
| unverified local account loses its password when linked | someone who registered your address before you keeping a way in |
| link by Google's `sub`, not by email | a changed Google email creating a second account |
| API: `aud` must be one of our client ids | an ID token made for another app being replayed here |

## The API, for mobile apps

The app signs in with the Google Sign-In SDK and sends the ID token:

```http
POST /api/v1/auth/google
{"id_token": "eyJhbGciOi..."}
```

The server asks Google's `tokeninfo` endpoint whether the token is valid and for whom (`aud` must be
`GOOGLE_CLIENT_ID` or one of `GOOGLE_CLIENT_IDS`), then answers `{"user": ..., "token": ...}`, the same as
`/api/v1/auth/login`. Errors are Laravel's 422 shape.

## Settings

`config/services.go` (`services.google.*`) from `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`,
`GOOGLE_CLIENT_IDS`, `GOOGLE_AUTO_CREATE`. See the module README for the Google Cloud console steps.

## What the tests prove

`tests/feature/social_auth_test.go` runs against a fake Google (`httptest`), so no keys are needed: the redirect
carries the client id, scope, state and a PKCE challenge; a new person gets a verified, linked account and is signed
in; the second sign-in finds them by `sub` even with a new email; an existing account is linked, and an unverified one
loses its old password while a verified one keeps it; a forged, missing or reused state, a bad code or Google's
"access denied" signs nobody in; an unverified Google email and "no account" with auto-create off are refused;
without a client id the button is gone; the API accepts our ID token, refuses another app's or a broken one, and
accepts mobile client ids from `GOOGLE_CLIENT_IDS`.

<!-- foot:start -->
---
[Previous: 32 Contact form](32-contact-form.md) | [Index](00-index.md) | Next: none
<!-- foot:end -->
