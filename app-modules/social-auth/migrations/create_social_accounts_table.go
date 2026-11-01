// Package migrations holds the social-auth module's table.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261102130000CreateSocialAccountsTable is social_accounts: which outside account (identity "google:<sub>")
// belongs to which user. The columns are the Rust starter's.
type M20261102130000CreateSocialAccountsTable struct{}

func (r *M20261102130000CreateSocialAccountsTable) Signature() string {
	return "20261102130000_create_social_accounts_table"
}

func (r *M20261102130000CreateSocialAccountsTable) Up() error {
	if facades.Schema().HasTable("social_accounts") {
		return nil
	}
	return facades.Schema().Create("social_accounts", func(table schema.Blueprint) {
		table.ID()
		table.String("provider")
		table.String("identity")
		table.String("email")
		table.UnsignedBigInteger("user_id")
		table.TimestampsTz()
		table.Unique("identity")
		table.Index("user_id")
		table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
	})
}

func (r *M20261102130000CreateSocialAccountsTable) Down() error {
	return facades.Schema().DropIfExists("social_accounts")
}
