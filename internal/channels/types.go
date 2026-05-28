package channels

import (
	"context"
	"time"
)

type Channel interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status() ChannelStatus
}

type ChannelStatus struct {
	Name      string    `json:"name"`
	Running   bool      `json:"running"`
	LastError string    `json:"last_error,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
}
