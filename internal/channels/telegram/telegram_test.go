package telegram

import (
	"context"
	"strings"
	"testing"

	"github.com/pafthang/paw/internal/config"
)

func TestStartMissingToken(t *testing.T) {
	ch := New(config.Settings{})
	if err := ch.Start(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSnippetTruncates(t *testing.T) {
	long := strings.Repeat("a", 300)
	if got := snippet(long); len(got) != 200 {
		t.Fatalf("len=%d", len(got))
	}
}
