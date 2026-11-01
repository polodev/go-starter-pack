package bootstrap

import (
	activitylog "goravel/app-modules/activity-log"
	adminlayout "goravel/app-modules/admin-layout"
	aiservice "goravel/app-modules/ai-service"
	auth "goravel/app-modules/auth"
	media "goravel/app-modules/media"
	pages "goravel/app-modules/pages"
	pdf "goravel/app-modules/pdf"
	rolepermission "goravel/app-modules/role-permission"
	sitelayout "goravel/app-modules/site-layout"
	"goravel/app/modular"
)

// Modules lists the feature modules in app-modules/ (Laravel discovers them through Composer; Go needs the list).
// `go run . artisan make:module <name>` adds new ones above the marker.
func Modules() []modular.Module {
	return []modular.Module{
		&sitelayout.Module{},
		&auth.Module{},
		&rolepermission.Module{},
		&adminlayout.Module{},
		&activitylog.Module{},
		&media.Module{},
		&pdf.Module{},
		&aiservice.Module{},
		&pages.Module{},
		// modules-inject (do not remove)
	}
}
