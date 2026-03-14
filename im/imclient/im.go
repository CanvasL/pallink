package imclient

import (
	"context"

	"pallink/im/im"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	ConversationInfo             = im.ConversationInfo
	GetConversationListRequest   = im.GetConversationListRequest
	GetConversationListResponse  = im.GetConversationListResponse
	GetMessagesRequest           = im.GetMessagesRequest
	GetMessagesResponse          = im.GetMessagesResponse
	MarkConversationReadRequest  = im.MarkConversationReadRequest
	MarkConversationReadResponse = im.MarkConversationReadResponse
	MessageInfo                  = im.MessageInfo
	OpenConversationRequest      = im.OpenConversationRequest
	SendMessageRequest           = im.SendMessageRequest

	Im interface {
		OpenConversation(ctx context.Context, in *OpenConversationRequest, opts ...grpc.CallOption) (*ConversationInfo, error)
		GetConversationList(ctx context.Context, in *GetConversationListRequest, opts ...grpc.CallOption) (*GetConversationListResponse, error)
		SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*MessageInfo, error)
		GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error)
		MarkConversationRead(ctx context.Context, in *MarkConversationReadRequest, opts ...grpc.CallOption) (*MarkConversationReadResponse, error)
	}

	defaultIm struct {
		cli zrpc.Client
	}
)

func NewIm(cli zrpc.Client) Im {
	return &defaultIm{cli: cli}
}

func (m *defaultIm) OpenConversation(ctx context.Context, in *OpenConversationRequest, opts ...grpc.CallOption) (*ConversationInfo, error) {
	client := im.NewImClient(m.cli.Conn())
	return client.OpenConversation(ctx, in, opts...)
}

func (m *defaultIm) GetConversationList(ctx context.Context, in *GetConversationListRequest, opts ...grpc.CallOption) (*GetConversationListResponse, error) {
	client := im.NewImClient(m.cli.Conn())
	return client.GetConversationList(ctx, in, opts...)
}

func (m *defaultIm) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*MessageInfo, error) {
	client := im.NewImClient(m.cli.Conn())
	return client.SendMessage(ctx, in, opts...)
}

func (m *defaultIm) GetMessages(ctx context.Context, in *GetMessagesRequest, opts ...grpc.CallOption) (*GetMessagesResponse, error) {
	client := im.NewImClient(m.cli.Conn())
	return client.GetMessages(ctx, in, opts...)
}

func (m *defaultIm) MarkConversationRead(ctx context.Context, in *MarkConversationReadRequest, opts ...grpc.CallOption) (*MarkConversationReadResponse, error) {
	client := im.NewImClient(m.cli.Conn())
	return client.MarkConversationRead(ctx, in, opts...)
}
