// Package migrations holds the pages tables.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101190000CreatePagesTables creates pages (slug, published, order) and page_translations (one row per
// language; English is required, other languages optional, empty parts NULL).
type M20261101190000CreatePagesTables struct{}

func (r *M20261101190000CreatePagesTables) Signature() string {
	return "20261101190000_create_pages_tables"
}

func (r *M20261101190000CreatePagesTables) Up() error {
	if !facades.Schema().HasTable("pages") {
		if err := facades.Schema().Create("pages", func(table schema.Blueprint) {
			table.ID()
			table.String("slug", 120)
			table.Boolean("is_published").Default(false)
			table.Integer("sort_order").Default(0)
			table.TimestampsTz()
			table.Unique("slug")
		}); err != nil {
			return err
		}
	}
	if !facades.Schema().HasTable("page_translations") {
		return facades.Schema().Create("page_translations", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("page_id")
			table.String("locale", 10)
			table.String("title").Nullable()
			table.String("description", 500).Nullable()
			table.LongText("body").Nullable()
			table.TimestampsTz()
			table.Unique("page_id", "locale")
			table.Foreign("page_id").References("id").On("pages").CascadeOnDelete()
		})
	}
	return nil
}

func (r *M20261101190000CreatePagesTables) Down() error {
	if err := facades.Schema().DropIfExists("page_translations"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("pages")
}
