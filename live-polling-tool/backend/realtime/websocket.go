package realtime

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"live-polling-tool/backend/repositories"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 4096, CheckOrigin: func(r *http.Request) bool { return true }}

type Hub struct {
	Redis *Redis
	Repo  *repositories.Repo
}

func (h *Hub) Serve(c *gin.Context) {
	if h.Redis == nil {
		c.JSON(503, gin.H{"error": gin.H{"code": "REALTIME_UNAVAILABLE", "message": "Realtime service unavailable"}})
		return
	}
	conn, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	sub := h.Redis.Subscribe(ctx, c.Param("id"))
	defer sub.Close()
	if r, e := h.Redis.CachedResult(ctx, c.Param("id")); e == nil {
		_ = conn.WriteJSON(r)
	} else if r, e := h.Repo.Results(ctx, c.Param("id")); e == nil {
		_ = conn.WriteJSON(r)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, e := conn.ReadMessage()
			if e != nil {
				return
			}
		}
	}()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg := <-sub.Channel():
			if msg != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			}
		case <-ticker.C:
			_ = conn.WriteMessage(websocket.PingMessage, nil)
		case <-done:
			return
		case <-ctx.Done():
			return
		}
	}
}
