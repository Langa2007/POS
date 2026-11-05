package salon

import (
	"testing"
	"time"
)

func TestAppointmentConflict(t *testing.T) {
	start := time.Now()
	if !Conflicts(Appointment{"s", start, start.Add(time.Hour)}, Appointment{"s", start.Add(30 * time.Minute), start.Add(2 * time.Hour)}) {
		t.Fatal("conflict not detected")
	}
}
