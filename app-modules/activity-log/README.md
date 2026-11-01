# activity-log

Spatie laravel-activitylog's `activity_log` table and its admin screens. Write entries from any module with
`app/activity`: `activity.In("users").CausedBy(id).PerformedOn("user", id).Event("updated").Changes(before, after)
.LogQuietly("updated user")`.

| Route | Permission |
|---|---|
| `GET /admin/activity-logs` (filters `log`, `event`, `causer`, `q`), `GET /admin/activity-logs/{id}` | `activity_logs.view` |
| `POST /admin/activity-logs/{id}/delete`, `POST /admin/activity-logs/clean` | `activity_logs.manage` |

`activity-log:clean --days=365` (default `ACTIVITY_LOG_RETENTION_DAYS`) runs daily at 02:00.
