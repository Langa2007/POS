package payments

import "errors"

type StatusQuerier interface {
	Status(reference string) (Status, error)
}

type ReconcileDecision struct {
	Reference string
	From      Status
	To        Status
}

func ResolveStuckIntent(reference string, current Status, queried Status) (ReconcileDecision, error) {
	if reference == "" {
		return ReconcileDecision{}, errors.New("payment reference is required")
	}
	if current != Pending {
		return ReconcileDecision{Reference: reference, From: current, To: current}, nil
	}
	if queried != Captured && queried != Failed && queried != Cancelled {
		return ReconcileDecision{}, errors.New("provider status is not terminal")
	}
	if err := Transition(current, queried); err != nil {
		return ReconcileDecision{}, err
	}
	return ReconcileDecision{Reference: reference, From: current, To: queried}, nil
}
