# queue-watcher

The queue monitor in the admin, for Goravel's database queue (like Laravel Horizon's dashboard or the Laravel jobs
monitor):

- **Job history** (`queue_jobs`): every run of every job, with its status (running, finished, failed, stale), start,
  finish, duration and error. Nothing to add to your jobs: at start-up the module wraps every registered job
  (Goravel's own mail job too), and the worker runs the wrapper. Arguments are not stored.
- **Pending jobs** (`jobs`): waiting, delayed or running; delete a waiting one.
- **Failed jobs** (`failed_jobs`): the whole error, retry one or all (Goravel's `queue:retry`), delete one or all.

| Route | Permission |
|---|---|
| `GET /admin/queue-watcher` (counts, per queue, newest runs) | `queue.view` |
| `GET /admin/queue-watcher/jobs` (`status`, `q`), `GET /admin/queue-watcher/jobs/{id}` | `queue.view` |
| `GET /admin/queue-watcher/pending` (`queue`), `GET /admin/queue-watcher/pending/{id}` | `queue.view` |
| `GET /admin/queue-watcher/failed` (`queue`), `GET /admin/queue-watcher/failed/{id}` | `queue.view` |
| `POST /admin/queue-watcher/jobs/{id}/delete`, `.../jobs/mark-stale`, `.../jobs/purge` | `queue.manage` |
| `POST /admin/queue-watcher/pending/{id}/delete` | `queue.manage` |
| `POST /admin/queue-watcher/failed/{id}/retry`, `.../failed/{id}/delete`, `.../failed/retry-all`, `.../failed/flush` | `queue.manage` |

Employees get `queue.view`, admins `queue.*`. Every change is written to the activity log (log `queue`).

`queue-watcher:purge --days=30` runs daily at 02:30: runs still "running" after `QUEUE_WATCHER_STALE_AFTER_MINUTES`
(60) become "stale" (their worker stopped mid-run), and history older than `QUEUE_WATCHER_RETENTION_DAYS` (30) goes.
Waiting and failed jobs are never touched by it.

Safety:

- The arguments of Goravel's mail job are never shown (`HiddenArgumentJobs`): a mail can hold a password reset or
  sign-in link. Add your own job there when its arguments are secret.
- A job a worker is running cannot be deleted. Retry refuses a failed job without a uuid (Goravel would otherwise
  retry them all).
- Jobs run by the `sync` connection are not recorded: they never pass through the worker.

`queue_jobs` has the Laravel jobs monitor's columns (and the Rust starter's), so either side can read the history.
