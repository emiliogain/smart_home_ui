package websocket

import (
	"context"
	"log"
	"net/http"

	ctxdomain "github.com/emiliogain/smart-home-backend/internal/domain/context"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// Hub wraps a Socket.IO server and implements secondary.EventBroadcaster.
type Hub struct {
	server *socketio.Server
}

// NewHub creates a Socket.IO server ready to be mounted on an HTTP router.
func NewHub() (*Hub, error) {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool {
					return true // allow all origins; CORS is handled by gin middleware
				},
			},
		},
	})

	server.OnConnect("/", func(c socketio.Conn) error {
		log.Printf("[ws] client connected: %s", c.ID())
		c.Join("dashboard")
		return nil
	})

	server.OnDisconnect("/", func(c socketio.Conn, reason string) {
		log.Printf("[ws] client disconnected: %s (%s)", c.ID(), reason)
	})

	server.OnEvent("/", "device_command", func(c socketio.Conn, data interface{}) {
		log.Printf("[ws] device_command from %s: %v", c.ID(), data)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Printf("[ws] serve error: %v", err)
		}
	}()

	return &Hub{server: server}, nil
}

// Handler returns the http.Handler for mounting on a router.
func (h *Hub) Handler() http.Handler {
	return h.server
}

// Close gracefully shuts down the Socket.IO server.
func (h *Hub) Close() error {
	return h.server.Close()
}

// BroadcastContextUpdate emits a context_update event to all connected clients.
func (h *Hub) BroadcastContextUpdate(_ context.Context, update ctxdomain.ContextUpdate) error {
	h.server.BroadcastToRoom("/", "dashboard", "context_update", update)
	return nil
}

// BroadcastDeviceStateUpdate emits a device_state_update event to all connected clients.
func (h *Hub) BroadcastDeviceStateUpdate(_ context.Context, deviceID string, state map[string]interface{}) error {
	payload := map[string]interface{}{
		"deviceId": deviceID,
		"state":    state,
	}
	h.server.BroadcastToRoom("/", "dashboard", "device_state_update", payload)
	return nil
}
