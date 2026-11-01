package aiservice

import (
	"context"
	"errors"
	"strings"
	"time"

	contractsai "github.com/goravel/framework/contracts/ai"

	"goravel/app-modules/ai-service/models"
	"goravel/app/facades"
)

// ImageRequest asks a model for a picture. Purpose is required, like every AI call.
type ImageRequest struct {
	Purpose   string
	Prompt    string
	Provider  string // "" = AI_PROVIDER (it must make images: openai or gemini)
	Model     string // "" = the provider's image model (OPENAI_IMAGE_MODEL)
	Size      contractsai.ImageSize
	Quality   contractsai.ImageQuality
	Sensitive bool // the prompt is not stored
}

// ImageResult is the picture: its bytes and type. Pictures are never stored in the AI API logs.
type ImageResult struct {
	Bytes    []byte
	MimeType string
	LogID    uint
}

// Image generates a picture through Goravel's AI SDK (facades.AI().Image) and records the call.
func Image(ctx context.Context, request ImageRequest) (ImageResult, error) {
	if strings.TrimSpace(request.Purpose) == "" {
		return ImageResult{}, errors.New("every AI call needs a purpose")
	}
	provider := request.Provider
	if provider == "" {
		provider = facades.Config().GetString("ai.default")
	}
	model := request.Model
	if model == "" {
		model = facades.Config().GetString("ai.providers." + provider + ".models.image.default")
	}
	call := facades.AI().WithContext(ctx).Image(request.Prompt).Provider(provider)
	if model != "" {
		call = call.Model(model)
	}
	switch request.Size {
	case contractsai.ImageSizePortrait:
		call = call.Portrait()
	case contractsai.ImageSizeLandscape:
		call = call.Landscape()
	default:
		call = call.Square()
	}
	if request.Quality != "" {
		call = call.Quality(request.Quality)
	}

	started := time.Now()
	response, err := call.Generate()
	row := models.AiApiLog{
		Kind: "image", Provider: provider, ModelName: model, Purpose: truncate(request.Purpose, 100), Status: models.Success,
		DurationMs: uint(time.Since(started).Milliseconds()), Sensitive: request.Sensitive,
	}
	if !request.Sensitive && LogContent() {
		prompt := request.Prompt
		row.Prompt = &prompt
	}
	var result ImageResult
	if err == nil {
		result.Bytes, err = response.Content()
		result.MimeType = response.MimeType()
		if usage := response.Usage(); usage != nil {
			row.InputTokens, row.OutputTokens = uint(usage.Input()), uint(usage.Output())
			row.Cost = EstimateCost(model, row.InputTokens, row.OutputTokens)
		}
	}
	if err != nil {
		row.Status = models.Failed
		message := truncate(err.Error(), 2000)
		row.Error = &message
	}
	if logErr := facades.Orm().Query().Create(&row); logErr != nil {
		facades.Log().Error("AI API log not written: " + logErr.Error())
	}
	result.LogID = row.ID
	return result, err
}
