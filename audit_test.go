package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLogWritesJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	if err := initAuditLog(path); err != nil {
		t.Fatalf("initAuditLog: %v", err)
	}
	t.Cleanup(func() { initAuditLog("") })

	if err := initAuditLog(path); err != nil {
		t.Fatalf("initAuditLog: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/ram?verbose=1", nil)
	req.RemoteAddr = "203.0.113.7:44321"
	req = req.WithContext(withAuditUser(req.Context(), "admin"))

	auditLogRequest(req, http.StatusOK)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("audit log not written: %v", err)
	}
	line := strings.TrimSpace(string(data))

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("audit line is not JSON: %q", line)
	}
	checks := map[string]interface{}{
		"ip": "203.0.113.7", "method": "GET", "path": "/ram",
		"status": float64(200), "user": "admin", "port": "44321",
	}
	for k, want := range checks {
		if entry[k] != want {
			t.Errorf("audit %s = %v, want %v", k, entry[k], want)
		}
	}
}

func TestAuditDisabledWhenNoPath(t *testing.T) {
	if err := initAuditLog(""); err != nil {
		t.Fatalf("empty path must be a no-op, got %v", err)
	}
	// Must not panic without an open audit file.
	req := httptest.NewRequest(http.MethodGet, "/ram", nil)
	auditLogRequest(req, http.StatusOK)
}

func TestStatusRecorderCapturesStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	sr := &statusRecorder{ResponseWriter: rec, status: http.StatusOK}

	sr.WriteHeader(http.StatusTeapot)

	if sr.status != http.StatusTeapot {
		t.Fatalf("statusRecorder.status = %d, want 418", sr.status)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("underlying writer got %d, want 418", rec.Code)
	}
}

func TestAuditMiddlewareLogsDownstreamStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	if err := initAuditLog(path); err != nil {
		t.Fatalf("initAuditLog: %v", err)
	}
	t.Cleanup(func() { initAuditLog("") })

	handler := auditMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("audit log not written: %v", err)
	}
	if !strings.Contains(string(data), `"status":404`) {
		t.Fatalf("audit log missing status 404, got %s", string(data))
	}
}
