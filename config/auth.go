package config

import (
	"goravel/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("auth", map[string]any{
		// "web" signs people in to the site with a session cookie; "api" uses bearer tokens (JWT) under /api/v1.
		// The default is "api" because facades.Auth(ctx) builds the default guard first, and API routes have no
		// session; web code always asks for Guard("web").
		"defaults": map[string]any{
			"guard": "api",
		},
		"guards": map[string]any{
			"web": map[string]any{
				"driver":   "session",
				"provider": "user",
			},
			"api": map[string]any{
				"driver":   "jwt",
				"provider": "user",
			},
		},
		"providers": map[string]any{
			"user": map[string]any{
				"driver": "orm",
			},
		},
	})
}
