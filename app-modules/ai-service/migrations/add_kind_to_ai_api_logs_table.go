package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261102100000AddKindToAiApiLogsTable adds kind (text or image) to ai_api_logs.
type M20261102100000AddKindToAiApiLogsTable struct{}

func (r *M20261102100000AddKindToAiApiLogsTable) Signature() string {
	return "20261102100000_add_kind_to_ai_api_logs_table"
}

func (r *M20261102100000AddKindToAiApiLogsTable) Up() error {
	if facades.Schema().HasColumn("ai_api_logs", "kind") {
		return nil
	}
	return facades.Schema().Table("ai_api_logs", func(table schema.Blueprint) {
		table.String("kind", 20).Default("text")
	})
}

func (r *M20261102100000AddKindToAiApiLogsTable) Down() error {
	if !facades.Schema().HasColumn("ai_api_logs", "kind") {
		return nil
	}
	return facades.Schema().Table("ai_api_logs", func(table schema.Blueprint) {
		table.DropColumn("kind")
	})
}
