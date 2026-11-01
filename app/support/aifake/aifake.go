// Package aifake is a fake AI provider for tests (Laravel's AI::fake()). config/ai.go registers it as provider
// "fake" only when the app runs under `go test`, so tests never call a paid API.
package aifake

import (
	"context"
	"errors"
	"strings"
	"sync"

	contractsai "github.com/goravel/framework/contracts/ai"
)

var (
	mutex  sync.Mutex
	reply  = "Hello from the fake model"
	tokens = [2]int{1200, 300}
)

// Reply sets what the fake model answers next.
func Reply(text string) {
	mutex.Lock()
	reply = text
	mutex.Unlock()
}

// Provider answers instantly. An input containing "FAIL" fails, like a provider that is down.
type Provider struct{}

func (Provider) Prompt(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.AgentResponse, error) {
	if strings.Contains(prompt.Input, "FAIL") {
		return nil, errors.New("provider is down")
	}
	mutex.Lock()
	defer mutex.Unlock()
	return &response{text: reply, usage: usage{tokens[0], tokens[1]}}, nil
}

func (Provider) Stream(ctx context.Context, prompt contractsai.AgentPrompt) (contractsai.StreamableAgentResponse, error) {
	return nil, errors.New("the fake provider does not stream")
}

type usage struct{ input, output int }

func (u usage) Input() int  { return u.input }
func (u usage) Output() int { return u.output }
func (u usage) Total() int  { return u.input + u.output }

type response struct {
	text  string
	usage usage
}

func (r *response) Text() string                      { return r.text }
func (r *response) Usage() contractsai.Usage          { return r.usage }
func (r *response) ToolCalls() []contractsai.ToolCall { return nil }
func (r *response) Then(callback func(contractsai.AgentResponse)) contractsai.AgentResponse {
	callback(r)
	return r
}
