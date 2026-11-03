package services

import (
	"encoding/json"
	"fmt"
	"member-shop-api/pkg/logger"
	"sync"
)

type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type ClientChan chan string

type SSEHub struct {
	mu      sync.RWMutex
	clients map[ClientChan]bool
}

var GlobalSSEHub = NewSSEHub()

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[ClientChan]bool),
	}
}

func (h *SSEHub) Subscribe() ClientChan {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(ClientChan, 32)
	h.clients[ch] = true
	logger.Log.Debug().Int("active_sse_clients", len(h.clients)).Msg("SSE client connected")
	return ch
}

func (h *SSEHub) Unsubscribe(ch ClientChan) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
		logger.Log.Debug().Int("active_sse_clients", len(h.clients)).Msg("SSE client disconnected")
	}
}

func (h *SSEHub) Broadcast(eventName string, data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Failed to marshal SSE payload")
		return
	}

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(bytes))

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
			// client buffer full, skip or drop to avoid blocking
		}
	}
}
