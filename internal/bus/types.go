package bus

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// MediaFile represents an inbound media file with its MIME type.
type MediaFile struct {
	Path     string `json:"path"`
	MimeType string `json:"mime_type,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// InboundMessage represents a message received from a connector/channel.
//
// WorkspaceID is the primary isolation boundary in PAW. It replaces the
// GoClaw-style TenantID terminology and should be set on every message that
// can touch workspace-scoped state: sessions, memory, files, tools, or agents.
type InboundMessage struct {
	Channel     string            `json:"channel"`
	SenderID    string            `json:"sender_id"`
	ChatID      string            `json:"chat_id"`
	Content     string            `json:"content"`
	Media       []MediaFile       `json:"media,omitempty"`
	SessionKey  string            `json:"session_key,omitempty"`
	PeerKind    string            `json:"peer_kind,omitempty"`
	WorkspaceID uuid.UUID         `json:"workspace_id,omitempty"`
	AgentID     uuid.UUID         `json:"agent_id,omitempty"`
	UserID      uuid.UUID         `json:"user_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// OutboundMessage represents a message to be sent to a connector/channel.
type OutboundMessage struct {
	Channel     string            `json:"channel"`
	ChatID      string            `json:"chat_id"`
	Content     string            `json:"content"`
	Media       []MediaAttachment `json:"media,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	WorkspaceID uuid.UUID         `json:"workspace_id,omitempty"`
	AgentID     uuid.UUID         `json:"agent_id,omitempty"`
	UserID      uuid.UUID         `json:"user_id,omitempty"`
}

// MediaAttachment represents a media file to be sent with a message.
type MediaAttachment struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Caption     string `json:"caption,omitempty"`
}

// Event represents a server-side event to broadcast to subscribers.
//
// WorkspaceID is intentionally not serialized by default. It is used by server
// filters to prevent leaking workspace-scoped events to other clients.
type Event struct {
	Name        string    `json:"name"`
	Payload     any       `json:"payload,omitempty"`
	WorkspaceID uuid.UUID `json:"-"`
}

// Cache invalidation kind constants.
const (
	CacheKindUser      = "user"
	CacheKindWorkspace = "workspace"
	CacheKindProject   = "project"
	CacheKindAgent     = "agent"
	CacheKindSession   = "session"
	CacheKindMemory    = "memory"
	CacheKindFile      = "file"
	CacheKindAPIKeys   = "api_keys"
	CacheKindAuth      = "auth"
)

// Topic constants for MessageBus.Subscribe() / Broadcast().
const (
	TopicCacheInvalidate = "cache:invalidate"
	TopicAudit           = "audit"
	TopicAuth            = "auth"
	TopicUser            = "user"
	TopicWorkspace       = "workspace"
	TopicProject         = "project"
	TopicAgent           = "agent"
	TopicSession         = "session"
	TopicChat            = "chat"
	TopicFile            = "file"
	TopicMemory          = "memory"
	TopicConfigChanged   = "config:changed"
)

const (
	EventUserCreated      = "user.created"
	EventUserUpdated      = "user.updated"
	EventUserDisabled     = "user.disabled"
	EventUserEnabled      = "user.enabled"
	EventAuthLogin        = "auth.login"
	EventAuthLogout       = "auth.logout"
	EventWorkspaceCreated = "workspace.created"
	EventWorkspaceUpdated = "workspace.updated"
	EventWorkspaceDeleted = "workspace.deleted"
	EventMemberAdded      = "workspace.member.added"
	EventMemberRemoved    = "workspace.member.removed"
	EventMemberUpdated    = "workspace.member.updated"
	EventProjectCreated   = "project.created"
	EventProjectUpdated   = "project.updated"
	EventProjectDeleted   = "project.deleted"
	EventAgentRunStarted  = "agent.run.started"
	EventAgentRunFinished = "agent.run.finished"
	EventAgentRunFailed   = "agent.run.failed"
	EventChatMessage      = "chat.message"
	EventFileChanged      = "file.changed"
	EventMemoryChanged    = "memory.changed"
	EventCacheInvalidate  = "cache.invalidate"
	EventAuditLog         = "audit.log"
)

// AuditEventPayload carries audit log data emitted by handlers.
type AuditEventPayload struct {
	ActorType   string          `json:"actor_type"`
	ActorID     string          `json:"actor_id"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	IPAddress   string          `json:"ip_address,omitempty"`
	Details     json.RawMessage `json:"details,omitempty"`
	WorkspaceID uuid.UUID       `json:"workspace_id,omitempty"`
}

// CacheInvalidatePayload signals cache layers to evict stale entries.
type CacheInvalidatePayload struct {
	Kind        string    `json:"kind"`
	Key         string    `json:"key,omitempty"`
	WorkspaceID uuid.UUID `json:"workspace_id,omitempty"`
}

// MessageHandler handles an inbound message from a specific connector/channel.
type MessageHandler func(InboundMessage) error

// EventHandler handles a broadcast event.
type EventHandler func(Event)

// EventPublisher abstracts event broadcast + subscription.
type EventPublisher interface {
	Subscribe(id string, handler EventHandler)
	Unsubscribe(id string)
	Broadcast(event Event)
}

// MessageRouter abstracts inbound/outbound message routing between connectors
// and the agent/runtime layer.
type MessageRouter interface {
	PublishInbound(msg InboundMessage)
	ConsumeInbound(ctx context.Context) (InboundMessage, bool)
	PublishOutbound(msg OutboundMessage)
	SubscribeOutbound(ctx context.Context) (OutboundMessage, bool)
}

// IsInternalSender returns true if senderID belongs to an internal system component.
func IsInternalSender(senderID string) bool {
	return strings.HasPrefix(senderID, "system:") ||
		strings.HasPrefix(senderID, "notification:") ||
		strings.HasPrefix(senderID, "agent:") ||
		strings.HasPrefix(senderID, "tool:") ||
		senderID == "session_send_tool"
}
