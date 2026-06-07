package redeye

import (
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// wsHandler upgrades the connection to WebSocket, subscribes to the global
// EventBus, and streams events as JSON until the client disconnects.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow all origins; fine for a LAN camera tool
	})
	if err != nil {
		slog.Error("ws accept error", "err", err)
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	// CloseRead starts a goroutine that drains and discards incoming messages.
	// It cancels the returned ctx when the client sends a close frame or drops
	// the connection, which lets the write loop below exit promptly.
	ctx := c.CloseRead(r.Context())

	ch := make(chan Event, 8)
	Events.Subscribe(ch)
	defer Events.Unsubscribe(ch)

	// Send a "connected" event immediately so the client (and tests) know the
	// subscription is active before any further events are published.
	if err := wsjson.Write(ctx, c, NewEvent("connected", nil)); err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			if err := wsjson.Write(ctx, c, e); err != nil {
				return
			}
		}
	}
}
