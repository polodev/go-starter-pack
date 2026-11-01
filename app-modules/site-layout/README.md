# site-layout

The public layout: header (navigation, account links, language and theme switch), footer, the home page, the
language switch and the error page.

| Route | |
|---|---|
| `GET /` | home page |
| `GET /language/{locale}?next=/path` | remembers the language in a cookie for a year, then goes back (only to this site) |

Templates: `site::header`, `site::footer`, `site::home`, `site::error`. Pages of other modules wrap their content in
`{{ template "site::header" . }}` and `{{ template "site::footer" . }}`. The language comes from `?lang=`, the `lang`
cookie, then `Accept-Language` (English, Bangla).
