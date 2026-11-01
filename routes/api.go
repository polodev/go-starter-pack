package routes

import (
	"github.com/goravel/framework/contracts/route"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

// Api is the JSON API. Everything lives under /api/v1 (versioned, JSON only).
func Api() {
	auth := controllers.NewAuthController()
	facades.Route().Prefix("/api/v1").Group(func(router route.Router) {
		router.Post("/auth/register", auth.Register)
		router.Post("/auth/login", auth.Login)
		router.Middleware(&middleware.Authenticate{}).Get("/auth/me", auth.Me)
	})
}
