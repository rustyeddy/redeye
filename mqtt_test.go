package redeye

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMessage satisfies the mqtt.Message interface for unit testing.
type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool  { return false }
func (m *mockMessage) Qos() byte        { return 0 }
func (m *mockMessage) Retained() bool   { return false }
func (m *mockMessage) MessageID() uint16 { return 0 }
func (m *mockMessage) Topic() string    { return m.topic }
func (m *mockMessage) Payload() []byte  { return m.payload }
func (m *mockMessage) Ack()             {}

func jsonPayload(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// --- Construction ---

func TestNewMessengerSetsFields(t *testing.T) {
	m := NewMessenger("tcp://localhost:1883", "/redeye")
	assert.Equal(t, "tcp://localhost:1883", m.Broker)
	assert.Equal(t, "/redeye", m.Prefix)
	assert.NotEmpty(t, m.Name, "Name should default to hostname or fallback")
	assert.NotNil(t, m.quit)
}

func TestNewMessengerNameFallback(t *testing.T) {
	m := NewMessenger("", "")
	// Name is either the real hostname or "redeye"; either way non-empty.
	assert.NotEmpty(t, m.Name)
}

// --- Topic helpers ---

func TestMessengerAnnounceTopic(t *testing.T) {
	m := NewMessenger("", "/cam")
	assert.Equal(t, "/cam/announce/camera", m.AnnounceTopic())
}

func TestMessengerCommandTopic(t *testing.T) {
	m := NewMessenger("", "/cam")
	m.Name = "node1"
	assert.Equal(t, "/cam/camera/node1", m.CommandTopic())
}

func TestMessengerTopicsRespectPrefix(t *testing.T) {
	m := NewMessenger("", "/custom/prefix")
	m.Name = "cam42"
	assert.True(t, strings.HasPrefix(m.AnnounceTopic(), "/custom/prefix"))
	assert.True(t, strings.HasPrefix(m.CommandTopic(), "/custom/prefix"))
	assert.Contains(t, m.CommandTopic(), "cam42")
}

// --- Status (no broker) ---

func TestMessengerStatusDisconnected(t *testing.T) {
	m := NewMessenger("tcp://localhost:1883", "/redeye")
	s := m.Status()
	assert.Equal(t, "tcp://localhost:1883", s.Broker)
	assert.False(t, s.Connected, "should not be connected before Connect()")
	assert.Empty(t, s.Subscriptions)
}

// --- Command dispatch via handleMessage ---

func TestHandleSnapCommandCallsSnapper(t *testing.T) {
	ms := &mockSnapper{}
	prev := activeSnapper.Load()
	SetSnapper(ms)
	defer activeSnapper.Store(prev)

	m := NewMessenger("", "/redeye")
	m.handleMessage(nil, &mockMessage{
		topic:   "/redeye/camera/node1",
		payload: jsonPayload(t, Command{Command: "snap", File: "test.jpg"}),
	})

	assert.Equal(t, "test.jpg", ms.calledWith)
}

func TestHandleSnapCommandDefaultFilename(t *testing.T) {
	ms := &mockSnapper{}
	prev := activeSnapper.Load()
	SetSnapper(ms)
	defer activeSnapper.Store(prev)

	m := NewMessenger("", "/redeye")
	m.handleMessage(nil, &mockMessage{
		topic:   "/redeye/camera/node1",
		payload: jsonPayload(t, Command{Command: "snap"}),
	})

	assert.Contains(t, ms.calledWith, "snapshot-")
	assert.Contains(t, ms.calledWith, ".jpg")
}

func TestHandleSnapCommandPropagatesSnapError(t *testing.T) {
	ms := &mockSnapper{returnErr: fmt.Errorf("disk full")}
	prev := activeSnapper.Load()
	SetSnapper(ms)
	defer activeSnapper.Store(prev)

	m := NewMessenger("", "/redeye")
	// Should log the error and not panic.
	assert.NotPanics(t, func() {
		m.handleMessage(nil, &mockMessage{
			topic:   "/redeye/camera/node1",
			payload: jsonPayload(t, Command{Command: "snap", File: "out.jpg"}),
		})
	})
	assert.Equal(t, "out.jpg", ms.calledWith)
}

func TestHandleSnapCommandNoSnapper(t *testing.T) {
	prev := activeSnapper.Load()
	SetSnapper(nil)
	defer activeSnapper.Store(prev)

	m := NewMessenger("", "/redeye")
	assert.NotPanics(t, func() {
		m.handleMessage(nil, &mockMessage{
			topic:   "/redeye/camera/node1",
			payload: jsonPayload(t, Command{Command: "snap", File: "out.jpg"}),
		})
	})
}

func TestHandleUnknownCommandDoesNotPanic(t *testing.T) {
	m := NewMessenger("", "/redeye")
	assert.NotPanics(t, func() {
		m.handleMessage(nil, &mockMessage{
			topic:   "/redeye/camera/node1",
			payload: jsonPayload(t, Command{Command: "reboot"}),
		})
	})
}

func TestHandleMalformedPayloadDoesNotPanic(t *testing.T) {
	m := NewMessenger("", "/redeye")
	assert.NotPanics(t, func() {
		m.handleMessage(nil, &mockMessage{
			topic:   "/redeye/camera/node1",
			payload: []byte("not-json"),
		})
	})
}

// --- Close ---

func TestMessengerCloseIdempotent(t *testing.T) {
	m := NewMessenger("", "/redeye")
	// Close on a Messenger that was never connected must not panic.
	assert.NotPanics(t, func() {
		m.Close()
		m.Close() // second call must also be safe
	})
}
