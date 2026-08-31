package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPStatusFor(t *testing.T) {
	tests := []struct {
		result string
		want   int
	}{
		{"ok", http.StatusOK},
		{"Total: 8192 MB, Used: 2048 MB, Free: 6144 MB", http.StatusOK},
		{"Usage: osctl [command]", http.StatusBadRequest},
		{"Invalid action 'x'", http.StatusBadRequest},
		{"Unknown command: x", http.StatusBadRequest},
		{"Unsupported OS for package update", http.StatusBadRequest},
		{"Error getting RAM usage: boom", http.StatusInternalServerError},
		{"Failed to stop service nginx. Error: boom", http.StatusInternalServerError},
		// "Failed login attempts" inside the audit summary body must not be
		// mistaken for an error prefix.
		{"=== SECURITY AUDIT SUMMARY ===\nFailed login attempts: 0", http.StatusOK},
		{"Process Count by State:\nRunning: 3", http.StatusOK},
	}

	for _, tt := range tests {
		if got := httpStatusFor(tt.result); got != tt.want {
			t.Errorf("httpStatusFor(%q) = %d, want %d", tt.result, got, tt.want)
		}
	}
}

func TestIsErrorResult(t *testing.T) {
	if isErrorResult("all good") {
		t.Error(`isErrorResult("all good") = true, want false`)
	}
	if !isErrorResult("Failed to do the thing") {
		t.Error(`isErrorResult("Failed to do the thing") = false, want true`)
	}
}

func TestHandleRequestShutdownRejectedOverGET(t *testing.T) {
	// GET must never trigger shutdown/reboot (a browser navigating to the
	// URL must not power off the host).
	req := httptest.NewRequest(http.MethodGet, "/shutdown", nil)
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /shutdown: expected 405, got %d", rec.Code)
	}
	if allow := rec.Header().Get("Allow"); allow != "POST" {
		t.Fatalf("GET /shutdown: expected Allow: POST header, got %q", allow)
	}
}

func TestHandleRequestShutdownRejectedOverGETForRebootToo(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/reboot", nil)
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /reboot: expected 405, got %d", rec.Code)
	}
}

func TestHandleRequestUptime(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/uptime", nil)
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /uptime: expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET /uptime: body is not valid JSON: %v", err)
	}
	if _, ok := body["result"]; !ok {
		t.Fatalf("GET /uptime: expected body with 'result' key, got %s", rec.Body.String())
	}
}

func TestHandleRequestUnknownPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/definitely-not-a-command", nil)
	rec := httptest.NewRecorder()

	handleRequest(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET unknown path: expected 404, got %d", rec.Code)
	}
}
