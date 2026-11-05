package salon

import "time"

type Appointment struct {
	StaffID string
	Start   time.Time
	End     time.Time
}

func Conflicts(a, b Appointment) bool {
	return a.StaffID == b.StaffID && a.Start.Before(b.End) && b.Start.Before(a.End)
}

func Commission(amountMinor int64, rate float64) int64 {
	if amountMinor <= 0 || rate <= 0 {
		return 0
	}
	return int64(float64(amountMinor) * rate)
}
