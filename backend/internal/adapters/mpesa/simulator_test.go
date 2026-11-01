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
