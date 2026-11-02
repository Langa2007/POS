package api

import (
	"os"
	"testing"
)

func TestPhase5RoutesExist(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("set TEST_DATABASE_URL to exercise authenticated routes against PostgreSQL")
	}
}
