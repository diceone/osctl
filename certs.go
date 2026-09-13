package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// certExpiryWarningDays is how soon expiry must be before a cert is flagged.
const certExpiryWarningDays = 30

// certTarget is one certificate checked by `osctl certs`.
type certTarget struct {
	source   string // file path or host:port
	subject  string
	notAfter time.Time
}

// checkCertificates inspects TLS certificates (PEM files or host:port
// targets) and reports expiry status, updating Prometheus gauges.
func checkCertificates(targets []string) string {
	paths := targets
	if len(paths) == 0 {
		paths = defaultCertLocations()
	}
	if len(paths) == 0 {
		return "No TLS certificates found in the default locations. Pass certificate file paths or host:port targets."
	}

	var found []certTarget
	var problems []string
	seen := map[string]bool{}

	for _, target := range paths {
		if seen[target] {
			continue
		}
		seen[target] = true
		info, err := inspectCertTarget(target)
		if err != nil {
			problems = append(problems, fmt.Sprintf("  %s: %v", target, err))
			continue
		}
		found = append(found, info)
		certExpirySeconds.WithLabelValues(target).Set(float64(info.notAfter.Unix()))
	}

	var output strings.Builder
	output.WriteString("Certificate Expiry Check:\n\n")

	now := time.Now()
	expiring, expired := 0, 0
	for _, c := range found {
		days := int(time.Until(c.notAfter).Hours() / 24)
		status := "OK"
		switch {
		case now.After(c.notAfter):
			status = "EXPIRED"
			expired++
		case days < certExpiryWarningDays:
			status = "EXPIRING SOON"
			expiring++
		}
		output.WriteString(c.source + "\n")
		output.WriteString(fmt.Sprintf("  Subject: %s\n", c.subject))
		output.WriteString(fmt.Sprintf("  Expires: %s (%d days)\n", c.notAfter.UTC().Format("2006-01-02 15:04 MST"), days))
		output.WriteString(fmt.Sprintf("  Status: %s\n\n", status))
	}

	if len(problems) > 0 {
		output.WriteString("Could not inspect:\n")
		for _, p := range problems {
			output.WriteString(p + "\n")
		}
		output.WriteString("\n")
	}

	output.WriteString(fmt.Sprintf("Summary: %d certificates checked, %d expiring within %d days, %d expired.",
		len(found), expiring, certExpiryWarningDays, expired))
	return output.String()
}

// defaultCertLocations expands the standard certificate locations scanned
// when no explicit targets are passed.
func defaultCertLocations() []string {
	var paths []string
	for _, pattern := range []string{
		"/etc/letsencrypt/live/*/fullchain.pem",
		"/etc/letsencrypt/live/*/cert.pem",
		"/etc/ssl/certs/*.crt",
		"/etc/pki/tls/certs/*.crt",
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		paths = append(paths, matches...)
	}
	return paths
}

// inspectCertTarget reads the first certificate of a PEM file or performs a
// TLS handshake when the target is host:port.
func inspectCertTarget(target string) (certTarget, error) {
	data, err := os.ReadFile(target) //nolint:gosec // explicit user-provided path
	if err == nil {
		cert, err := parseFirstCertificate(data)
		if err != nil {
			return certTarget{}, err
		}
		return certTarget{source: target, subject: cert.Subject.CommonName, notAfter: cert.NotAfter}, nil
	}
	if os.IsPermission(err) {
		return certTarget{}, fmt.Errorf("cannot read %s: permission denied (try sudo)", target)
	}
	if os.IsNotExist(err) && !looksLikeHostPort(target) {
		return certTarget{}, fmt.Errorf("cannot read %s: file not found", target)
	}
	return inspectRemoteCert(target)
}

// looksLikeHostPort reports whether a target should be treated as a network
// endpoint rather than a file path.
func looksLikeHostPort(target string) bool {
	if strings.Contains(target, ":") {
		return true
	}
	// A bare hostname like "example.com" has no path separators and no file
	// extension; treat it as a TLS host.
	return !strings.HasPrefix(target, "/") && !strings.Contains(target, string(os.PathSeparator)) && filepath.Ext(target) == ""
}

func parseFirstCertificate(data []byte) (*x509.Certificate, error) {
	rest := data
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, fmt.Errorf("no PEM certificate found")
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %v", err)
		}
		return cert, nil
	}
}

// inspectRemoteCert performs an insecure TLS handshake to read the served
// certificate; verification is skipped because only the expiry is inspected.
func inspectRemoteCert(hostPort string) (certTarget, error) {
	if !strings.Contains(hostPort, ":") {
		hostPort = net.JoinHostPort(hostPort, "443")
	}
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostPort, &tls.Config{InsecureSkipVerify: true}) //nolint:gosec
	if err != nil {
		return certTarget{}, fmt.Errorf("TLS handshake failed: %v", err)
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return certTarget{}, fmt.Errorf("no certificate served")
	}
	return certTarget{source: hostPort, subject: certs[0].Subject.CommonName, notAfter: certs[0].NotAfter}, nil
}
