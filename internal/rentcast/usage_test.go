package rentcast

import (
	"path/filepath"
	"testing"
)

func TestUsageTrackerPersistAndResetMonth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.json")
	tr := NewUsageTrackerForTest(path, 50)
	tr.RecordSuccess()
	tr.RecordSuccess()
	snap := tr.Snapshot()
	if snap.RequestsUsed != 2 || snap.RequestsLeft != 48 || !snap.Persistent {
		t.Fatalf("%+v", snap)
	}

	tr2 := NewUsageTrackerForTest(path, 50)
	snap2 := tr2.Snapshot()
	if snap2.RequestsUsed != 2 {
		t.Fatalf("reload got %+v", snap2)
	}
}

func TestUsageTrackerDisabled(t *testing.T) {
	t.Setenv("RENTCAST_USAGE_TRACK", "0")
	tr := NewUsageTrackerFromEnv()
	tr.RecordSuccess()
	if tr.Snapshot() != nil {
		t.Fatal("expected nil when disabled")
	}
}

func TestUsageTrackerQuotaEnv(t *testing.T) {
	t.Setenv("RENTCAST_USAGE_TRACK", "1")
	t.Setenv("RENTCAST_MONTHLY_QUOTA", "10")
	t.Setenv("RENTCAST_USAGE_FILE", filepath.Join(t.TempDir(), "u.json"))
	tr := NewUsageTrackerFromEnv()
	for range 3 {
		tr.RecordSuccess()
	}
	snap := tr.Snapshot()
	if snap.RequestsPerMonth != 10 || snap.RequestsLeft != 7 {
		t.Fatalf("%+v", snap)
	}
}
