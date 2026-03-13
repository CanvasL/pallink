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
