package happ

import "testing"

func TestResolveOpenTarget_vlessWins(t *testing.T) {
	target, kind := ResolveOpenTarget("vless://u@h:443", "abc", true, "")
	if kind != OpenTargetVless {
		t.Fatalf("kind %v", kind)
	}
	want := "happ://add/vless://u@h:443"
	if target != want {
		t.Fatalf("got %q want %q", target, want)
	}
}

func TestResolveOpenTarget_routing(t *testing.T) {
	target, kind := ResolveOpenTarget("", "abc", true, "")
	if kind != OpenTargetRouting {
		t.Fatalf("kind %v", kind)
	}
	if target != RoutingOnAddPrefix+"abc" {
		t.Fatalf("got %q", target)
	}
}
