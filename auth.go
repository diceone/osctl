package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func getAuthCredentials() (string, string) {
	username := os.Getenv("OSCTL_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("OSCTL_PASSWORD")
	if password == "" {
		password = "password"
	}
	return username, password
}

// Basic auth brute-force protection: max failed attempts per client IP.
const (
	maxAuthFailures       = 10
	authFailureWindow     = 5 * time.Minute
	maxAuthFailureClients = 10000
)

var (
	authFailuresMu sync.Mutex
	authFailures   = make(map[string][]time.Time)

	// Rate-limit state persistence (throttled writes).
	persistPath     string
	lastPersistAt   time.Time
	persistInterval = 2 * time.Second
)

// osctlStateDir returns a writable directory for persistent state: OSCTL_STATE_DIR
// first, then /var/lib/osctl (created if possible), then the temp dir.
func osctlStateDir() string {
	if dir := os.Getenv("OSCTL_STATE_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err == nil {
			return dir
		}
	}
	if err := os.MkdirAll("/var/lib/osctl", 0o700); err == nil {
		return "/var/lib/osctl"
	}
	return os.TempDir()
}

// loadAuthFailures restores the per-IP failure state so restarts do not reset
// brute-force protection.
func loadAuthFailures(dir string) {
	data, err := os.ReadFile(filepath.Join(dir, "auth_failures.json"))
	if err != nil {
		return
	}
	var persisted map[string][]int64
	if err := json.Unmarshal(data, &persisted); err != nil {
		return
	}
	now := time.Now()

	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	persistPath = filepath.Join(dir, "auth_failures.json")
	for ip, stamps := range persisted {
		for _, sec := range stamps {
			t := time.Unix(sec, 0)
			if now.Sub(t) < authFailureWindow {
				authFailures[ip] = append(authFailures[ip], t)
			}
		}
	}
}

// persistAuthFailuresLocked writes the current failure state, at most once per
// persistInterval, so an attacker cannot flood the disk. Caller must hold
// authFailuresMu.
func persistAuthFailuresLocked() {
	if persistPath == "" || time.Since(lastPersistAt) < persistInterval {
		return
	}
	lastPersistAt = time.Now()

	persisted := make(map[string][]int64, len(authFailures))
	for ip, stamps := range authFailures {
		for _, t := range stamps {
			persisted[ip] = append(persisted[ip], t.Unix())
		}
	}
	data, err := json.Marshal(persisted)
	if err != nil {
		return
	}
	tmp := persistPath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err == nil {
		os.Rename(tmp, persistPath) //nolint:errcheck // best-effort persistence
	}
}

// clientIP extracts the host portion of r.RemoteAddr, falling back to the raw value.
func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func pruneAuthFailures(now time.Time) {
	for ip, stamps := range authFailures {
		recent := stamps[:0]
		for _, t := range stamps {
			if now.Sub(t) < authFailureWindow {
				recent = append(recent, t)
			}
		}
		if len(recent) == 0 {
			delete(authFailures, ip)
		} else {
			authFailures[ip] = recent
		}
	}
}

func allowAuthAttempt(ip string) bool {
	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	pruneAuthFailures(time.Now())
	return len(authFailures[ip]) < maxAuthFailures
}

func recordAuthFailure(ip string) {
	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	if len(authFailures) >= maxAuthFailureClients {
		authFailures = make(map[string][]time.Time)
	}
	authFailures[ip] = append(authFailures[ip], time.Now())
	persistAuthFailuresLocked()
}

func clearAuthFailures(ip string) {
	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	delete(authFailures, ip)
	persistAuthFailuresLocked()
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r.RemoteAddr)
		if !allowAuthAttempt(ip) {
			http.Error(w, "Too many failed authentication attempts. Try again later.", http.StatusTooManyRequests)
			return
		}

		auth := r.Header.Get("Authorization")
		const prefix = "Basic "

		// Alternative auth scheme: static API token (OSCTL_API_TOKEN),
		// compared in constant time as well.
		if token := os.Getenv("OSCTL_API_TOKEN"); token != "" && strings.HasPrefix(auth, "Bearer ") {
			presented := strings.TrimSpace(auth[len("Bearer "):])
			if subtle.ConstantTimeCompare([]byte(presented), []byte(token)) == 1 {
				clearAuthFailures(ip)
				next.ServeHTTP(w, r.WithContext(withAuditUser(r.Context(), "api-token")))
				return
			}
			recordAuthFailure(ip)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(auth[len(prefix):]))
		if err != nil {
			recordAuthFailure(ip)
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, password := getAuthCredentials()
		pair := strings.SplitN(string(payload), ":", 2)

		// Compare SHA-256 digests in constant time so the
		// response cannot be used to guess credentials byte by byte.
		givenUser := sha256.Sum256([]byte(pair[0]))
		givenPass := sha256.Sum256([]byte(pair[1]))
		wantUser := sha256.Sum256([]byte(username))
		wantPass := sha256.Sum256([]byte(password))
		userOK := len(pair) == 2 && subtle.ConstantTimeCompare(givenUser[:], wantUser[:]) == 1
		passOK := len(pair) == 2 && subtle.ConstantTimeCompare(givenPass[:], wantPass[:]) == 1

		if !userOK || !passOK {
			recordAuthFailure(ip)
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		clearAuthFailures(ip)
		next.ServeHTTP(w, r.WithContext(withAuditUser(r.Context(), pair[0])))
	})
}
