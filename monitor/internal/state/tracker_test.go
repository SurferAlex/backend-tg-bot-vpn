package state

import (
	"testing"
	"time"

	"vpn-monitor/internal/vpnapi"
)

const (
	testEmail = "user-abc12345"
	testUUID  = "550e8400-e29b-41d4-a716-446655440000"
)

func testTargets(maxIPs int) []vpnapi.Target {
	return []vpnapi.Target{{
		ClientUUID:     testUUID,
		MaxIPs:         maxIPs,
		XUIClientEmail: testEmail,
	}}
}

func TestRecord_limit2_noViolationUntilThirdIP(t *testing.T) {
	tr := NewTracker(time.Hour, time.Hour)
	tr.SetTargets(testTargets(2))

	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	for _, ip := range []string{"1.1.1.1", "2.2.2.2"} {
		if _, ok := tr.Record(testEmail, ip, now); ok {
			t.Fatalf("unexpected violation for ip %s", ip)
		}
	}

	v, ok := tr.Record(testEmail, "3.3.3.3", now)
	if !ok {
		t.Fatal("expected violation on third IP")
	}
	if v.ClientUUID != testUUID || v.Email != testEmail {
		t.Fatalf("violation identity: %+v", v)
	}
	if v.MaxIPs != 2 || v.IPCount != 3 {
		t.Fatalf("ip_count=%d max_ips=%d", v.IPCount, v.MaxIPs)
	}
	if len(v.IPs) != 3 {
		t.Fatalf("ips=%v", v.IPs)
	}
}

func TestRecord_sameIPCountedOnce(t *testing.T) {
	tr := NewTracker(time.Hour, time.Hour)
	tr.SetTargets(testTargets(2))
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	tr.Record(testEmail, "1.1.1.1", now)
	if _, ok := tr.Record(testEmail, "1.1.1.1", now.Add(time.Second)); ok {
		t.Fatal("duplicate IP should not trigger violation alone")
	}
	if _, ok := tr.Record(testEmail, "2.2.2.2", now); ok {
		t.Fatal("two unique IPs within limit 2")
	}
}

func TestRecord_unknownEmailIgnored(t *testing.T) {
	tr := NewTracker(time.Hour, time.Hour)
	tr.SetTargets(testTargets(2))
	now := time.Now()

	if _, ok := tr.Record("unknown@x", "1.1.1.1", now); ok {
		t.Fatal("unknown email should be ignored")
	}
}

func TestRecord_dedupSuppressesRepeatViolation(t *testing.T) {
	tr := NewTracker(time.Hour, 10*time.Minute)
	tr.SetTargets(testTargets(2))
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	tr.Record(testEmail, "1.1.1.1", now)
	tr.Record(testEmail, "2.2.2.2", now)
	if _, ok := tr.Record(testEmail, "3.3.3.3", now); !ok {
		t.Fatal("first violation expected")
	}
	if _, ok := tr.Record(testEmail, "4.4.4.4", now.Add(time.Minute)); ok {
		t.Fatal("second violation within dedup window expected suppressed")
	}
}

func TestRecord_dedupAllowsViolationAfterWindow(t *testing.T) {
	tr := NewTracker(time.Hour, 5*time.Minute)
	tr.SetTargets(testTargets(2))
	start := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	tr.Record(testEmail, "1.1.1.1", start)
	tr.Record(testEmail, "2.2.2.2", start)
	tr.Record(testEmail, "3.3.3.3", start)

	later := start.Add(6 * time.Minute)
	if _, ok := tr.Record(testEmail, "4.4.4.4", later); !ok {
		t.Fatal("expected violation after dedup window")
	}
}

func TestRecord_ipWindowPrunesOldIPs(t *testing.T) {
	tr := NewTracker(30*time.Minute, time.Hour)
	tr.SetTargets(testTargets(2))

	t0 := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	tr.Record(testEmail, "1.1.1.1", t0)
	tr.Record(testEmail, "2.2.2.2", t0)

	t1 := t0.Add(31 * time.Minute)
	if _, ok := tr.Record(testEmail, "3.3.3.3", t1); ok {
		t.Fatal("after window prune old IPs gone; single IP should not violate")
	}
	tr.Record(testEmail, "4.4.4.4", t1)
	if _, ok := tr.Record(testEmail, "5.5.5.5", t1); !ok {
		t.Fatal("expected violation when third IP appears in new window")
	}
}

func TestSetTargets_dropsRemovedClientState(t *testing.T) {
	tr := NewTracker(time.Hour, time.Hour)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	tr.SetTargets(testTargets(1))
	tr.Record(testEmail, "1.1.1.1", now)
	tr.Record(testEmail, "2.2.2.2", now) // violation with max 1

	tr.SetTargets(nil)
	if _, ok := tr.Record(testEmail, "3.3.3.3", now); ok {
		t.Fatal("removed target should not produce violation")
	}
}

func TestRecord_emptyEmailOrIPIgnored(t *testing.T) {
	tr := NewTracker(time.Hour, time.Hour)
	tr.SetTargets(testTargets(2))
	now := time.Now()

	if _, ok := tr.Record("", "1.1.1.1", now); ok {
		t.Fatal("empty email")
	}
	if _, ok := tr.Record(testEmail, "", now); ok {
		t.Fatal("empty ip")
	}
}
