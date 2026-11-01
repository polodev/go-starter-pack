package bootstrap

import (
	activitylog "goravel/app-modules/activity-log"
	adminlayout "goravel/app-modules/admin-layout"
	auth "goravel/app-modules/auth"
	media "goravel/app-modules/media"
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
		// modules-inject (do not remove)
	}
}
