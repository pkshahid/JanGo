package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	godjangohttp "github.com/pkshahid/JanGo/http"
)

type mockWSView struct {
	connected    atomic.Bool
	disconnected atomic.Bool
	lastMessage  atomic.Value
}

func (m *mockWSView) Connect(conn *WebSocketConn, req *godjangohttp.Request) error {
	m.connected.Store(true)
	return nil
}

func (m *mockWSView) Receive(conn *WebSocketConn, messageType int, data []byte) error {
	m.lastMessage.Store(string(data))
	if string(data) == "echo" {
		conn.Send([]byte("echo_reply"))
	}
	return nil
}

func (m *mockWSView) Disconnect(conn *WebSocketConn, code int, reason string) error {
	m.disconnected.Store(true)
	return nil
}

func TestWebSocketUpgrader(t *testing.T) {
	view := &mockWSView{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := godjangohttp.NewRequest(r)
		ServeWebSocket(view, req, w)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Could not open websocket connection: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)
	if !view.connected.Load() {
		t.Error("Expected Connect to be called")
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte("echo"))
	if err != nil {
		t.Fatalf("Could not write to ws: %v", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Could not read from ws: %v", err)
	}

	if string(msg) != "echo_reply" {
		t.Errorf("Expected 'echo_reply', got %s", msg)
	}

	conn.Close()
	time.Sleep(100 * time.Millisecond)
	if !view.disconnected.Load() {
		t.Error("Expected Disconnect to be called")
	}
}

func TestInMemoryChannelLayer(t *testing.T) {
	cl := NewInMemoryChannelLayer()

	cl.GroupAdd("chat", "user1")
	cl.GroupAdd("chat", "user2")

	cl.GroupSend("chat", "hello")

	msg1, _ := cl.Receive("user1")
	if msg1.(string) != "hello" {
		t.Errorf("user1 expected hello, got %v", msg1)
	}

	msg2, _ := cl.Receive("user2")
	if msg2.(string) != "hello" {
		t.Errorf("user2 expected hello, got %v", msg2)
	}

	cl.GroupDiscard("chat", "user1")
	cl.GroupSend("chat", "world")

	// This would block if empty, so we use a select in real life or assume Send works correctly
	msg3, _ := cl.Receive("user2")
	if msg3.(string) != "world" {
		t.Errorf("user2 expected world, got %v", msg3)
	}
}
