package mq

type AuditMessage struct {
	Type string `json:"type"`
	ID   uint64 `json:"id"`
}

type CommentNotifyEvent struct {
	CommentId         uint64 `json:"comment_id"`
	ActivityId        uint64 `json:"activity_id"`
	ParentId          uint64 `json:"parent_id"`
	ActorId           uint64 `json:"actor_id"`
	ActivityCreatorId uint64 `json:"activity_creator_id"`
	ParentUserId      uint64 `json:"parent_user_id"`
	Content           string `json:"content"`
}

type ImMessageNotifyEvent struct {
	MessageId      uint64 `json:"message_id"`
	ConversationId uint64 `json:"conversation_id"`
	ActorId        uint64 `json:"actor_id"`
	ReceiverId     uint64 `json:"receiver_id"`
	Content        string `json:"content"`
}

type ImRealtimeEvent struct {
	Type           string   `json:"type"`
	Targets        []uint64 `json:"targets"`
	MessageId      uint64   `json:"message_id"`
	ConversationId uint64   `json:"conversation_id"`
	SenderId       uint64   `json:"sender_id"`
	ReceiverId     uint64   `json:"receiver_id"`
	Content        string   `json:"content"`
	AuditStatus    int32    `json:"audit_status"`
	CreatedAt      int64    `json:"created_at"`
}
