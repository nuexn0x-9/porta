package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// IsWebSocketRequest checks if the incoming request is asking for a WebSocket upgrade
func IsWebSocketRequest(r *http.Request) bool {
	connHeader := strings.ToLower(r.Header.Get("Connection"))
	upgradeHeader := strings.ToLower(r.Header.Get("Upgrade"))
	return strings.Contains(connHeader, "upgrade") && upgradeHeader == "websocket"
}

// ProxyWebSocket proxies an upgraded WebSocket connection directly to the target upstream address
func ProxyWebSocket(w http.ResponseWriter, r *http.Request, targetAddr string, rewritePath string) error {
	// 1. Connect to backend target
	d := net.Dialer{Timeout: 5 * time.Second}
	backendConn, err := d.DialContext(r.Context(), "tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to upstream WebSocket server at %s: %w", targetAddr, err)
	}
	defer backendConn.Close()

	// 2. Hijack client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return fmt.Errorf("underlying server does not support connection hijacking")
	}

	clientConn, brw, err := hijacker.Hijack()
	if err != nil {
		return fmt.Errorf("failed to hijack client connection: %w", err)
	}
	defer clientConn.Close()

	// 3. Reconstruct handshake request with modified path and forward to backend
	originalPath := r.URL.Path
	r.URL.Path = rewritePath
	r.URL.Scheme = "http"
	r.URL.Host = targetAddr

	err = r.Write(backendConn)
	r.URL.Path = originalPath // restore
	if err != nil {
		return fmt.Errorf("failed to write WebSocket handshake to backend: %w", err)
	}

	// 4. Bidirectional data piping
	errChan := make(chan error, 2)

	// Client -> Backend (include any buffered bytes from brw)
	go func() {
		if brw.Reader.Buffered() > 0 {
			bufferedData := make([]byte, brw.Reader.Buffered())
			_, _ = brw.Reader.Read(bufferedData)
			_, _ = backendConn.Write(bufferedData)
		}
		_, copyErr := io.Copy(backendConn, clientConn)
		errChan <- copyErr
	}()

	// Backend -> Client
	go func() {
		_, copyErr := io.Copy(clientConn, backendConn)
		errChan <- copyErr
	}()

	// Wait until one direction terminates
	<-errChan
	return nil
}
