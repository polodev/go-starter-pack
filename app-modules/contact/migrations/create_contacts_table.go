// Package migrations holds the contact module's table.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261102120000CreateContactsTable is contacts: the messages sent through the contact form. The columns are the
// Rust starter's, so both can share one inbox.
type M20261102120000CreateContactsTable struct{}

func (r *M20261102120000CreateContactsTable) Signature() string {
	return "20261102120000_create_contacts_table"
}

func (r *M20261102120000CreateContactsTable) Up() error {
	if facades.Schema().HasTable("contacts") {
		return nil
	}
	return facades.Schema().Create("contacts", func(table schema.Blueprint) {
		table.ID()
		table.String("name", 80)
		table.String("email", 254)
		table.String("subject", 150).Nullable()
		table.Text("message")
		table.String("status", 16).Default("new")
		table.TimestampTz("replied_at").Nullable()
		table.String("locale", 8)
		table.String("ip_address", 45).Nullable()
		table.String("user_agent", 255).Nullable()
		table.TimestampsTz()
		table.Index("status", "created_at")
	})
}

func (r *M20261102120000CreateContactsTable) Down() error {
	return facades.Schema().DropIfExists("contacts")
}
