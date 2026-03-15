package logging

import (
	"net/url"
	"strings"
)

// RedactURL removes query and fragments (often contain short-lived tokens).
func RedactURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// RedactDSN tries to strip password from "user:pass@tcp(host:port)/db?...".
func RedactDSN(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return ""
	}
	// naive but effective for common mysql DSN.
	// user[:pass]@... -> user@...
	at := strings.Index(dsn, "@")
	if at <= 0 {
		return ""
	}
	cred := dsn[:at]
	rest := dsn[at:]
	if i := strings.Index(cred, ":"); i >= 0 {
		cred = cred[:i]
	}
	return cred + rest
}
