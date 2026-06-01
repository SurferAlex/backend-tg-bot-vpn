package happ

import (
	"strings"
	"testing"
)

func TestRoutingDeeplink_onadd(t *testing.T) {
	got := RoutingDeeplink("abc123", true)
	if got != RoutingOnAddPrefix+"abc123" {
		t.Fatalf("got %q", got)
	}
}

func TestParseRoutingInput_happLink(t *testing.T) {
	link := RoutingOnAddPrefix + DefaultRoutingProfileB64
	b64, err := ParseRoutingInput(link)
	if err != nil {
		t.Fatal(err)
	}
	if b64 != DefaultRoutingProfileB64 {
		t.Fatalf("got %q", b64)
	}
}

func TestOpenRedirectURL_routingParam(t *testing.T) {
	got := OpenRedirectURL("https://host/happ/open", "abc")
	if !strings.Contains(got, "routing=abc") {
		t.Fatalf("got %q", got)
	}
}
