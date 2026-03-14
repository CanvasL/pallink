package server

import (
	"context"

	"pallink/im/im"
	"pallink/im/internal/logic"
	"pallink/im/internal/svc"
)

type ImServer struct {
	svcCtx *svc.ServiceContext
	im.UnimplementedImServer
}

func NewImServer(svcCtx *svc.ServiceContext) *ImServer {
	return &ImServer{
		svcCtx: svcCtx,
	}
}

func (s *ImServer) OpenConversation(ctx context.Context, in *im.OpenConversationRequest) (*im.ConversationInfo, error) {
	l := logic.NewOpenConversationLogic(ctx, s.svcCtx)
	return l.OpenConversation(in)
}

func (s *ImServer) GetConversationList(ctx context.Context, in *im.GetConversationListRequest) (*im.GetConversationListResponse, error) {
	l := logic.NewGetConversationListLogic(ctx, s.svcCtx)
	return l.GetConversationList(in)
}

func (s *ImServer) SendMessage(ctx context.Context, in *im.SendMessageRequest) (*im.MessageInfo, error) {
	l := logic.NewSendMessageLogic(ctx, s.svcCtx)
	return l.SendMessage(in)
}

func (s *ImServer) GetMessages(ctx context.Context, in *im.GetMessagesRequest) (*im.GetMessagesResponse, error) {
	l := logic.NewGetMessagesLogic(ctx, s.svcCtx)
	return l.GetMessages(in)
}

func (s *ImServer) MarkConversationRead(ctx context.Context, in *im.MarkConversationReadRequest) (*im.MarkConversationReadResponse, error) {
	l := logic.NewMarkConversationReadLogic(ctx, s.svcCtx)
	return l.MarkConversationRead(in)
}
