package dao

import (
	"context"
	"errors"
	"strings"
	"time"

	"pallink/im/im"
	"pallink/im/internal/dao/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func GetOrCreateConversation(ctx context.Context, db sqlc.DBTX, userID, peerUserID uint64) (*im.ConversationInfo, error) {
	user1, user2, err := normalizeUsers(userID, peerUserID)
	if err != nil {
		return nil, err
	}

	q := sqlc.New(db)
	row, err := q.UpsertConversation(ctx, sqlc.UpsertConversationParams{
		User1ID: int64(user1),
		User2ID: int64(user2),
	})
	if err != nil {
		return nil, err
	}
	if err := q.EnsureConversationMember(ctx, sqlc.EnsureConversationMemberParams{
		ConversationID: row.ID,
		UserID:         int64(user1),
	}); err != nil {
		return nil, err
	}
	if err := q.EnsureConversationMember(ctx, sqlc.EnsureConversationMemberParams{
		ConversationID: row.ID,
		UserID:         int64(user2),
	}); err != nil {
		return nil, err
	}

	summary, err := q.GetConversationSummary(ctx, sqlc.GetConversationSummaryParams{
		ConversationID: row.ID,
		UserID:         int64(userID),
	})
	if err != nil {
		return nil, err
	}
	return conversationInfoFromSummary(summary), nil
}

