package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/app/modular"
)

// Seeders are every module's seeders; `go run . artisan db:seed` runs them all.
func Seeders() []seeder.Seeder {
	return modular.Collect(Modules(), func(m modular.HasSeeders) []seeder.Seeder { return m.Seeders() })
}
