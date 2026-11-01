// Package migrations holds the role and permission tables, the same columns as the Laravel role-permission module.
package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

// M20261101150000CreateRolePermissionTables creates roles, permissions, permission_role, role_user and
// permission_user (direct exceptions for one person).
type M20261101150000CreateRolePermissionTables struct{}

func (r *M20261101150000CreateRolePermissionTables) Signature() string {
	return "20261101150000_create_role_permission_tables"
}

func (r *M20261101150000CreateRolePermissionTables) Up() error {
	schemaFacade := facades.Schema()
	steps := []struct {
		table  string
		create func(table schema.Blueprint)
	}{
		{"roles", func(table schema.Blueprint) {
			table.ID()
			table.String("slug")
			table.String("name")
			table.Integer("rank").Default(0)
			table.Text("description").Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.Unique("slug")
			table.Index("is_active")
		}},
		{"permissions", func(table schema.Blueprint) {
			table.ID()
			table.String("key")
			table.String("name")
			table.String("group")
			table.Text("description").Nullable()
			table.Boolean("is_active").Default(true)
			table.TimestampsTz()
			table.Unique("key")
			table.Index("group")
			table.Index("is_active")
		}},
		{"permission_role", func(table schema.Blueprint) {
			table.UnsignedBigInteger("role_id")
			table.UnsignedBigInteger("permission_id")
			table.TimestampsTz()
			table.Primary("role_id", "permission_id")
			table.Foreign("role_id").References("id").On("roles").CascadeOnDelete()
			table.Foreign("permission_id").References("id").On("permissions").CascadeOnDelete()
		}},
		{"role_user", func(table schema.Blueprint) {
			table.UnsignedBigInteger("user_id")
			table.UnsignedBigInteger("role_id")
			table.TimestampsTz()
			table.Primary("user_id", "role_id")
			table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
			table.Foreign("role_id").References("id").On("roles").CascadeOnDelete()
		}},
		{"permission_user", func(table schema.Blueprint) {
			table.UnsignedBigInteger("user_id")
			table.UnsignedBigInteger("permission_id")
			table.TimestampsTz()
			table.Primary("user_id", "permission_id")
			table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
			table.Foreign("permission_id").References("id").On("permissions").CascadeOnDelete()
		}},
	}
	for _, step := range steps {
		if schemaFacade.HasTable(step.table) {
			continue
		}
		if err := schemaFacade.Create(step.table, step.create); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20261101150000CreateRolePermissionTables) Down() error {
	for _, table := range []string{"permission_user", "role_user", "permission_role", "permissions", "roles"} {
		if err := facades.Schema().DropIfExists(table); err != nil {
			return err
		}
	}
	return nil
}
