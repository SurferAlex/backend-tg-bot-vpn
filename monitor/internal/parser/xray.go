package parser

import (
	"regexp"
	"strings"
)

var (
	emailRe = regexp.MustCompile(`(?i)email:\s*([^\s\]]+)`)
	fromIPRe = regexp.MustCompile(`(?i)(?:\bfrom\b|\bsrc_ip\b)\s+(\d{1,3}(?:\.\d{1,3}){3})`)
)

// ParseAccessLine extracts x-ui client email and source IP from an Xray access log line.
func ParseAccessLine(line string) (email, ip string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}

	em := emailRe.FindStringSubmatch(line)
	if len(em) < 2 {
		return "", "", false
	}
	email = strings.TrimSpace(em[1])
	if email == "" {
		return "", "", false
	}

	ipm := fromIPRe.FindStringSubmatch(line)
	if len(ipm) < 2 {
		return "", "", false
	}
	return email, ipm[1], true
}
