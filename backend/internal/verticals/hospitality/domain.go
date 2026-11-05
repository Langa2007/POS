package hospitality

import (
	"errors"
	"time"
)

var ErrInvalidStay = errors.New("checkout must be after check-in")

type Stay struct {
	CheckIn  time.Time
	CheckOut time.Time
}

func Overlaps(a, b Stay) bool {
	return a.CheckIn.Before(b.CheckOut) && b.CheckIn.Before(a.CheckOut)
}

func ValidateStay(stay Stay) error {
	if !stay.CheckOut.After(stay.CheckIn) {
		return ErrInvalidStay
	}
	return nil
}

func NightCount(stay Stay) (int, error) {
	if err := ValidateStay(stay); err != nil {
		return 0, err
	}
	return int(stay.CheckOut.Sub(stay.CheckIn).Hours() / 24), nil
}

type FolioEntry struct {
	ID          string
	AmountMinor int64
}

func AppendOnce(entries []FolioEntry, entry FolioEntry) []FolioEntry {
	for _, existing := range entries {
		if existing.ID == entry.ID {
			return entries
		}
	}
	return append(entries, entry)
}
