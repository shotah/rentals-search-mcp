package rentcast

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultMonthlyQuota = 50

// Usage is a local quota snapshot (RentCast has no public account API).
type Usage struct {
	RequestsUsed     int    `json:"requests_used"`
	RequestsLeft     int    `json:"requests_left"`
	RequestsPerMonth int    `json:"requests_per_month"`
	Period           string `json:"period"`        // YYYY-MM (local calendar month)
	PeriodResets     string `json:"period_resets"` // first day of next calendar month (approx)
	Source           string `json:"source"`        // local_counter
	Persistent       bool   `json:"persistent"`    // false if in-memory only
	Note             string `json:"note,omitempty"`
}

type usageFile struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
}

// UsageTracker counts successful RentCast HTTP calls locally.
type UsageTracker struct {
	mu       sync.Mutex
	path     string
	quota    int
	period   string
	count    int
	loaded   bool
	disabled bool
}

// NewUsageTrackerFromEnv builds a tracker. Disable with RENTCAST_USAGE_TRACK=0.
// Optional: RENTCAST_USAGE_FILE, RENTCAST_MONTHLY_QUOTA (default 50).
func NewUsageTrackerFromEnv() *UsageTracker {
	if disabledEnv(os.Getenv("RENTCAST_USAGE_TRACK")) {
		return &UsageTracker{disabled: true, quota: defaultMonthlyQuota}
	}
	quota := defaultMonthlyQuota
	if v := strings.TrimSpace(os.Getenv("RENTCAST_MONTHLY_QUOTA")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			quota = n
		}
	}
	path := strings.TrimSpace(os.Getenv("RENTCAST_USAGE_FILE"))
	if path == "" {
		if dir, err := os.UserConfigDir(); err == nil && dir != "" {
			path = filepath.Join(dir, "rentals-search-mcp", "usage.json")
		}
	}
	return &UsageTracker{path: path, quota: quota}
}

// NewUsageTrackerForTest creates a tracker writing under path (tests).
func NewUsageTrackerForTest(path string, quota int) *UsageTracker {
	if quota <= 0 {
		quota = defaultMonthlyQuota
	}
	return &UsageTracker{path: path, quota: quota}
}

func disabledEnv(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "0", "false", "off", "no":
		return true
	default:
		return false
	}
}

func (t *UsageTracker) Snapshot() *Usage {
	if t == nil || t.disabled {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureLoadedLocked(time.Now())
	return t.snapshotLocked(time.Now())
}

// RecordSuccess increments after a successful (2xx) RentCast response.
func (t *UsageTracker) RecordSuccess() {
	if t == nil || t.disabled {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	t.ensureLoadedLocked(now)
	period := now.Format("2006-01")
	if t.period != period {
		t.period = period
		t.count = 0
	}
	t.count++
	t.persistLocked()
}

func (t *UsageTracker) ensureLoadedLocked(now time.Time) {
	if t.loaded {
		period := now.Format("2006-01")
		if t.period != "" && t.period != period {
			t.period = period
			t.count = 0
		}
		return
	}
	t.loaded = true
	t.period = now.Format("2006-01")
	if t.path == "" {
		return
	}
	data, err := os.ReadFile(t.path)
	if err != nil {
		return
	}
	var f usageFile
	if json.Unmarshal(data, &f) != nil || f.Period == "" {
		return
	}
	if f.Period == t.period {
		t.count = max(0, f.Count)
	}
}

func (t *UsageTracker) persistLocked() {
	if t.path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(t.path), 0o755); err != nil { //nolint:gosec // G301: local usage dir
		return
	}
	data, err := json.MarshalIndent(usageFile{Period: t.period, Count: t.count}, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(t.path, append(data, '\n'), 0o644) //nolint:gosec // G306: non-secret local counter
}

func (t *UsageTracker) snapshotLocked(now time.Time) *Usage {
	left := max(0, t.quota-t.count)
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	note := "Local counter of successful RentCast calls from this machine/binary. " +
		"Resets on the 1st of each calendar month (local time) — period_resets is that date. " +
		"Free tier ≈ 50/month (~1–2/day) — treat requests_left as a hard budget. " +
		"Not the official dashboard; RentCast's billing cycle may differ. Verify at https://app.rentcast.io/ when unsure."
	if t.path == "" {
		note = "In-memory only (no writable usage file). " + note
	}
	return &Usage{
		RequestsUsed:     t.count,
		RequestsLeft:     left,
		RequestsPerMonth: t.quota,
		Period:           t.period,
		PeriodResets:     next.Format("2006-01-02"),
		Source:           "local_counter",
		Persistent:       t.path != "",
		Note:             note,
	}
}
