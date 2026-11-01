// Package aiservice is the one way to call AI models (Claude, OpenAI, Gemini) in this app: it uses Goravel's AI
// SDK (facades.AI, providers from config/ai.go) and records every call in the AI API logs with its purpose,
// tokens, duration and estimated cost. Sensitive calls keep their prompt and reply out of the logs.
package aiservice

import (
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/database/schema"
	contractsschedule "github.com/goravel/framework/contracts/schedule"

	"goravel/app-modules/ai-service/migrations"
	rolepermission "goravel/app-modules/role-permission"
	"goravel/app/facades"
	"goravel/app/modular"
)

// Module registers the AI API logs table, its screens and the purge command.
type Module struct{}

func (m *Module) Name() string      { return "ai-service" }
func (m *Module) ViewsPath() string { return "app-modules/ai-service/resources/views" }

func (m *Module) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20261101180000CreateAiApiLogsTable{},
		&migrations.M20261102100000AddKindToAiApiLogsTable{},
	}
}

func (m *Module) Commands() []console.Command {
	return []console.Command{&PurgeCommand{}, &TryTextCommand{}, &TryImageCommand{}}
}

func (m *Module) Schedule() []contractsschedule.Event {
	return []contractsschedule.Event{facades.Schedule().Command("ai-service:purge-ai-api-logs").DailyAt("02:30").OnOneServer()}
}

func (m *Module) AdminMenu() []modular.MenuItem {
	return []modular.MenuItem{{Section: "System", Label: "AI API logs", Href: "/admin/ai-api-logs", Permission: "ai_api_logs.view", Icon: "✦", Order: 2}}
}

func (m *Module) Routes() {
	rolepermission.Admin(func(admin *rolepermission.AdminRouter) {
		admin.Get("/ai-api-logs", "ai_api_logs.view", Index)
		admin.Get("/ai-api-logs/{id}", "ai_api_logs.view", Show)
		admin.Post("/ai-api-logs/{id}/delete", "ai_api_logs.manage", Delete)
	})
}

// PurgeCommand is `ai-service:purge-ai-api-logs --days=90`.
type PurgeCommand struct{}

func (r *PurgeCommand) Signature() string { return "ai-service:purge-ai-api-logs" }
func (r *PurgeCommand) Description() string {
	return "Delete AI API logs older than N days (default AI_LOG_RETENTION_DAYS or 90)"
}
func (r *PurgeCommand) Extend() command.Extend {
	return command.Extend{Category: "ai-service", Flags: []command.Flag{&command.IntFlag{Name: "days", Usage: "keep this many days"}}}
}

func (r *PurgeCommand) Handle(ctx console.Context) error {
	days := ctx.OptionInt("days")
	if days <= 0 {
		days = facades.Config().GetInt("ai.retention_days", 90)
		if text, ok := facades.Config().Env("AI_LOG_RETENTION_DAYS", "").(string); ok && text != "" {
			if parsed, err := strconv.Atoi(text); err == nil && parsed > 0 {
				days = parsed
			}
		}
	}
	removed, err := PurgeOlderThan(days)
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("removed " + strconv.FormatInt(removed, 10) + " AI API log(s) older than " + strconv.Itoa(days) + " days")
	return nil
}
