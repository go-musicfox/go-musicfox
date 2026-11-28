package externalapi

import (
	"sync"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *wsClient) writeJSON(value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(value)
}

type hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
}

func newHub() *hub {
	return &hub{clients: make(map[*wsClient]struct{})}
}

func (h *hub) add(conn *websocket.Conn) *wsClient {
	client := &wsClient{conn: conn}
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	return client
}

func (h *hub) remove(client *wsClient) {
	h.mu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		_ = client.conn.Close()
	}
	h.mu.Unlock()
}

func (h *hub) count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *hub) broadcast(eventType string, data any) {
	h.mu.RLock()
	clients := make([]*wsClient, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	payload := map[string]any{"kind": "event", "type": eventType, "data": data}
	for _, client := range clients {
		if err := client.writeJSON(payload); err != nil {
			h.remove(client)
		}
	}
}

func (h *hub) close() {
	h.mu.Lock()
	for client := range h.clients {
		_ = client.conn.Close()
		delete(h.clients, client)
	}
	h.mu.Unlock()
}
