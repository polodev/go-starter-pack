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

// PNG is the picture the fake provider "draws": a valid 1x1 PNG.
var PNG = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0, 0, 0, 0x0d, 0x49, 0x48, 0x44, 0x52, 0, 0, 0, 1, 0, 0, 0, 1,
	8, 6, 0, 0, 0, 0x1f, 0x15, 0xc4, 0x89, 0, 0, 0, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0, 1, 0, 0, 5, 0, 1,
	0x0d, 0x0a, 0x2d, 0xb4, 0, 0, 0, 0, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82}

// Image makes the fake provider an image provider too. A prompt containing "FAIL" fails.
func (Provider) Image(ctx context.Context, prompt contractsai.ImagePrompt) (contractsai.ImageResponse, error) {
	if strings.Contains(prompt.Prompt, "FAIL") {
		return nil, errors.New("provider is down")
	}
	return &imageResponse{}, nil
}

type imageResponse struct{}

func (r *imageResponse) Content() ([]byte, error)             { return PNG, nil }
func (r *imageResponse) MimeType() string                     { return "image/png" }
func (r *imageResponse) Store(disk ...string) (string, error) { return "", errors.New("not used") }
func (r *imageResponse) StoreAs(path string, disk ...string) (string, error) {
	return "", errors.New("not used")
}
func (r *imageResponse) Usage() contractsai.Usage { return usage{20, 200} }
func (r *imageResponse) Then(callback func(contractsai.ImageResponse)) contractsai.ImageResponse {
	callback(r)
	return r
}
