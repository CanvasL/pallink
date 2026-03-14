package ws

import (
	"context"
	"encoding/json"
	"errors"

	"pallink/common/mq"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartRealtimeConsumer(ctx context.Context, sub *mq.FanoutSubscriber, hub *Hub) {
	if sub == nil || hub == nil {
		return
	}

	msgs, err := sub.Consume()
	if err != nil {
		logx.Error(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				logx.Error("realtime mq channel closed")
				return
			}

			if err := handleRealtimeMessage(msg.Body, hub); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func handleRealtimeMessage(body []byte, hub *Hub) error {
	var evt mq.ImRealtimeEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return err
	}
	if evt.Type != "im_message" || evt.MessageId == 0 || evt.ConversationId == 0 {
		return errors.New("invalid realtime event")
	}

	data, err := json.Marshal(ServerEvent{
		Type: "message",
		Data: evt,
	})
	if err != nil {
		return err
	}

	for _, userID := range evt.Targets {
		if userID == 0 {
			continue
		}
		hub.SendRaw(userID, data)
	}
	return nil
}
