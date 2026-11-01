package queuewatcher

import (
	"time"

	"github.com/google/uuid"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/spf13/cast"

	"goravel/app-modules/queue-watcher/models"
	"goravel/app/facades"
)

// trackedJob runs a job and writes the run to queue_jobs: started, then finished or failed with the error.
// Arguments are not stored: they can hold mail content, links or personal data.
type trackedJob struct {
	inner contractsqueue.Job
}

func (t *trackedJob) Signature() string { return t.inner.Signature() }

func (t *trackedJob) Handle(args ...any) error {
	started := time.Now().UTC()
	signature := t.inner.Signature()
	connection := facades.Config().GetString("queue.default")
	entry := models.QueueJob{
		JobID: uuid.NewString(), Name: &signature, JobClass: &signature, Connection: &connection,
		StartedAt: &started, Attempt: 1, Status: models.StatusRunning,
	}
	if err := facades.Orm().Query().Create(&entry); err != nil {
		// The history is a convenience: a database hiccup must not stop the job itself.
		facades.Log().Warning("queue-watcher: could not record " + signature + ": " + err.Error())
		return t.inner.Handle(args...)
	}

	jobErr := t.inner.Handle(args...)

	finished := time.Now().UTC()
	changes := map[string]any{"finished_at": finished, "status": models.StatusFinished, "updated_at": finished}
	if jobErr != nil {
		changes["status"] = models.StatusFailed
		changes["failed_at"] = finished
		changes["exception_message"] = truncate(jobErr.Error(), 60000)
	}
	if _, err := facades.Orm().Query().Model(&models.QueueJob{}).Where("id", entry.ID).Update(changes); err != nil {
		facades.Log().Warning("queue-watcher: could not finish the record of " + signature + ": " + err.Error())
	}
	return jobErr
}

// Track wraps one job so its runs show up in the queue monitor. Wrapping twice is harmless.
func Track(job contractsqueue.Job) contractsqueue.Job {
	if _, already := job.(*trackedJob); already {
		return job
	}
	return &trackedJob{inner: job}
}

// TrackAll wraps every job registered so far, Goravel's own mail job included. The worker finds a job by its
// signature, so it runs the wrapper. bootstrap calls this after every module has registered its jobs; a test that
// registers jobs later calls it again. (The sync driver runs the job it was given, so sync jobs are not tracked.)
func TrackAll() {
	storer := facades.Queue().JobStorer()
	jobs := storer.All()
	wrapped := make([]contractsqueue.Job, 0, len(jobs))
	for _, job := range jobs {
		wrapped = append(wrapped, Track(job))
	}
	storer.Register(wrapped)
}

// MarkStale marks runs that started more than olderThan ago and never finished: their worker stopped mid-run.
func MarkStale(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan)
	result, err := facades.Orm().Query().Model(&models.QueueJob{}).Where("status", models.StatusRunning).
		Where("started_at < ?", cutoff).Update(map[string]any{"status": models.StatusStale, "updated_at": time.Now().UTC()})
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

// PurgeHistory deletes finished, failed and stale runs older than days; running ones stay.
func PurgeHistory(days int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	result, err := facades.Orm().Query().Where("created_at < ?", cutoff).Where("status <> ?", models.StatusRunning).
		Delete(&models.QueueJob{})
	if err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

// RetentionDays is QUEUE_WATCHER_RETENTION_DAYS (default 30): how long the history keeps a run.
func RetentionDays() int { return positiveEnv("QUEUE_WATCHER_RETENTION_DAYS", 30) }

// StaleAfter is QUEUE_WATCHER_STALE_AFTER_MINUTES (default 60): a run that long without finishing is stale.
func StaleAfter() time.Duration {
	return time.Duration(positiveEnv("QUEUE_WATCHER_STALE_AFTER_MINUTES", 60)) * time.Minute
}

func positiveEnv(name string, fallback int) int {
	if value := cast.ToInt(facades.Config().Env(name, fallback)); value > 0 {
		return value
	}
	return fallback
}

func truncate(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	return text[:limit]
}
