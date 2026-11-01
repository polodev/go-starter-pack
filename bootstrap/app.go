package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsconfiguration "github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"

	"github.com/goravel/framework/support/path"

	"goravel/app/facades"
	"goravel/app/http/middleware"
	"goravel/app/modular"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).
		WithSchedule(Schedule).
		WithSeeders(Seeders).
		WithCommands(Commands).
		WithMigrations(Migrations).
		WithMiddleware(func(handler contractsconfiguration.Middleware) {
			// Global: every page, API answer, static file and 404 gets the browser security headers.
			handler.Append(&middleware.SecurityHeaders{})
		}).
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

// bootModules registers every module's views (Laravel's loadViewsFrom), then runs each module's Boot. Translations
// come through config app.lang_fs (modular.LangFS).
func bootModules() {
	modular.Register(Modules())
	for _, module := range Modules() {
		if views, ok := module.(modular.HasViews); ok {
			facades.View().LoadViewsFrom(path.Base(views.ViewsPath()))
		}
	}
	for _, module := range Modules() {
		if booting, ok := module.(modular.HasBoot); ok {
			booting.Boot()
		}
	}
}
