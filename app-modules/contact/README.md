# contact

The public contact form and the admin inbox.

| Route | Who |
|---|---|
| `GET /contact`, `POST /contact` | everyone (web group: session, CSRF, language) |
| `GET /admin/contacts` (`status`, `q` over name, email and subject), `GET /admin/contacts/{id}` | `contacts.view` |
| `POST /admin/contacts/{id}/status` (`new` or `replied`), `POST /admin/contacts/{id}/delete` | `contacts.manage` |

Business development reads the inbox (`contacts.view`); employees and above answer and delete (`contacts.manage`).
Status changes and deletions go to the activity log (log `contacts`); a deletion entry holds no name or email,
because a deletion may be someone asking to be forgotten.

The form, against spam and mistakes:

- **CSRF**: the `_token` field (419 without it).
- **Honeypot**: a field called `website`, moved off-screen with classes (no `style=`: the Content-Security-Policy
  blocks inline styles). When it is filled the visitor sees the usual thank-you and nothing is stored.
- **Rate limit**: 5 posts per 10 minutes per address (`facades.RateLimiter()` limiter `contact`); the sixth gets 429
  with the typed message kept.
- **Validation** on the server, lengths in characters (Bangla counts like English): name 2 to 80, a real email
  address, subject up to 150 (optional), message 10 to 5000. Errors answer 422 with the typed values kept.
- Everything a visitor reads is a translation key in `lang/en.json` and `lang/bn.json`; the message stores the
  visitor's language.

`CONTACT_NOTIFICATION_EMAIL` gets a plain-text mail (queued) for each new message; unset, nothing is sent.
`contacts` has the Rust starter's columns, so both can share one inbox.
