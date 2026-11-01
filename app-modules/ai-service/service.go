package aiservice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goravelai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"

	"goravel/app-modules/ai-service/models"
	"goravel/app/facades"
)

// Request is one question to an AI model. Purpose is required: the AI API logs say what each call was for.
type Request struct {
	Purpose      string // e.g. "summarise contact message"
	Instructions string // the system prompt
	Input        string
	Provider     string // "" = AI_PROVIDER
	Model        string // "" = the provider's configured model
	Sensitive    bool   // about a person (IDs, health, faces): the prompt and reply are not stored
}

// Result is the model's answer and what the call cost.
type Result struct {
	Text         string
	InputTokens  int
	OutputTokens int
	LogID        uint
}

// agent is a one-off Goravel AI agent carrying this call's instructions and the logging middleware.
type agent struct {
	instructions string
	logger       *callLogger
}

func (a *agent) Instructions() string                 { return a.instructions }
func (a *agent) Messages() []contractsai.Message      { return nil }
func (a *agent) Middleware() []contractsai.Middleware { return []contractsai.Middleware{a.logger} }
func (a *agent) Tools() []contractsai.Tool            { return nil }

// Chat asks a model through Goravel's AI SDK (facades.AI) and records the call in the AI API logs.
func Chat(ctx context.Context, request Request) (Result, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return Result{}, errors.New("every AI call needs a purpose")
	}
	provider := request.Provider
	if provider == "" {
		provider = facades.Config().GetString("ai.default")
	}
	logger := &callLogger{request: request, provider: provider}
	options := []contractsai.Option{goravelai.WithProvider(provider)}
	if request.Model != "" {
		options = append(options, goravelai.WithModel(request.Model))
	}
	conversation, err := facades.AI().WithContext(ctx).Agent(&agent{instructions: request.Instructions, logger: logger}, options...)
	if err != nil {
		logger.record(request.Model, time.Now(), nil, err)
		return Result{LogID: logger.logID}, err
	}
	response, err := conversation.Prompt(request.Input)
	if err != nil {
		return Result{LogID: logger.logID}, err
	}
	usage := response.Usage()
	result := Result{Text: response.Text(), LogID: logger.logID}
	if usage != nil {
		result.InputTokens, result.OutputTokens = usage.Input(), usage.Output()
	}
	return result, nil
}

// JSON asks for a JSON reply and decodes it into target. A model's reply is input, not truth: check the fields.
func JSON(ctx context.Context, request Request, target any) (Result, error) {
	request.Instructions = strings.TrimSpace(request.Instructions + "\n\nReply with JSON only: no prose, no code fences.")
	result, err := Chat(ctx, request)
	if err != nil {
		return result, err
	}
	text := strings.TrimSpace(result.Text)
	text = strings.TrimPrefix(strings.TrimPrefix(text, "```json"), "```")
	text = strings.TrimSpace(strings.TrimSuffix(text, "```"))
	if err := json.Unmarshal([]byte(text), target); err != nil {
		return result, errors.New("the model did not answer with the expected JSON: " + err.Error())
	}
	return result, nil
}

// callLogger is a Goravel AI middleware: it times the call and writes the AI API log row.
type callLogger struct {
	request  Request
	provider string
	logID    uint
}

func (l *callLogger) Handle(ctx context.Context, prompt contractsai.AgentPrompt, next contractsai.Next) (contractsai.AgentResponse, error) {
	started := time.Now()
	response, err := next(ctx, prompt)
	l.record(prompt.Model, started, response, err)
	return response, err
}

func (l *callLogger) record(model string, started time.Time, response contractsai.AgentResponse, callErr error) {
	if model == "" {
		model = l.request.Model
	}
	if model == "" {
		model = facades.Config().GetString("ai.providers." + l.provider + ".models.text.default")
	}
	row := models.AiApiLog{
		Provider: l.provider, ModelName: model, Purpose: truncate(l.request.Purpose, 100), Status: models.Success,
		DurationMs: uint(time.Since(started).Milliseconds()), Sensitive: l.request.Sensitive,
	}
	if response != nil && response.Usage() != nil {
		row.InputTokens, row.OutputTokens = uint(response.Usage().Input()), uint(response.Usage().Output())
		row.Cost = EstimateCost(model, row.InputTokens, row.OutputTokens)
	}
	if !l.request.Sensitive && LogContent() {
		prompt := l.request.Input
		row.Prompt = &prompt
		if response != nil {
			text := response.Text()
			row.Response = &text
		}
	}
	if callErr != nil {
		row.Status = models.Failed
		message := truncate(callErr.Error(), 2000)
		row.Error = &message
	}
	if err := facades.Orm().Query().Create(&row); err != nil {
		facades.Log().Error("AI API log not written: " + err.Error())
		return
	}
	l.logID = row.ID
}

// LogContent is AI_LOG_CONTENT (default true): store prompts and replies of calls that are not sensitive.
func LogContent() bool {
	value, _ := facades.Config().Env("AI_LOG_CONTENT", "true").(string)
	return value != "false" && value != "0"
}

func truncate(text string, limit int) string {
	if len(text) > limit {
		return text[:limit]
	}
	return text
}
