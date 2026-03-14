package logic

import (
	"context"
	"errors"

	"pallink/activity/activity"
	"pallink/activity/internal/dao"
	"pallink/activity/internal/svc"
	"pallink/common/mq"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCommentLogic) CreateComment(in *activity.CreateCommentRequest) (*activity.CreateCommentResponse, error) {
	if in.ActivityId == 0 || in.UserId == 0 {
		return nil, errors.New("activity_id/user_id required")
	}
	if in.Content == "" {
		return nil, errors.New("content required")
	}

	activityCreatorID, err := dao.GetActivityCreator(l.ctx, l.svcCtx.DB, in.ActivityId)
	if err != nil {
		return nil, err
	}

	id, parentUserID, err := dao.CreateComment(l.ctx, l.svcCtx.DB, in.ActivityId, in.UserId, in.ParentId, in.Content)
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.MQ.PublishJSON(l.ctx, mq.AuditMessage{Type: "comment", ID: id}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.NotificationMQ.PublishJSON(l.ctx, mq.CommentNotificationEvent{
		CommentId:         id,
		ActivityId:        in.ActivityId,
		ParentId:          in.ParentId,
		ActorId:           in.UserId,
		ActivityCreatorId: activityCreatorID,
		ParentUserId:      parentUserID,
		Content:           in.Content,
	}); err != nil {
		return nil, err
	}

	return &activity.CreateCommentResponse{CommentId: id}, nil
}
