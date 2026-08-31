package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// auditUserKey carries the authenticated username through the request
// context from the auth middleware to the audit logger.
type auditUserKey struct{}

// withAuditUser attaches the authenticated username to a request context.
func withAuditUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, auditUserKey{}, user)
}

var (
	auditMu   sync.Mutex
	auditFile *os.File
)

// initAuditLog opens the JSONL audit log for appending. A no-op when empty.
func initAuditLog(path string) error {
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	auditMu.Lock()
	defer auditMu.Unlock()
	if auditFile != nil {
		auditFile.Close()
	}
	auditFile = f
	return nil
}

// auditLogRequest appends one JSONL line per request. Best effort: errors are
// never allowed to break request handling.
func auditLogRequest(r *http.Request, status int) {
	auditMu.Lock()
	f := auditFile
	auditMu.Unlock()
	if f == nil {
		return
	}

	user, _ := r.Context().Value(auditUserKey{}).(string)
	ip, port, _ := net.SplitHostPort(r.RemoteAddr)
	entry := struct {
		Time   string `json:"time"`
		IP     string `json:"ip"`
		Port   string `json:"port,omitempty"`
		Method string `json:"method"`
		Path   string `json:"path"`
		Query  string `json:"query,omitempty"`
		Status int    `json:"status"`
		User   string `json:"user,omitempty"`
	}{
		Time:   time.Now().UTC().Format(time.RFC3339Nano),
		IP:     ip,
		Port:   port,
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.RawQuery,
		Status: status,
		User:   user,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	auditMu.Lock()
	defer auditMu.Unlock()
	f.Write(append(data, '\n')) //nolint:errcheck // best-effort audit write
}

// statusRecorder captures the response status code for auditing.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// auditMiddleware logs every completed request routed through it.
func auditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		auditLogRequest(r, rec.status)
	})
}
