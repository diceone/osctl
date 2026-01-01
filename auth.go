package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
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

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username, password := getAuthCredentials()
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="osctl"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
