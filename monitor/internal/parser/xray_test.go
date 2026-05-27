package parser

import "testing"

func TestParseAccessLine(t *testing.T) {
	line := `2026/05/20 12:00:00 from 203.0.113.10:54321 accepted tcp:example.com:443 email: testuser-abc12345`
	email, ip, ok := ParseAccessLine(line)
	if !ok {
		t.Fatal("expected ok")
	}
	if email != "testuser-abc12345" {
		t.Fatalf("email=%q", email)
	}
	if ip != "203.0.113.10" {
		t.Fatalf("ip=%q", ip)
	}
}
