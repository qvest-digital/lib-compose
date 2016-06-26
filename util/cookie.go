package util

import (
	"net/http"
	"strings"
)

// taken and adapted from net/http
func ReadCookieValue(h http.Header, cookieName string) (string, bool) {
	lines, ok := h["Cookie"]
	if !ok {
		return "", false
	}

	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ";")
		if len(parts) == 1 && parts[0] == "" {
			continue
		}
		for i := 0; i < len(parts); i++ {
			parts[i] = strings.TrimSpace(parts[i])
			if len(parts[i]) == 0 {
				continue
			}
			name, val := parts[i], ""
			if j := strings.Index(name, "="); j >= 0 {
				name, val = name[:j], name[j+1:]
			}
			if cookieName == name {
				if len(val) > 1 && val[0] == '"' && val[len(val)-1] == '"' {
					val = val[1 : len(val)-1]
				}
				return val, true
			}
		}
	}
	return "", false
}
