package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"os"
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
)

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
}

func clearAuthFailures(ip string) {
	authFailuresMu.Lock()
	defer authFailuresMu.Unlock()
	delete(authFailures, ip)
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
		next.ServeHTTP(w, r)
	})
}
