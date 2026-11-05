package api

import (
	"os"
	"testing"
)

func TestPhase8PostgresGateConfigured(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL Phase 8 integration tests")
	}
}
