// Package migrations holds the auth module's tables.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101140000CreateUserSecurityTokensTable keeps the one-time links sent by email (verify the address, reset
// the password). Only the SHA-256 of the secret is stored.
type M20261101140000CreateUserSecurityTokensTable struct{}

func (r *M20261101140000CreateUserSecurityTokensTable) Signature() string {
	return "20261101140000_create_user_security_tokens_table"
}

func (r *M20261101140000CreateUserSecurityTokensTable) Up() error {
	if facades.Schema().HasTable("user_security_tokens") {
		return nil
	}
	return facades.Schema().Create("user_security_tokens", func(table schema.Blueprint) {
		table.ID()
		table.UnsignedBigInteger("user_id")
		table.String("purpose", 40)
		table.String("token_hash", 64)
		table.TimestampTz("expires_at")
		table.TimestampTz("used_at").Nullable()
		table.TimestampsTz()
		table.Unique("token_hash")
		table.Index("user_id", "purpose")
	})
}

func (r *M20261101140000CreateUserSecurityTokensTable) Down() error {
	return facades.Schema().DropIfExists("user_security_tokens")
}
