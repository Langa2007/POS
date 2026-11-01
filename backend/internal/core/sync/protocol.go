package sync

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
)

type Operation struct {
	ClientID         string
	TerminalID       string
	EntityType       string
	EntityID         string
	Operation        string
	Payload          []byte
	ClientOccurredAt time.Time
}

type Ack struct {
	ClientID string `json:"client_id"`
	EntityID string `json:"entity_id,omitempty"`
	Status   string `json:"status"`
	Error    *Error `json:"error,omitempty"`
	Result   any    `json:"result,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Change struct {
	Cursor     int64
	EntityType string
	EntityID   string
	Operation  string
	Payload    []byte
	OccurredAt time.Time
}

func ValidateOperation(operation Operation, now time.Time) error {
	if strings.TrimSpace(operation.ClientID) == "" {
		return errors.New("client_id is required")
	}
	if strings.TrimSpace(operation.EntityType) == "" || strings.TrimSpace(operation.EntityID) == "" {
		return errors.New("entity_type and entity_id are required")
	}
	switch operation.Operation {
	case OperationCreate, OperationUpdate, OperationDelete:
	default:
		return fmt.Errorf("unsupported operation %q", operation.Operation)
	}
	if operation.Operation != OperationDelete && len(operation.Payload) == 0 {
		return errors.New("payload is required for create and update")
	}
	if operation.ClientOccurredAt.IsZero() {
		return errors.New("client_occurred_at is required")
	}
	if operation.ClientOccurredAt.After(now.Add(5 * time.Minute)) {
		return errors.New("client clock is too far ahead")
	}
	return nil
}

func OrderOperations(operations []Operation) ([]Operation, error) {
	ordered := append([]Operation(nil), operations...)
	return SortOperations(ordered), nil
}

func SortOperations(operations []Operation) []Operation {
	ordered := append([]Operation(nil), operations...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ClientOccurredAt.Equal(ordered[j].ClientOccurredAt) {
			return ordered[i].ClientID < ordered[j].ClientID
		}
		return ordered[i].ClientOccurredAt.Before(ordered[j].ClientOccurredAt)
	})
	return ordered
}

func NextCursor(changes []Change, current int64) int64 {
	next := current
	for _, change := range changes {
		if change.Cursor > next {
			next = change.Cursor
		}
	}
	return next
}
