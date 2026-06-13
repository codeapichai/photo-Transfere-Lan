package websocket

import (
	"encoding/json"
	"sync"

	fiberws "github.com/gofiber/contrib/websocket"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*fiberws.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[*fiberws.Conn]struct{}{}}
}

func (h *Hub) Handle(conn *fiberws.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *Hub) Publish(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		_ = conn.WriteMessage(fiberws.TextMessage, payload)
	}
}
