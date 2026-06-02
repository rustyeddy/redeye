package redeye

import (
	"encoding/json"
	"sync"
)

// Event is the JSON envelope sent to every connected WebSocket client.
// Payload is raw JSON so each event type can carry its own structure
// without requiring a type switch on the receiving end.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// NewEvent marshals payload into an Event. If payload is nil, Payload is omitted.
func NewEvent(typ string, payload any) Event {
	if payload == nil {
		return Event{Type: typ}
	}
	b, _ := json.Marshal(payload)
	return Event{Type: typ, Payload: b}
}

// EventBus is a thread-safe fan-out channel: Publish delivers a copy of each
// Event to every active subscriber. Slow subscribers are silently skipped
// (non-blocking send) so a stalled client never blocks the camera pipeline.
type EventBus struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

// Events is the process-wide event bus. Filters publish here; WebSocket
// connections subscribe here.
var Events = &EventBus{subs: make(map[chan Event]struct{})}

// Subscribe registers ch to receive future events. The caller must call
// Unsubscribe when done.
func (b *EventBus) Subscribe(ch chan Event) {
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
}

// Unsubscribe removes ch from the bus. Events sent after this call will not
// reach ch.
func (b *EventBus) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.subs, ch)
	b.mu.Unlock()
}

// Publish delivers e to every subscriber. If a subscriber's channel buffer is
// full the event is dropped for that subscriber (the pipeline is never blocked).
func (b *EventBus) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- e:
		default:
		}
	}
}

// Len returns the current number of active subscribers.
func (b *EventBus) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subs)
}
