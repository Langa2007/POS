package dealership

import "testing"

func TestScheduleBalances(t *testing.T) {
	plan, err := Schedule(1001, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	var sum int64
	for _, installment := range plan {
		sum += installment.DueMinor
	}
	if sum != 1001 {
		t.Fatalf("sum=%d", sum)
	}
}
