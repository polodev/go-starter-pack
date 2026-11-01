package config

import (
	anthropicfacades "github.com/goravel/anthropic/facades"
	"github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/support"
	geminifacades "github.com/goravel/gemini/facades"
	openaifacades "github.com/goravel/openai/facades"

	"goravel/app/facades"
	"goravel/app/support/aifake"
)

func init() {
	config := facades.Config()
	config.Add("ai", map[string]any{
		// Default AI Provider
		//
		// This option controls the default AI provider that will be used.
		"default": config.Env("AI_PROVIDER", "anthropic"),

		// AI Providers
		//
		// Here you may configure each AI provider used by your application.
		// A variety of drivers are available, and each provider may also
		// configure the models available to your application.
		"providers": map[string]any{
			"openai": map[string]any{
				"key": config.Env("OPENAI_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default": config.Env("OPENAI_MODEL", ""),
					},
					"audio": map[string]any{
						"default": "",
					},
					"transcription": map[string]any{
						"default": "",
					},
					"image": map[string]any{
						"default": config.Env("OPENAI_IMAGE_MODEL", ""),
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("OPENAI_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return openaifacades.OpenAI("openai")
				},
			},
			"anthropic": map[string]any{
				"key": config.Env("ANTHROPIC_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default":    config.Env("CLAUDE_MODEL", ""),
						"max_tokens": 4096,
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("ANTHROPIC_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return anthropicfacades.Anthropic("anthropic")
				},
			},
			"gemini": map[string]any{
				"key": config.Env("GEMINI_API_KEY", ""),
				"models": map[string]any{
					"text": map[string]any{
						"default": config.Env("GEMINI_MODEL", ""),
					},
					"audio": map[string]any{
						"default": "",
					},
					"transcription": map[string]any{
						"default": "",
					},
					"image": map[string]any{
						"default": "",
					},
				},
				"failover": map[string][]string{},
				"url":      config.Env("GEMINI_BASE_URL", ""),
				"via": func() (ai.Provider, error) {
					return geminifacades.Gemini("gemini")
				},
			},
		},
	})

	// Under `go test`, provider "fake" answers instantly (app/support/aifake), so tests never call a paid API.
	if support.RuntimeMode == support.RuntimeTest {
		config.Add("ai.providers.fake", map[string]any{
			"key":    "test",
			"models": map[string]any{"text": map[string]any{"default": "fake-model-1"}, "image": map[string]any{"default": "fake-image-1"}},
			"via":    func() (ai.Provider, error) { return aifake.Provider{}, nil },
		})
	}
}
