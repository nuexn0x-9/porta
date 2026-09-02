package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	tokenQueryRegex = regexp.MustCompile(`([?&]porta_token=)([^&]+)`)
)

// LogEntry represents structured JSON access log entry
type LogEntry struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Upstream   string `json:"upstream"`
	ClientIP   string `json:"client_ip"`
	Error      string `json:"error,omitempty"`
}

// Logger handles thread-safe structured access logging and terminal output with credential redaction
type Logger struct {
	mu         sync.Mutex
	logFile    *os.File
	outWriter  io.Writer
	liveFeed   chan string
	subscribers []chan string
}

var globalLogger *Logger

// InitLogger initializes the global logger with disk persistence and live channel
func InitLogger(logDir string, stdout io.Writer) (*Logger, error) {
	if stdout == nil {
		stdout = os.Stdout
	}

	var file *os.File
	if logDir != "" {
		if err := os.MkdirAll(logDir, 0700); err == nil {
			logPath := filepath.Join(logDir, "access.log")
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err == nil {
				file = f
			}
		}
	}

	l := &Logger{
		logFile:     file,
		outWriter:   stdout,
		liveFeed:    make(chan string, 100),
		subscribers: make([]chan string, 0),
	}
	globalLogger = l
	return l, nil
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		l, _ := InitLogger("", os.Stdout)
		return l
	}
	return globalLogger
}

// RedactURL masks sensitive query parameters such as ?porta_token=
func RedactURL(rawURL string) string {
	return tokenQueryRegex.ReplaceAllString(rawURL, "${1}[REDACTED]")
}

// RedactPath sanitizes request path and query string
func RedactPath(reqURL *url.URL) string {
	if reqURL == nil {
		return "/"
	}
	path := reqURL.Path
	if reqURL.RawQuery != "" {
		sanitizedQuery := tokenQueryRegex.ReplaceAllString("?"+reqURL.RawQuery, "${1}[REDACTED]")
		path += sanitizedQuery
	}
	return path
}

// LogRequest records an HTTP request/response cycle with token masking
func (l *Logger) LogRequest(r *http.Request, status int, duration time.Duration, upstream string, errStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	sanitizedPath := RedactPath(r.URL)
	durationMs := duration.Milliseconds()

	entry := LogEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Level:      "info",
		Method:     r.Method,
		Path:       sanitizedPath,
		Status:     status,
		DurationMs: durationMs,
		Upstream:   upstream,
		ClientIP:   clientIP,
		Error:      errStr,
	}

	if status >= 400 {
		entry.Level = "warn"
	}
	if status >= 500 {
		entry.Level = "error"
	}

	// 1. Write to JSON log file
	if l.logFile != nil {
		data, err := json.Marshal(entry)
		if err == nil {
			_, _ = l.logFile.Write(append(data, '\n'))
		}
	}

	// 2. Format live human-readable line for TUI
	timeStr := time.Now().Format("15:04:05")
	statusColor := getColorForStatus(status)
	resetColor := "\033[0m"

	logLine := fmt.Sprintf("%s %s[%d]%s %-6s %s -> %s (%dms)",
		timeStr, statusColor, status, resetColor, r.Method, sanitizedPath, upstream, durationMs)

	// Broadcast to subscribers
	for _, ch := range l.subscribers {
		select {
		case ch <- logLine:
		default:
		}
	}
}

// Subscribe returns a channel receiving real-time log lines
func (l *Logger) Subscribe() chan string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch := make(chan string, 50)
	l.subscribers = append(l.subscribers, ch)
	return ch
}

// Close flushes and closes log files
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logFile != nil {
		_ = l.logFile.Close()
		l.logFile = nil
	}
}

func getColorForStatus(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\033[32m" // Green
	case status >= 300 && status < 400:
		return "\033[36m" // Cyan
	case status >= 400 && status < 500:
		return "\033[33m" // Yellow
	default:
		return "\033[31m" // Red
	}
}
