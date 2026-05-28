package bus

import "github.com/google/uuid"

// BroadcastForWorkspace broadcasts an event with explicit workspace scoping.
// Use it for events that carry workspace-specific data so WebSocket/API filters
// can prevent cross-workspace leaks.
func BroadcastForWorkspace(pub EventPublisher, name string, workspaceID uuid.UUID, payload any) {
	pub.Broadcast(Event{Name: name, WorkspaceID: workspaceID, Payload: payload})
}

// BroadcastGlobal broadcasts a process-wide event with no workspace scope.
// Use this only for safe global events such as health/config notifications.
func BroadcastGlobal(pub EventPublisher, name string, payload any) {
	pub.Broadcast(Event{Name: name, Payload: payload})
}
