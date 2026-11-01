# pages

Bilingual website pages (`pages`, `page_translations`).

- English is required; Bangla is optional; each empty part (title, description, text) falls back to English on its
  own, and clearing a whole language deletes its row. Drafts are hidden.
- Staff HTML is cleaned with bluemonday before it is stored.
- `GET /pages/{slug}` shows a published page in the visitor's language.
- Admin: `GET /admin/pages`, `/admin/pages/{id}` (`pages.view`); create and edit (`pages.manage`).
- `go run . artisan db:seed --force` adds About, Privacy policy and Terms of service in both languages.
