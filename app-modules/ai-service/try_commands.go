package aiservice

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// TryTextCommand is `ai-service:try-text --prompt="..."`: one real text call, logged like any other.
type TryTextCommand struct{}

func (r *TryTextCommand) Signature() string { return "ai-service:try-text" }
func (r *TryTextCommand) Description() string {
	return "Ask the configured AI provider one question (a real, paid call)"
}
func (r *TryTextCommand) Extend() command.Extend {
	return command.Extend{Category: "ai-service", Flags: []command.Flag{
		&command.StringFlag{Name: "prompt", Value: "Say hello to a Laravel developer trying Go, in one sentence.", Usage: "the question"},
		&command.StringFlag{Name: "provider", Usage: "anthropic, openai or gemini (default AI_PROVIDER)"},
	}}
}

func (r *TryTextCommand) Handle(ctx console.Context) error {
	result, err := Chat(context.Background(), Request{Purpose: "try-out from the command line", Provider: ctx.Option("provider"), Input: ctx.Option("prompt")})
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Line(result.Text)
	ctx.Success("tokens " + strconv.Itoa(result.InputTokens) + " in / " + strconv.Itoa(result.OutputTokens) + " out, AI API log " + strconv.Itoa(int(result.LogID)))
	return nil
}

// TryImageCommand is `ai-service:try-image --prompt="..." --output=storage/app/ai-image.png`.
type TryImageCommand struct{}

func (r *TryImageCommand) Signature() string { return "ai-service:try-image" }
func (r *TryImageCommand) Description() string {
	return "Generate one picture with the configured AI provider (a real, paid call)"
}
func (r *TryImageCommand) Extend() command.Extend {
	return command.Extend{Category: "ai-service", Flags: []command.Flag{
		&command.StringFlag{Name: "prompt", Value: "A friendly gopher and an elephant building a website together, flat illustration", Usage: "what to draw"},
		&command.StringFlag{Name: "output", Value: "storage/app/ai-image.png", Usage: "where to save it"},
		&command.StringFlag{Name: "provider", Usage: "openai or gemini (default AI_PROVIDER)"},
	}}
}

func (r *TryImageCommand) Handle(ctx console.Context) error {
	result, err := Image(context.Background(), ImageRequest{Purpose: "try-out from the command line", Provider: ctx.Option("provider"), Prompt: ctx.Option("prompt"), Quality: "low"})
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	output := ctx.Option("output")
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(output, result.Bytes, 0o644); err != nil {
		return err
	}
	ctx.Success("saved " + output + " (" + result.MimeType + ", " + strconv.Itoa(len(result.Bytes)) + " bytes), AI API log " + strconv.Itoa(int(result.LogID)))
	return nil
}
