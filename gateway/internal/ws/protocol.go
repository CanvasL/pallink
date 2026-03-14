package ws

import "encoding/json"

type ClientCommand struct {
	Action    string          `json:"action"`
	RequestID string          `json:"request_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type ServerEvent struct {
	Type      string `json:"type"`
	Action    string `json:"action,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
}
