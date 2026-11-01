package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/modular"
	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	// App migrations first (users, jobs), then each module's, in module order.
	return append([]schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20261101121130CreateUsersTable{},
	}, modular.Collect(Modules(), func(m modular.HasMigrations) []schema.Migration { return m.Migrations() })...)
}
