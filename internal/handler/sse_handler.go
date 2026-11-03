package handler

import (
	"io"
	"member-shop-api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	hub *services.SSEHub
}

func NewSSEHandler(hub *services.SSEHub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

func (h *SSEHandler) StreamEvents(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := h.hub.Subscribe()
	defer h.hub.Unsubscribe(clientChan)

	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	// Gửi heartbeat ban đầu
	c.Writer.Write([]byte("event: connected\ndata: {\"status\":\"connected\"}\n\n"))
	c.Writer.Flush()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}
			_, err := w.Write([]byte(msg))
			return err == nil
		case <-c.Request.Context().Done():
			return false
		}
	})
}
