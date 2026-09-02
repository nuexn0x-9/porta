package proxy

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsWebSocketRequest(t *testing.T) {
	reqNormal := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	if IsWebSocketRequest(reqNormal) {
		t.Fatalf("expected normal request not to be detected as websocket")
	}

	reqWS := httptest.NewRequest(http.MethodGet, "/socket.io", nil)
	reqWS.Header.Set("Connection", "Upgrade")
	reqWS.Header.Set("Upgrade", "websocket")

	if !IsWebSocketRequest(reqWS) {
		t.Fatalf("expected websocket upgrade request to be detected")
	}
}

func TestWebSocketProxyEcho(t *testing.T) {
	// 1. Mock TCP echo backend
	backendListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind mock backend: %v", err)
	}
	defer backendListener.Close()

	go func() {
		conn, err := backendListener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		// Read initial simulated handshake
		for {
			line, err := reader.ReadString('\n')
			if err != nil || strings.TrimSpace(line) == "" {
				break
			}
		}

		// Send fake upgrade response
		_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))

		// Echo back data
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			_, _ = conn.Write(buf[:n])
		}
	}()

	backendAddr := backendListener.Addr().String()

	// 2. Mock client and proxy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = ProxyWebSocket(w, r, backendAddr, "/ws")
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// 3. Connect raw TCP to test server and send handshake
	clientConn, err := net.Dial("tcp", server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("failed to dial test server: %v", err)
	}
	defer clientConn.Close()

	handshake := "GET /ws HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"
	_, _ = clientConn.Write([]byte(handshake))

	clientReader := bufio.NewReader(clientConn)
	statusLine, err := clientReader.ReadString('\n')
	if err != nil {
		t.Fatalf("failed to read upgrade response: %v", err)
	}

	if !strings.Contains(statusLine, "101") {
		t.Fatalf("expected 101 Switching Protocols, got: %s", statusLine)
	}
}
