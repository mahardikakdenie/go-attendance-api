package events

import (
	"context"
	"go-attendance-api/internal/model"
)

type NavInvalidator interface {
	InvalidateAllNavCaches(ctx context.Context)
}

type SSEBroadcaster interface {
	BroadcastSSEEvent(ctx context.Context, event model.SSEEvent) error
}

func RegisterHandlers(menuService NavInvalidator, notificationService SSEBroadcaster) {
	dispatcher := GetDispatcher()

	// 1. When Menu changes, invalidate all nav caches and notify all users
	dispatcher.Register(MenuChangedEvent, func(ctx context.Context, event Event) {
		bgCtx := context.Background()
		menuService.InvalidateAllNavCaches(bgCtx)

		_ = notificationService.BroadcastSSEEvent(bgCtx, model.SSEEvent{
			Type: "RELOAD_NAV",
			Data: map[string]string{"action": "refresh_sidebar"},
		})
	})

	// 2. When Role Permissions change, invalidate all nav caches and notify all users
	dispatcher.Register(RolePermissionsChanged, func(ctx context.Context, event Event) {
		bgCtx := context.Background()
		menuService.InvalidateAllNavCaches(bgCtx)

		_ = notificationService.BroadcastSSEEvent(bgCtx, model.SSEEvent{
			Type: "SYNC_PERMISSIONS",
			Data: map[string]string{"action": "refresh_sidebar"},
		})
	})
}
