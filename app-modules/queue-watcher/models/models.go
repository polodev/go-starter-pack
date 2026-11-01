// Package models holds the queue monitor's history rows and read models of Goravel's jobs and failed_jobs tables.
package models

import (
	"time"

	"github.com/goravel/framework/database/orm"
)

// Statuses of a run in the history.
const (
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"
	StatusStale    = "stale" // started, never finished: the worker stopped or crashed mid-run
)

// Statuses lists them in the order the screens show them.
var Statuses = []string{StatusRunning, StatusFinished, StatusFailed, StatusStale}

// QueueJob is one run of a job (queue_jobs).
type QueueJob struct {
	orm.Model
	JobID            string     `gorm:"column:job_id"`
	Name             *string    `gorm:"column:name"`
	JobClass         *string    `gorm:"column:job_class"`
	JobDescription   *string    `gorm:"column:job_description"`
	Queue            *string    `gorm:"column:queue"`
	Connection       *string    `gorm:"column:connection"`
	StartedAt        *time.Time `gorm:"column:started_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at"`
	FailedAt         *time.Time `gorm:"column:failed_at"`
	Attempt          int        `gorm:"column:attempt"`
	MaxTries         *int       `gorm:"column:max_tries"`
	Timeout          *int       `gorm:"column:timeout"`
	Tags             *string    `gorm:"column:tags"`
	Progress         *int       `gorm:"column:progress"`
	Status           string     `gorm:"column:status"`
	ExceptionMessage *string    `gorm:"column:exception_message"`
	ExceptionTrace   *string    `gorm:"column:exception_trace"`
	Data             *string    `gorm:"column:data"`
	Payload          *string    `gorm:"column:payload"`
	JobParameters    *string    `gorm:"column:job_parameters"`
}

func (QueueJob) TableName() string { return "queue_jobs" }

// Duration is how long the run took, "" while it runs.
func (j QueueJob) Duration() string {
	if j.StartedAt == nil || j.FinishedAt == nil {
		return ""
	}
	return j.FinishedAt.Sub(*j.StartedAt).Round(time.Millisecond).String()
}

// PendingJob is a job waiting in Goravel's database queue (jobs). The worker owns these rows; the screens only read
// and delete them.
type PendingJob struct {
	ID          uint       `gorm:"column:id;primaryKey"`
	Queue       string     `gorm:"column:queue"`
	Payload     string     `gorm:"column:payload"`
	Attempts    int        `gorm:"column:attempts"`
	ReservedAt  *time.Time `gorm:"column:reserved_at"`
	AvailableAt time.Time  `gorm:"column:available_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (PendingJob) TableName() string { return "jobs" }

// State is "running" (a worker holds it), "delayed" (not due yet) or "waiting".
func (j PendingJob) State(now time.Time) string {
	switch {
	case j.ReservedAt != nil:
		return "running"
	case j.AvailableAt.After(now):
		return "delayed"
	default:
		return "waiting"
	}
}

// FailedJob is a job that used up its tries (failed_jobs).
type FailedJob struct {
	ID         uint      `gorm:"column:id;primaryKey"`
	UUID       string    `gorm:"column:uuid"`
	Connection string    `gorm:"column:connection"`
	Queue      string    `gorm:"column:queue"`
	Payload    string    `gorm:"column:payload"`
	Exception  string    `gorm:"column:exception"`
	FailedAt   time.Time `gorm:"column:failed_at"`
}

func (FailedJob) TableName() string { return "failed_jobs" }
