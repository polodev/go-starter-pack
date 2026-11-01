// Package migrations holds the AI API logs table.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101180000CreateAiApiLogsTable is ai_api_logs: one row per call to an AI provider.
type M20261101180000CreateAiApiLogsTable struct{}

func (r *M20261101180000CreateAiApiLogsTable) Signature() string {
	return "20261101180000_create_ai_api_logs_table"
}

func (r *M20261101180000CreateAiApiLogsTable) Up() error {
	if facades.Schema().HasTable("ai_api_logs") {
		return nil
	}
	return facades.Schema().Create("ai_api_logs", func(table schema.Blueprint) {
		table.ID()
		table.String("provider", 20)
		table.String("model", 100)
		table.String("purpose", 100)
		table.String("status", 20) // success, failed
		table.UnsignedInteger("duration_ms").Default(0)
		table.UnsignedInteger("input_tokens").Default(0)
		table.UnsignedInteger("output_tokens").Default(0)
		table.Double("cost").Nullable() // estimate in USD, when a price is set
		table.Boolean("sensitive").Default(false)
		table.LongText("prompt").Nullable()
		table.LongText("response").Nullable()
		table.Text("error").Nullable()
		table.TimestampsTz()
		table.Index("status", "created_at")
		table.Index("provider")
		table.Index("created_at")
	})
}

func (r *M20261101180000CreateAiApiLogsTable) Down() error {
	return facades.Schema().DropIfExists("ai_api_logs")
}
