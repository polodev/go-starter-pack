# media

A Spatie laravel-medialibrary compatible media library (`media` table).

```go
media.Define("user", media.Single("avatar").Images().AsPublic().Max(2<<20))
item, err := media.Add(media.UserOwner(user.ID), "avatar", file) // or AddBytes for generated files
url := media.URL(item)
```

- Collections are single (a new file replaces the old) or multiple (`KeepNewest(n)`), with accepted types and a size
  limit (default 10 MB). Files sit at `<MEDIA_PREFIX>/<id>/<file_name>`.
- Public files are served from their public address; private ones only to the owner and staff with `media.view`
  (unknown and forbidden both answer 404).
- `MEDIA_STORAGE=local` keeps files on this server; `r2` uses `r2_public` and `r2_private` (`R2_PRIVATE_BUCKET`).

| Route | |
|---|---|
| `GET /media/{id}/{file_name}` | serves a file after the access check |
| `GET/POST /account/avatar`, `POST /account/avatar/delete` | the person's picture |
| `GET /admin/media` (`media.view`), `POST /admin/media`, `POST /admin/media/{id}/delete` (`media.manage`) | admin |
