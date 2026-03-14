package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"pallink/common/auth"
	"pallink/gateway/internal/svc"
	"pallink/gateway/internal/types"
	gatewayws "pallink/gateway/internal/ws"
	"pallink/im/imclient"

	"github.com/gorilla/websocket"
)

var imUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ImWsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("Authorization"))
		}

		userID, err := auth.ParseUserIDFromToken(svcCtx.Config.Auth.AccessSecret, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := imUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := gatewayws.NewClient(userID, conn, svcCtx.WsHub)
		client.Serve(func(client *gatewayws.Client, cmd gatewayws.ClientCommand) {
			handleImWsCommand(r.Context(), client, cmd, svcCtx)
		})
	}
}

func handleImWsCommand(ctx context.Context, client *gatewayws.Client, cmd gatewayws.ClientCommand, svcCtx *svc.ServiceContext) {
	switch cmd.Action {
	case "ping":
		_ = client.SendEvent(gatewayws.ServerEvent{
			Type:      "pong",
			Action:    cmd.Action,
			RequestID: cmd.RequestID,
		})
	case "open_conversation":
		var req types.OpenConversationReq
		if err := json.Unmarshal(cmd.Data, &req); err != nil {
			sendWsError(client, cmd, "invalid open_conversation payload")
			return
		}
		resp, err := svcCtx.ImRpc.OpenConversation(
			ctx,
			&imclient.OpenConversationRequest{
				UserId:     client.UserID,
				PeerUserId: req.PeerUserId,
			},
		)
		if err != nil {
			sendWsError(client, cmd, err.Error())
			return
		}
		_ = client.SendEvent(gatewayws.ServerEvent{
			Type:      "ack",
			Action:    cmd.Action,
			RequestID: cmd.RequestID,
			Data: types.OpenConversationResp{
				Conversation: wsConversationInfo(resp),
			},
		})
	case "send_message":
		var req types.SendMessageReq
		if err := json.Unmarshal(cmd.Data, &req); err != nil {
			sendWsError(client, cmd, "invalid send_message payload")
			return
		}
		resp, err := svcCtx.ImRpc.SendMessage(
			ctx,
			&imclient.SendMessageRequest{
				SenderId:   client.UserID,
				PeerUserId: req.PeerUserId,
				Content:    req.Content,
			},
		)
		if err != nil {
			sendWsError(client, cmd, err.Error())
			return
		}
		_ = client.SendEvent(gatewayws.ServerEvent{
			Type:      "ack",
			Action:    cmd.Action,
			RequestID: cmd.RequestID,
			Data: map[string]uint64{
				"message_id":      resp.Id,
				"conversation_id": resp.ConversationId,
			},
		})
	case "mark_read":
		var req types.MarkConversationReadReq
		if err := json.Unmarshal(cmd.Data, &req); err != nil {
			sendWsError(client, cmd, "invalid mark_read payload")
			return
		}
		resp, err := svcCtx.ImRpc.MarkConversationRead(
			ctx,
			&imclient.MarkConversationReadRequest{
				UserId:         client.UserID,
				ConversationId: req.ConversationId,
				LastReadMsgId:  req.LastReadMsgId,
			},
		)
		if err != nil {
			sendWsError(client, cmd, err.Error())
			return
		}
		_ = client.SendEvent(gatewayws.ServerEvent{
			Type:      "ack",
			Action:    cmd.Action,
			RequestID: cmd.RequestID,
			Data: types.MarkConversationReadResp{
				Success:       resp.Success,
				Message:       resp.Message,
				LastReadMsgId: resp.LastReadMsgId,
			},
		})
	default:
		sendWsError(client, cmd, "unsupported action")
	}
}

func sendWsError(client *gatewayws.Client, cmd gatewayws.ClientCommand, message string) {
	_ = client.SendEvent(gatewayws.ServerEvent{
		Type:      "error",
		Action:    cmd.Action,
		RequestID: cmd.RequestID,
		Error:     message,
	})
}

func wsConversationInfo(in *imclient.ConversationInfo) types.ConversationInfo {
	if in == nil {
		return types.ConversationInfo{}
	}
	lastMessageAt := int64(0)
	if in.LastMessageAt != nil {
		lastMessageAt = in.LastMessageAt.AsTime().Unix()
	}
	createdAt := int64(0)
	if in.CreatedAt != nil {
		createdAt = in.CreatedAt.AsTime().Unix()
	}

	return types.ConversationInfo{
		Id:            in.Id,
		PeerUserId:    in.PeerUserId,
		PeerNickname:  in.PeerNickname,
		PeerAvatar:    in.PeerAvatar,
		LastMessageId: in.LastMessageId,
		LastSenderId:  in.LastSenderId,
		LastMessage:   in.LastMessage,
		LastMessageAt: lastMessageAt,
		UnreadCount:   in.UnreadCount,
		CreatedAt:     createdAt,
	}
}
