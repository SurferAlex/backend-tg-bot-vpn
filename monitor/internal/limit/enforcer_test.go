package limit

import (
	"context"
	"testing"
	"time"

	"vpn-monitor/internal/redisstore"
	"vpn-monitor/internal/targets"
	"vpn-monitor/internal/vpnapi"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestEnforcer_alertOnly_dedupPerClientIP(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	store := redisstore.New(rdb, 5*time.Minute, 30*time.Minute)
	reg := targets.NewRegistry()
	reg.Replace([]vpnapi.Target{{
		ClientUUID:     "uuid-1",
		XUIClientEmail: "user@test",
		MaxIPs:         2,
	}})

	enc := New(store, reg)

	ctx := context.Background()
	at := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	if _, err := enc.HandleConnection(ctx, "user@test", "1.1.1.1", at); err != nil {
		t.Fatal(err)
	}
	if _, err := enc.HandleConnection(ctx, "user@test", "2.2.2.2", at); err != nil {
		t.Fatal(err)
	}
	ev, err := enc.HandleConnection(ctx, "user@test", "3.3.3.3", at.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if ev == nil {
		t.Fatal("expected limit event")
	}
	if ev.IP != "3.3.3.3" {
		t.Fatalf("excess ip=%s want 3.3.3.3", ev.IP)
	}
	if ev.IPCount != 3 {
		t.Fatalf("ip_count=%d", ev.IPCount)
	}

	ev2, err := enc.HandleConnection(ctx, "user@test", "3.3.3.3", at.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if ev2 != nil {
		t.Fatal("dedup: same client+ip must not produce another event")
	}
}

func TestEnforcer_alertDedupPerClientUUID(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	store := redisstore.New(rdb, 5*time.Minute, 30*time.Minute)
	reg := targets.NewRegistry()
	reg.Replace([]vpnapi.Target{
		{ClientUUID: "uuid-a", XUIClientEmail: "a@test", MaxIPs: 2},
		{ClientUUID: "uuid-b", XUIClientEmail: "b@test", MaxIPs: 2},
	})

	enc := New(store, reg)
	ctx := context.Background()
	at := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	seed := func(email, ip1, ip2, ip3 string) *LimitEvent {
		t.Helper()
		_, _ = enc.HandleConnection(ctx, email, ip1, at)
		_, _ = enc.HandleConnection(ctx, email, ip2, at)
		ev, err := enc.HandleConnection(ctx, email, ip3, at.Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}
		return ev
	}

	evA := seed("a@test", "1.1.1.1", "2.2.2.2", "9.9.9.9")
	if evA == nil {
		t.Fatal("client A: expected event")
	}

	evB := seed("b@test", "4.4.4.4", "5.5.5.5", "9.9.9.9")
	if evB == nil {
		t.Fatal("client B: expected event for same excess IP as A")
	}
	if evB.IP != "9.9.9.9" {
		t.Fatalf("client B excess ip=%s", evB.IP)
	}

	evA2, err := enc.HandleConnection(ctx, "a@test", "9.9.9.9", at.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if evA2 != nil {
		t.Fatal("client A: dedup must suppress repeat alert for same ip")
	}

	evB2, err := enc.HandleConnection(ctx, "b@test", "9.9.9.9", at.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if evB2 != nil {
		t.Fatal("client B: dedup must suppress repeat alert for same ip")
	}
}
