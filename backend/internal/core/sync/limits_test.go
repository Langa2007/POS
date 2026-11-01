package sync

import (
	"strings"
	"testing"
	"time"
)

func TestOfflinePolicyBlocksExpiredTerminal(t *testing.T) {
	now := time.Now()
	policy := OfflinePolicy{MaxDuration: time.Hour, MaxSaleMinor: 10_000}
	err := policy.AllowsSale(now.Add(-2*time.Hour), 100, false, false, now)
	if err == nil || !strings.Contains(err.Error(), "duration") {
		t.Fatalf("error=%v", err)
	}
}
