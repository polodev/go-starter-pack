# auth

Signing in. Web pages use Goravel's session guard (`Guard("web")`), the token API uses its JWT guard (`Guard("api")`).

| Route | |
|---|---|
| `GET/POST /register` | create an account, sign in, email a verification link |
| `GET/POST /login` | throttled: 5 tries a minute per email and address |
| `POST /logout` | ends and renews the session |
| `GET /email/verify`, `POST /email/verification-notification` | "check your inbox", send again (3 an hour) |
| `GET /email/verify/{token}` | one-time link, 24 hours |
| `GET/POST /forgot-password`, `GET/POST /reset-password/{token}` | reset link, 60 minutes |
| `GET /account/profile` | name, email, verified or not |
| `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `GET /api/v1/auth/me` | token API |

Middleware: `Authenticate` (Laravel's `auth`), `Verified` (`verified`), `Guest` (`guest`), `ApiAuthenticate` (`auth:api`).
`auth.CurrentUser(ctx)` is the signed-in person. Links store only the SHA-256 of their secret
(`user_security_tokens`); `auth:prune-tokens` cleans them daily. Mail goes through `auth.SendMail`, which queues with
Goravel's mail (tests replace it).
