// Package models holds the AI API log model.
package models

import "github.com/goravel/framework/database/orm"

// Statuses of a call.
const (
	Success = "success"
	Failed  = "failed"
)

// AiApiLog is one call to an AI provider.
type AiApiLog struct {
	orm.Model
	Provider     string   `json:"provider"`
	ModelName    string   `gorm:"column:model" json:"model"`
	Purpose      string   `json:"purpose"`
	Status       string   `json:"status"`
	DurationMs   uint     `json:"duration_ms"`
	InputTokens  uint     `json:"input_tokens"`
	OutputTokens uint     `json:"output_tokens"`
	Cost         *float64 `json:"cost"`
	Sensitive    bool     `json:"sensitive"`
	Prompt       *string  `json:"prompt"`
	Response     *string  `json:"response"`
	Error        *string  `json:"error"`
}

func (AiApiLog) TableName() string { return "ai_api_logs" }

// Text is a nullable column for templates.
func Text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
