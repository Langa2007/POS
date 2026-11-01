package mpesa

import (
	"context"
	"testing"
)

func TestSimulatorSupportsSuccessAndFailureModes(t *testing.T) {
	simulator := &Simulator{}
	response, err := simulator.InitiateSTK(context.Background(), STKRequest{Phone: "254700000000", AmountMinor: 100})
	if err != nil || response.ResponseCode != "0" {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	simulator.Mode = "failed"
	status, err := simulator.QueryStatus(context.Background(), response.CheckoutRequestID)
	if err != nil || status != "Failed" {
		t.Fatalf("status=%s err=%v", status, err)
	}
}

func TestSimulatorCallbackDeliveryModes(t *testing.T) {
	callback := Callback{CheckoutRequestID: "checkout-1", ResultCode: 0}
	sim := &Simulator{Mode: "lost_callback"}
	if got := sim.DeliverCallback(callback); len(got) != 0 {
		t.Fatalf("lost callback delivered %d times", len(got))
	}
	sim.Mode = "duplicate_callback"
	got := sim.DeliverCallback(callback)
	if len(got) != 2 || got[0] != got[1] {
		t.Fatalf("duplicate callback sequence = %#v", got)
	}
}
