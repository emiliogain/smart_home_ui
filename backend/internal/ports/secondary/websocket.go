package secondary

import (
	"context"

	ctxdomain "github.com/emiliogain/smart-home-backend/internal/domain/context"
)

// EventBroadcaster pushes real-time updates to connected frontend clients.
type EventBroadcaster interface {
	BroadcastContextUpdate(ctx context.Context, update ctxdomain.ContextUpdate) error
	BroadcastDeviceStateUpdate(ctx context.Context, deviceID string, state map[string]interface{}) error
}

// NoopBroadcaster is a no-op implementation used when WebSocket is not configured.
type NoopBroadcaster struct{}

func (NoopBroadcaster) BroadcastContextUpdate(_ context.Context, _ ctxdomain.ContextUpdate) error {
	return nil
}

func (NoopBroadcaster) BroadcastDeviceStateUpdate(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