func QueryConversationList(ctx context.Context, db sqlc.DBTX, userID uint64, page, pageSize int32) ([]*im.ConversationInfo, int32, error) {
	if userID == 0 {
		return nil, 0, errors.New("user_id required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := sqlc.New(db)
	total64, err := q.CountConversations(ctx, int64(userID))
	if err != nil {
		return nil, 0, err
	}

	rows, err := q.ListConversations(ctx, sqlc.ListConversationsParams{
		UserID:     int64(userID),
		PageOffset: (page - 1) * pageSize,
		PageLimit:  pageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	list := make([]*im.ConversationInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, conversationInfoFromList(row))
	}
	return list, int32(total64), nil
}

func CreateMessage(ctx context.Context, db sqlc.DBTX, senderID, peerUserID uint64, content string) (*im.MessageInfo, error) {
	user1, user2, err := normalizeUsers(senderID, peerUserID)
	if err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("content required")
	}

	q := sqlc.New(db)
	conversation, err := q.UpsertConversation(ctx, sqlc.UpsertConversationParams{
		User1ID: int64(user1),
		User2ID: int64(user2),
	})
	if err != nil {
		return nil, err
	}
	if err := q.EnsureConversationMember(ctx, sqlc.EnsureConversationMemberParams{
		ConversationID: conversation.ID,
		UserID:         int64(user1),
	}); err != nil {
		return nil, err
	}
	if err := q.EnsureConversationMember(ctx, sqlc.EnsureConversationMemberParams{
		ConversationID: conversation.ID,
		UserID:         int64(user2),
	}); err != nil {
		return nil, err
	}

	row, err := q.InsertMessage(ctx, sqlc.InsertMessageParams{
		ConversationID: conversation.ID,
		SenderID:       int64(senderID),
		Content:        content,
	})
	if err != nil {
		return nil, err
	}
	if _, err := q.UpdateConversationReadCursor(ctx, sqlc.UpdateConversationReadCursorParams{
		ConversationID: conversation.ID,
		UserID:         int64(senderID),
		LastReadMsgID:  row.ID,
	}); err != nil {
		return nil, err
	}

	return &im.MessageInfo{
		Id:             uint64(row.ID),
		ConversationId: uint64(row.ConversationID),
		SenderId:       uint64(row.SenderID),
		Content:        row.Content,
		AuditStatus:    int32(row.AuditStatus),
		CreatedAt:      timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
	}, nil
}

func QueryMessages(ctx context.Context, db sqlc.DBTX, userID, conversationID, beforeMessageID uint64, pageSize int32) ([]*im.MessageInfo, bool, error) {
	if userID == 0 || conversationID == 0 {
		return nil, false, errors.New("user_id/conversation_id required")
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	q := sqlc.New(db)
	if _, err := q.GetConversationMember(ctx, sqlc.GetConversationMemberParams{
		ConversationID: int64(conversationID),
		UserID:         int64(userID),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, errors.New("conversation not found")
		}
		return nil, false, err
	}

	rows, err := q.ListMessages(ctx, sqlc.ListMessagesParams{
		ConversationID:  int64(conversationID),
		BeforeMessageID: int64(beforeMessageID),
		PageLimit:       pageSize + 1,
	})
	if err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > int(pageSize)
	if hasMore {
		rows = rows[:pageSize]
	}

	list := make([]*im.MessageInfo, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		row := rows[i]
		list = append(list, &im.MessageInfo{
			Id:             uint64(row.ID),
			ConversationId: uint64(row.ConversationID),
			SenderId:       uint64(row.SenderID),
			Content:        row.Content,
			AuditStatus:    int32(row.AuditStatus),
			CreatedAt:      timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
		})
	}

	return list, hasMore, nil
}

func MarkConversationRead(ctx context.Context, db sqlc.DBTX, userID, conversationID, lastReadMsgID uint64) (uint64, error) {
	if userID == 0 || conversationID == 0 {
		return 0, errors.New("user_id/conversation_id required")
	}

	q := sqlc.New(db)
	if _, err := q.GetConversationMember(ctx, sqlc.GetConversationMemberParams{
		ConversationID: int64(conversationID),
		UserID:         int64(userID),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("conversation not found")
		}
		return 0, err
	}

	resolved := lastReadMsgID
	if resolved == 0 {
		id, err := q.GetLatestMessageID(ctx, int64(conversationID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, nil
			}
			return 0, err
		}
		resolved = uint64(id)
	} else {
		id, err := q.GetMessageInConversation(ctx, sqlc.GetMessageInConversationParams{
			ConversationID: int64(conversationID),
			MessageID:      int64(lastReadMsgID),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, errors.New("message not found in conversation")
			}
			return 0, err
		}
		resolved = uint64(id)
	}

	if _, err := q.UpdateConversationReadCursor(ctx, sqlc.UpdateConversationReadCursorParams{
		ConversationID: int64(conversationID),
		UserID:         int64(userID),
		LastReadMsgID:  int64(resolved),
	}); err != nil {
		return 0, err
	}
	return resolved, nil
}

func conversationInfoFromSummary(row sqlc.GetConversationSummaryRow) *im.ConversationInfo {
	return &im.ConversationInfo{
		Id:            uint64(row.ID),
		PeerUserId:    uint64(row.PeerUserID),
		LastMessageId: uint64(row.LastMessageID),
		LastSenderId:  uint64(row.LastSenderID),
		LastMessage:   row.LastMessage,
		LastMessageAt: timestamppb.New(timeFromTimestamptz(row.LastMessageAt)),
		UnreadCount:   int32(row.UnreadCount),
		CreatedAt:     timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
	}
}

func conversationInfoFromList(row sqlc.ListConversationsRow) *im.ConversationInfo {
	return &im.ConversationInfo{
		Id:            uint64(row.ID),
		PeerUserId:    uint64(row.PeerUserID),
		LastMessageId: uint64(row.LastMessageID),
		LastSenderId:  uint64(row.LastSenderID),
		LastMessage:   row.LastMessage,
		LastMessageAt: timestamppb.New(timeFromTimestamptz(row.LastMessageAt)),
		UnreadCount:   int32(row.UnreadCount),
		CreatedAt:     timestamppb.New(timeFromTimestamptz(row.CreatedAt)),
	}
}

func normalizeUsers(userID, peerUserID uint64) (uint64, uint64, error) {
	if userID == 0 || peerUserID == 0 {
		return 0, 0, errors.New("user_id/peer_user_id required")
	}
	if userID == peerUserID {
		return 0, 0, errors.New("cannot chat with self")
	}
	if userID < peerUserID {
		return userID, peerUserID, nil
	}
	return peerUserID, userID, nil
}

func timeFromTimestamptz(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}
