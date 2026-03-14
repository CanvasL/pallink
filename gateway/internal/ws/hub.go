package ws

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	mu    sync.RWMutex
	users map[uint64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		users: make(map[uint64]map[*Client]struct{}),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.users[client.UserID]
	if conns == nil {
		conns = make(map[*Client]struct{})
		h.users[client.UserID] = conns
	}
	conns[client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.users[client.UserID]
	if conns == nil {
		return
	}
	if _, ok := conns[client]; !ok {
		return
	}
	delete(conns, client)
	close(client.send)
	if len(conns) == 0 {
		delete(h.users, client.UserID)
	}
}

func (h *Hub) Send(userID uint64, event ServerEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.SendRaw(userID, data)
}

func (h *Hub) SendRaw(userID uint64, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.users[userID] {
		select {
		case client.send <- data:
		default:
			go client.Close()
		}
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, conns := range h.users {
		for client := range conns {
			close(client.send)
			_ = client.conn.Close()
		}
	}
	h.users = make(map[uint64]map[*Client]struct{})
}
