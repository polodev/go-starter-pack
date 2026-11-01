package aiservice

import (
	"encoding/json"
	"strings"

	"goravel/app/facades"
)

// Price is what a model costs in USD per million tokens.
type Price struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

// Prices come from AI_PRICES in .env, as JSON: {"claude-sonnet-5-5": {"input": 3, "output": 15}}. Providers change
// their prices, so none are built in: without a price the cost stays empty rather than wrong.
func Prices() map[string]Price {
	prices := map[string]Price{}
	text, _ := facades.Config().Env("AI_PRICES", "").(string)
	if strings.TrimSpace(text) != "" {
		_ = json.Unmarshal([]byte(text), &prices)
	}
	return prices
}

// EstimateCost is the cost of a call in USD, or nil when the model has no price.
func EstimateCost(model string, inputTokens, outputTokens uint) *float64 {
	price, ok := Prices()[model]
	if !ok {
		return nil
	}
	cost := (float64(inputTokens)*price.Input + float64(outputTokens)*price.Output) / 1_000_000
	return &cost
}
