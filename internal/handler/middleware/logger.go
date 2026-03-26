package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry represents a single request log.
type LogEntry struct {
	Time       time.Time `json:"time"`
	Status     int       `json:"status"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	ClientIP   string    `json:"client_ip"`
	Latency    string    `json:"latency"`
	Error      string    `json:"error,omitempty"`
	IsError    bool      `json:"is_error"`
}

// LogStore keeps the last N log entries in a ring buffer.
type LogStore struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

func NewLogStore(maxSize int) *LogStore {
	return &LogStore{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (s *LogStore) Add(e LogEntry) {
	s.mu.Lock()
	if len(s.entries) >= s.maxSize {
		// drop oldest
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
	s.mu.Unlock()
}

// All returns all entries, newest last.
func (s *LogStore) All() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]LogEntry, len(s.entries))
	copy(cp, s.entries)
	return cp
}

// Errors returns only error entries (status >= 400).
func (s *LogStore) Errors() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []LogEntry
	for _, e := range s.entries {
		if e.IsError {
			out = append(out, e)
		}
	}
	return out
}

// RequestLogger logs every request into the LogStore.
func RequestLogger(store *LogStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.String()
		}

		entry := LogEntry{
			Time:     start,
			Status:   status,
			Method:   c.Request.Method,
			Path:     c.Request.URL.Path,
			ClientIP: c.ClientIP(),
			Latency:  fmt.Sprintf("%dms", latency.Milliseconds()),
			Error:    errMsg,
			IsError:  status >= 400 || len(c.Errors) > 0,
		}

		store.Add(entry)
	}
}
