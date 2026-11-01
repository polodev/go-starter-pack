package routes

import (
	"goravel/app/facades"
)

// Web holds app-wide web routes; pages live in their modules (app-modules/*/module.go).
func Web() {
	facades.Route().Static("public", "./public")
}
