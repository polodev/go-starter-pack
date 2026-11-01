package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"

	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
	"goravel/app/modular"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).
		WithSchedule(Schedule).
		WithCommands(Commands).
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			for _, module := range Modules() {
				if routable, ok := module.(modular.HasRoutes); ok {
					routable.Routes()
				}
			}
			routes.Grpc()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		WithCallback(bootModules).
		Create()
}

// bootModules registers every module's views (Laravel's loadViewsFrom). Translations come through
// config app.lang_fs (modular.LangFS).
func bootModules() {
	for _, module := range Modules() {
		if views, ok := module.(modular.HasViews); ok {
			facades.View().LoadViewsFrom(path.Base(views.ViewsPath()))
		}
	}
}
