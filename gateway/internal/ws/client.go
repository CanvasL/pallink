package ws

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 8 * 1024
	sendBufferSize = 64
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
)

type Client struct {
	UserID uint64
	conn   *websocket.Conn
	hub    *Hub
	send   chan []byte
	once   sync.Once
}

func NewClient(userID uint64, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID: userID,
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, sendBufferSize),
	}
}

func (c *Client) Serve(handler func(*Client, ClientCommand)) {
	c.hub.Register(c)
	go c.writePump()
	_ = c.SendEvent(ServerEvent{
		Type: "connected",
		Data: map[string]uint64{"user_id": c.UserID},
	})
	c.readPump(handler)
}

func (c *Client) SendEvent(event ServerEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return c.sendRaw(data)
}

func (c *Client) Close() {
	c.once.Do(func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	})
}

func (c *Client) readPump(handler func(*Client, ClientCommand)) {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var cmd ClientCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			_ = c.SendEvent(ServerEvent{Type: "error", Error: "invalid command json"})
			continue
		}
		if cmd.Action == "" {
			_ = c.SendEvent(ServerEvent{Type: "error", RequestID: cmd.RequestID, Error: "action required"})
			continue
		}
		handler(c, cmd)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := writer.Write(message); err != nil {
				_ = writer.Close()
				return
			}
			if err := writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendRaw(data []byte) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("connection closed")
		}
	}()

	select {
	case c.send <- data:
		return nil
	default:
		go c.Close()
		return errors.New("send buffer full")
	}
}
