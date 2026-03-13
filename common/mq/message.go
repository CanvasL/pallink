package mq

type AuditMessage struct {
	Type string `json:"type"`
	ID   uint64 `json:"id"`
}
