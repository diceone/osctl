package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func resetAuthFailures() {
	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	authFailures = make(map[string][]time.Time)
}

func basicAuthRequest(auth string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/ram", nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	handler := basicAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rec, req)
	return rec
}

func validAuthHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:password"))
}

func TestBasicAuthNoHeader(t *testing.T) {
	t.Setenv("OSCTL_USERNAME", "admin")
	t.Setenv("OSCTL_PASSWORD", "password")
	resetAuthFailures()

	rec := basicAuthRequest("")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without Authorization header, got %d", rec.Code)
	}
	if rec.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected WWW-Authenticate header on 401")
	}
}

func TestBasicAuthWrongCredentials(t *testing.T) {
	t.Setenv("OSCTL_USERNAME", "admin")
	t.Setenv("OSCTL_PASSWORD", "password")
	resetAuthFailures()

	bad := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrong"))
	rec := basicAuthRequest(bad)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong password, got %d", rec.Code)
	}
}

func TestBasicAuthValidCredentials(t *testing.T) {
	t.Setenv("OSCTL_USERNAME", "admin")
	t.Setenv("OSCTL_PASSWORD", "password")
	resetAuthFailures()

	rec := basicAuthRequest(validAuthHeader())
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid credentials, got %d", rec.Code)
	}
}

func TestBasicAuthMalformedHeader(t *testing.T) {
	t.Setenv("OSCTL_USERNAME", "admin")
	t.Setenv("OSCTL_PASSWORD", "password")
	resetAuthFailures()

	rec := basicAuthRequest("NotBasic abcdef")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-Basic scheme, got %d", rec.Code)
	}

	rec = basicAuthRequest("Basic !!!not-base64!!!")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for malformed base64, got %d", rec.Code)
	}
}

func TestBasicAuthRateLimit(t *testing.T) {
	t.Setenv("OSCTL_USERNAME", "admin")
	t.Setenv("OSCTL_PASSWORD", "password")
	resetAuthFailures()

	bad := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrong"))
	for i := 0; i < maxAuthFailures; i++ {
		rec := basicAuthRequest(bad)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 on attempt %d, got %d", i+1, rec.Code)
		}
	}

	rec := basicAuthRequest(validAuthHeader())
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after %d failures, got %d", maxAuthFailures, rec.Code)
	}
}

func TestClientIP(t *testing.T) {
	if got := clientIP("192.0.2.10:5555"); got != "192.0.2.10" {
		t.Fatalf("expected host 192.0.2.10, got %q", got)
	}
	if got := clientIP("bogus-addr"); got != "bogus-addr" {
		t.Fatalf("expected raw fallback for malformed RemoteAddr, got %q", got)
	}
}
