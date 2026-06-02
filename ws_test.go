package redeye

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// dialWS dials the /ws endpoint on srv and returns the connection.
func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, wsURL, nil)
	require.NoError(t, err, "WebSocket dial")
	return c
}

// readEvent reads one event from the WebSocket with a 2-second timeout.
func readEvent(t *testing.T, c *websocket.Conn) Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var e Event
	require.NoError(t, wsjson.Read(ctx, c, &e))
	return e
}

func TestWSHandlerSendsConnectedEvent(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := dialWS(t, srv)
	defer c.Close(websocket.StatusNormalClosure, "")

	e := readEvent(t, c)
	assert.Equal(t, "connected", e.Type)
}

func TestWSHandlerPushesPublishedEvents(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := dialWS(t, srv)
	defer c.Close(websocket.StatusNormalClosure, "")

	// Consume the "connected" sentinel so the channel is clear.
	connected := readEvent(t, c)
	require.Equal(t, "connected", connected.Type)

	// Publish a test event — it must arrive at the client.
	Events.Publish(NewEvent("frame", map[string]int{"w": 1280, "h": 720}))

	got := readEvent(t, c)
	assert.Equal(t, "frame", got.Type)
}

func TestWSHandlerMultipleClients(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c1 := dialWS(t, srv)
	defer c1.Close(websocket.StatusNormalClosure, "")
	c2 := dialWS(t, srv)
	defer c2.Close(websocket.StatusNormalClosure, "")

	// Drain the "connected" sentinels on both.
	require.Equal(t, "connected", readEvent(t, c1).Type)
	require.Equal(t, "connected", readEvent(t, c2).Type)

	Events.Publish(NewEvent("ping", nil))

	assert.Equal(t, "ping", readEvent(t, c1).Type)
	assert.Equal(t, "ping", readEvent(t, c2).Type)
}

func TestWSHandlerUnsubscribesOnDisconnect(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)
	srv := httptest.NewServer(mux)

	c := dialWS(t, srv)
	readEvent(t, c) // drain "connected"

	before := Events.Len()

	// Close the client first; CloseRead sees the close frame and cancels ctx.
	c.Close(websocket.StatusNormalClosure, "")
	// srv.Close() blocks until all handler goroutines return, which includes
	// the deferred Events.Unsubscribe call — use it as the sync barrier.
	srv.Close()

	assert.Less(t, Events.Len(), before, "subscriber should be removed after disconnect")
}

func TestIndexHandlerServesHTML(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAPIRoutes(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
}
