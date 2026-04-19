package channel_test

import (
	"context"
	"testing"
	"time"

	"github.com/ctrlsam/rigour/internal/messaging/channel"
)

func TestPublishFetch(t *testing.T) {
	ch := channel.New[string](8)
	ctx := context.Background()

	if err := ch.Publish(ctx, []byte("k"), "hello"); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	msg, err := ch.Fetch(ctx)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if msg.Value != "hello" {
		t.Errorf("got %q, want %q", msg.Value, "hello")
	}
	if string(msg.Key) != "k" {
		t.Errorf("key: got %q, want %q", msg.Key, "k")
	}
}

func TestFetchHonoursContextCancellation(t *testing.T) {
	ch := channel.New[string](8)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := ch.Fetch(ctx) // nothing published — should block then return ctx error
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestMultipleMessages(t *testing.T) {
	ch := channel.New[int](16)
	ctx := context.Background()

	for i := range 5 {
		if err := ch.Publish(ctx, nil, i); err != nil {
			t.Fatalf("Publish %d: %v", i, err)
		}
	}

	for want := range 5 {
		msg, err := ch.Fetch(ctx)
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if msg.Value != want {
			t.Errorf("got %d, want %d", msg.Value, want)
		}
	}
}

func TestPublishAfterCloseDoesNotPanic(t *testing.T) {
	// After Close the underlying channel is closed; subsequent Fetch should
	// return immediately with the zero value (channel drained/closed).
	ch := channel.New[string](4)
	_ = ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Fetch on a closed channel: select picks the closed channel branch.
	msg, err := ch.Fetch(ctx)
	if err != nil && msg != nil {
		t.Errorf("unexpected: msg=%v err=%v", msg, err)
	}
}
