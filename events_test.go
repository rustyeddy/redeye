package redeye

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBus returns a fresh isolated EventBus for unit tests so they don't
// interact with the global Events bus or each other.
func newBus() *EventBus {
	return &EventBus{subs: make(map[chan Event]struct{})}
}

// --- NewEvent ---

func TestNewEventSetsType(t *testing.T) {
	e := NewEvent("status", nil)
	assert.Equal(t, "status", e.Type)
	assert.Nil(t, e.Payload)
}

func TestNewEventMarshalPayload(t *testing.T) {
	e := NewEvent("test", map[string]int{"count": 3})
	require.NotNil(t, e.Payload)

	var got map[string]int
	require.NoError(t, json.Unmarshal(e.Payload, &got))
	assert.Equal(t, 3, got["count"])
}

// --- EventBus ---

func TestEventBusPublishReachesSubscriber(t *testing.T) {
	b := newBus()
	ch := make(chan Event, 1)
	b.Subscribe(ch)

	b.Publish(NewEvent("ping", nil))

	select {
	case got := <-ch:
		assert.Equal(t, "ping", got.Type)
	case <-time.After(time.Second):
		t.Fatal("event not received within 1s")
	}
}

func TestEventBusUnsubscribeStopsDelivery(t *testing.T) {
	b := newBus()
	ch := make(chan Event, 4)
	b.Subscribe(ch)
	b.Unsubscribe(ch)

	b.Publish(NewEvent("should-not-arrive", nil))

	select {
	case <-ch:
		t.Fatal("received event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// correct: nothing arrived
	}
}

func TestEventBusFanOutToMultipleSubscribers(t *testing.T) {
	b := newBus()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	ch3 := make(chan Event, 1)
	b.Subscribe(ch1)
	b.Subscribe(ch2)
	b.Subscribe(ch3)

	b.Publish(NewEvent("broadcast", nil))

	for i, ch := range []chan Event{ch1, ch2, ch3} {
		select {
		case got := <-ch:
			assert.Equal(t, "broadcast", got.Type, "subscriber %d", i+1)
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d did not receive event", i+1)
		}
	}
}

func TestEventBusDropsEventForFullSubscriber(t *testing.T) {
	b := newBus()
	// Buffer of 1: first event fits, second is dropped.
	ch := make(chan Event, 1)
	b.Subscribe(ch)

	b.Publish(NewEvent("first", nil))
	b.Publish(NewEvent("dropped", nil)) // bus must not block

	got := <-ch
	assert.Equal(t, "first", got.Type)

	select {
	case extra := <-ch:
		t.Fatalf("expected channel to be empty, got %q", extra.Type)
	default:
		// correct: second event was dropped
	}
}

func TestEventBusLen(t *testing.T) {
	b := newBus()
	assert.Equal(t, 0, b.Len())

	ch := make(chan Event, 1)
	b.Subscribe(ch)
	assert.Equal(t, 1, b.Len())

	b.Unsubscribe(ch)
	assert.Equal(t, 0, b.Len())
}

func TestEventBusPublishToEmptyBusDoesNotPanic(t *testing.T) {
	b := newBus()
	assert.NotPanics(t, func() {
		b.Publish(NewEvent("noop", nil))
	})
}
