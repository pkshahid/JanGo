package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

// WebSocketView is the interface for handling WebSocket connections.
type WebSocketView interface {
	Connect(conn *WebSocketConn, req *godjangohttp.Request) error
	Receive(conn *WebSocketConn, messageType int, data []byte) error
	Disconnect(conn *WebSocketConn, code int, reason string) error
}

// WebSocketConn wraps gorilla websocket to provide a concurrent-safe API.
type WebSocketConn struct {
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
	closed bool
}

// NewWebSocketConn initializes a new WebSocketConn.
func NewWebSocketConn(c *websocket.Conn) *WebSocketConn {
	return &WebSocketConn{
		conn: c,
		send: make(chan []byte, 256),
	}
}

// Send queues data to be sent.
func (c *WebSocketConn) Send(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("websocket connection closed")
	}
	c.send <- data
	return nil
}

// SendJSON marshals v to JSON and sends it.
func (c *WebSocketConn) SendJSON(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(b)
}

// Close gracefully closes the connection.
func (c *WebSocketConn) Close(code int, reason string) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	msg := websocket.FormatCloseMessage(code, reason)
	c.conn.WriteMessage(websocket.CloseMessage, msg)

	// Close send channel to unblock writePump
	close(c.send)
	return nil
}

// writePump pumps messages from the send channel to the websocket connection.
func (c *WebSocketConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the view.
func (c *WebSocketConn) readPump(view WebSocketView, req *godjangohttp.Request) {
	defer func() {
		view.Disconnect(c, websocket.CloseNormalClosure, "Connection closed")
		c.Close(websocket.CloseNormalClosure, "Connection closed")
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}

		err = view.Receive(c, messageType, message)
		if err != nil {
			break
		}
	}
}

// Upgrader configures how HTTP connections are upgraded to WebSocket.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Default allows all, but should be checked against ALLOWED_HOSTS/CORS in a real app
		return true
	},
}

// ServeWebSocket upgrades the HTTP connection and spawns the read/write pumps.
func ServeWebSocket(view WebSocketView, req *godjangohttp.Request, w http.ResponseWriter) {
	conn, err := upgrader.Upgrade(w, req.Request, nil)
	if err != nil {
		fmt.Printf("upgrade error: %v", err)
		return
	}

	wsConn := NewWebSocketConn(conn)

	// Call Connect hook
	if err := view.Connect(wsConn, req); err != nil {
		wsConn.Close(websocket.CloseInternalServerErr, err.Error())
		return
	}

	// Allow collection of memory by doing all work in new goroutines.
	go wsConn.writePump()
	go wsConn.readPump(view, req)
}
