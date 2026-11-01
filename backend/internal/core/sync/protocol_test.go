package sync

import (
	"testing"
	"time"
)

func TestOrderOperationsIsDeterministic(t *testing.T) {
	now := time.Now().UTC()
	ordered, err := OrderOperations([]Operation{
		{ClientID: "b", EntityType: "sale", EntityID: "2", Operation: OperationCreate, Payload: []byte(`{}`), ClientOccurredAt: now},
		{ClientID: "a", EntityType: "sale", EntityID: "1", Operation: OperationCreate, Payload: []byte(`{}`), ClientOccurredAt: now},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ordered[0].ClientID != "a" {
		t.Fatalf("first operation=%q, want a", ordered[0].ClientID)
	}
}

func TestNextCursorNeverMovesBackwards(t *testing.T) {
	next := NextCursor([]Change{{Cursor: 4}, {Cursor: 2}}, 5)
	if next != 5 {
		t.Fatalf("next cursor=%d, want 5", next)
	}
}
