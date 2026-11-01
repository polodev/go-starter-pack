// Package migrations holds Spatie laravel-medialibrary's media table, column for column.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101170000CreateMediaTable is media. model_id is text so models with numeric and uuid keys both fit.
type M20261101170000CreateMediaTable struct{}

func (r *M20261101170000CreateMediaTable) Signature() string {
	return "20261101170000_create_media_table"
}

func (r *M20261101170000CreateMediaTable) Up() error {
	if facades.Schema().HasTable("media") {
		return nil
	}
	return facades.Schema().Create("media", func(table schema.Blueprint) {
		table.ID()
		table.String("model_type")
		table.String("model_id", 50)
		table.String("uuid", 36).Nullable()
		table.String("collection_name")
		table.String("name")
		table.String("file_name")
		table.String("mime_type").Nullable()
		table.String("disk")
		table.String("conversions_disk").Nullable()
		table.UnsignedBigInteger("size")
		table.LongText("manipulations")
		table.LongText("custom_properties")
		table.LongText("generated_conversions").Nullable()
		table.LongText("responsive_images")
		table.UnsignedInteger("order_column").Nullable()
		table.TimestampsTz()
		table.Unique("uuid")
		table.Index("model_type", "model_id")
		table.Index("order_column")
	})
}

func (r *M20261101170000CreateMediaTable) Down() error {
	return facades.Schema().DropIfExists("media")
}
