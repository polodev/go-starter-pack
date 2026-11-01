// Package migrations holds Spatie laravel-activitylog's table, column for column.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101160000CreateActivityLogTable is activity_log. The subject and causer ids are text, so models with
// numeric and uuid keys can both be recorded.
type M20261101160000CreateActivityLogTable struct{}

func (r *M20261101160000CreateActivityLogTable) Signature() string {
	return "20261101160000_create_activity_log_table"
}

func (r *M20261101160000CreateActivityLogTable) Up() error {
	if facades.Schema().HasTable("activity_log") {
		return nil // a Laravel app sharing the database may have made it
	}
	return facades.Schema().Create("activity_log", func(table schema.Blueprint) {
		table.ID()
		table.String("log_name").Nullable()
		table.Text("description")
		table.String("subject_type").Nullable()
		table.String("event").Nullable()
		table.String("subject_id", 50).Nullable()
		table.String("causer_type").Nullable()
		table.String("causer_id", 50).Nullable()
		table.LongText("properties").Nullable()
		table.String("batch_uuid", 36).Nullable()
		table.TimestampsTz()
		table.Index("subject_type", "subject_id")
		table.Index("causer_type", "causer_id")
		table.Index("log_name")
		table.Index("created_at")
	})
}

// Down drops the table, like Spatie's own migration (rollbacks and test refreshes rely on it).
func (r *M20261101160000CreateActivityLogTable) Down() error {
	return facades.Schema().DropIfExists("activity_log")
}
