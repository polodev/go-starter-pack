// Package migrations holds the queue monitor's history table.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261102110000CreateQueueJobsTable is queue_jobs, the history of job runs. Its columns are those of the Laravel
// jobs monitor (and the Rust starter's queue-watcher), so the screens of either can read the other's rows.
type M20261102110000CreateQueueJobsTable struct{}

func (r *M20261102110000CreateQueueJobsTable) Signature() string {
	return "20261102110000_create_queue_jobs_table"
}

func (r *M20261102110000CreateQueueJobsTable) Up() error {
	if facades.Schema().HasTable("queue_jobs") {
		return nil // a Laravel app sharing the database may have made it
	}
	return facades.Schema().Create("queue_jobs", func(table schema.Blueprint) {
		table.ID()
		table.String("job_id")
		table.String("name").Nullable()
		table.String("job_class").Nullable()
		table.String("job_description").Nullable()
		table.String("queue").Nullable()
		table.String("connection").Nullable()
		table.TimestampTz("started_at").Nullable()
		table.TimestampTz("finished_at").Nullable()
		table.TimestampTz("failed_at").Nullable()
		table.Integer("attempt").Default(0)
		table.Integer("max_tries").Nullable()
		table.Integer("timeout").Nullable()
		table.String("tags").Nullable()
		table.Integer("progress").Nullable()
		table.String("status").Default("queued")
		table.Text("exception_message").Nullable()
		table.LongText("exception_trace").Nullable()
		table.LongText("data").Nullable()
		table.LongText("payload").Nullable()
		table.LongText("job_parameters").Nullable()
		table.TimestampsTz()
		table.Unique("job_id")
		table.Index("status", "created_at")
		table.Index("queue")
		table.Index("created_at")
		table.Index("job_class")
	})
}

func (r *M20261102110000CreateQueueJobsTable) Down() error {
	return facades.Schema().DropIfExists("queue_jobs")
}
