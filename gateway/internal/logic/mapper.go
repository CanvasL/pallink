package logic

import (
	"pallink/activity/activityclient"
	"pallink/gateway/internal/types"
	"pallink/im/imclient"
	"pallink/user/userclient"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toUserInfo(in *userclient.UserInfo) types.UserInfo {
	if in == nil {
		return types.UserInfo{}
	}
	return types.UserInfo{
		Id:       in.Id,
		Mobile:   in.Mobile,
		Nickname: in.Nickname,
		Avatar:   in.Avatar,
	}
}

func toActivityInfo(in *activityclient.ActivityInfo) types.ActivityInfo {
	if in == nil {
		return types.ActivityInfo{}
	}
	return types.ActivityInfo{
		Id:            in.Id,
		CreatorId:     in.CreatorId,
		Title:         in.Title,
		Description:   in.Description,
		Location:      in.Location,
		StartTime:     tsToUnix(in.StartTime),
		EndTime:       tsToUnix(in.EndTime),
		MaxPeople:     in.MaxPeople,
		CurrentPeople: in.CurrentPeople,
		Status:        in.Status,
		AuditStatus:   in.AuditStatus,
		CreatedAt:     tsToUnix(in.CreatedAt),
		IsEnrolled:    in.IsEnrolled,
		CreatorName:   in.CreatorName,
		CreatorAvatar: in.CreatorAvatar,
	}
}

func toActivityBrief(in *activityclient.ActivityInfo) types.ActivityBrief {
	if in == nil {
		return types.ActivityBrief{}
	}
	return types.ActivityBrief{
		Id:            in.Id,
		CreatorId:     in.CreatorId,
		CreatorName:   in.CreatorName,
		CreatorAvatar: in.CreatorAvatar,
		Title:         in.Title,
		Location:      in.Location,
		StartTime:     tsToUnix(in.StartTime),
		EndTime:       tsToUnix(in.EndTime),
		MaxPeople:     in.MaxPeople,
		CurrentPeople: in.CurrentPeople,
		Status:        in.Status,
		AuditStatus:   in.AuditStatus,
		IsEnrolled:    in.IsEnrolled,
	}
}

func toCommentInfo(in *activityclient.CommentInfo) types.CommentInfo {
	if in == nil {
		return types.CommentInfo{}
	}
	return types.CommentInfo{
		Id:          in.Id,
		ActivityId:  in.ActivityId,
		UserId:      in.UserId,
		ParentId:    in.ParentId,
		Content:     in.Content,
		CreatedAt:   tsToUnix(in.CreatedAt),
		AuditStatus: in.AuditStatus,
		Nickname:    in.Nickname,
		Avatar:      in.Avatar,
	}
}

func toConversationInfo(in *imclient.ConversationInfo) types.ConversationInfo {
	if in == nil {
		return types.ConversationInfo{}
	}
	return types.ConversationInfo{
		Id:            in.Id,
		PeerUserId:    in.PeerUserId,
		PeerNickname:  in.PeerNickname,
		PeerAvatar:    in.PeerAvatar,
		LastMessageId: in.LastMessageId,
		LastSenderId:  in.LastSenderId,
		LastMessage:   in.LastMessage,
		LastMessageAt: tsToUnix(in.LastMessageAt),
		UnreadCount:   in.UnreadCount,
		CreatedAt:     tsToUnix(in.CreatedAt),
	}
}

func toMessageInfo(in *imclient.MessageInfo) types.MessageInfo {
	if in == nil {
		return types.MessageInfo{}
	}
	return types.MessageInfo{
		Id:             in.Id,
		ConversationId: in.ConversationId,
		SenderId:       in.SenderId,
		Content:        in.Content,
		AuditStatus:    in.AuditStatus,
		CreatedAt:      tsToUnix(in.CreatedAt),
	}
}

func tsToUnix(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.AsTime().Unix()
}
