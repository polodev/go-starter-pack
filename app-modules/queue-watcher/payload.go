package queuewatcher

import (
	"encoding/json"
)

// HiddenArgumentJobs are jobs whose arguments the screens never show. Goravel's mail job carries the whole mail,
// and a mail can hold a password reset or sign-in link that would let a reader act as the recipient.
var HiddenArgumentJobs = map[string]bool{"goravel_send_mail_job": true}

// Payload is what the screens show of Goravel's job payload ({"uuid", "signature", "args", "delay", "chain"}).
type Payload struct {
	UUID          string
	Signature     string
	Arguments     string // pretty JSON, or a note when hidden
	ChainLength   int
	ArgumentsHide bool
}

// ReadPayload reads a jobs / failed_jobs payload. A payload it cannot read still shows its signature as "unknown".
func ReadPayload(raw string) Payload {
	var task struct {
		UUID      string            `json:"uuid"`
		Signature string            `json:"signature"`
		Args      []json.RawMessage `json:"args"`
		Chain     []json.RawMessage `json:"chain"`
	}
	if err := json.Unmarshal([]byte(raw), &task); err != nil || task.Signature == "" {
		return Payload{Signature: "unknown", Arguments: "(the payload is not a Goravel job)", ArgumentsHide: true}
	}
	payload := Payload{UUID: task.UUID, Signature: task.Signature, ChainLength: len(task.Chain)}
	if HiddenArgumentJobs[task.Signature] {
		payload.ArgumentsHide = true
		payload.Arguments = "(hidden: this job's arguments can hold mail content and one-time links)"
		return payload
	}
	pretty, err := json.MarshalIndent(task.Args, "", "  ")
	if err != nil {
		pretty = []byte("[]")
	}
	payload.Arguments = string(pretty)
	return payload
}
