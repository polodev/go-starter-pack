package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261101121130CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20261101121130CreateUsersTable) Signature() string {
	return "20261101121130_create_users_table"
}

// Up Run the migrations.
func (r *M20261101121130CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			// Laravel's users table, so a Laravel export imports as is (bcrypt passwords included).
			table.ID()
			table.String("name")
			table.String("email")
			table.TimestampTz("email_verified_at").Nullable()
			table.String("password")
			table.String("remember_token", 100).Nullable()
			table.TimestampsTz()
			table.Unique("email")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20261101121130CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
