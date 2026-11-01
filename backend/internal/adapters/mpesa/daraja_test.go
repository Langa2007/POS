package mpesa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestDarajaOAuthAndSTKRequest(t *testing.T) {
	var gotSTK map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/v1/generate":
			if u, p, ok := r.BasicAuth(); !ok || u != "key" || p != "secret" {
				t.Errorf("basic auth=%q %q", u, p)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token", "expires_in": "3600"})
		case "/mpesa/stkpush/v1/processrequest":
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Errorf("authorization=%q", r.Header.Get("Authorization"))
			}
			_ = json.NewDecoder(r.Body).Decode(&gotSTK)
			_ = json.NewEncoder(w).Encode(STKResponse{CheckoutRequestID: "ws_CO_123", MerchantRequestID: "mr_123", ResponseCode: "0"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	d, _ := NewDaraja(DarajaConfig{Environment: "sandbox", ConsumerKey: "key", ConsumerSecret: "secret", BusinessShortCode: "174379", Passkey: "pass", CallbackURL: "https://pos.example/callback"}, server.Client())
	d.now = func() time.Time { return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC) }
	// The API host is fixed by environment; route through the test transport while retaining paths.
	d.Client = &http.Client{Transport: rewriteHost{target: server.URL}}
	resp, err := d.InitiateSTK(context.Background(), STKRequest{Phone: "0700000000", AmountMinor: 12500, AccountReference: "sale-123", Description: "POS sale"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CheckoutRequestID != "ws_CO_123" {
		t.Fatalf("response=%+v", resp)
	}
	if gotSTK["PhoneNumber"] != "254700000000" || gotSTK["Amount"] != "125" || gotSTK["CallBackURL"] != "https://pos.example/callback" {
		t.Fatalf("request payload=%v", gotSTK)
	}
}

type rewriteHost struct{ target string }

func (rt rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	copy := req.Clone(req.Context())
	base, _ := url.Parse(rt.target)
	copy.URL.Scheme = base.Scheme
	copy.URL.Host = base.Host
	return http.DefaultTransport.RoundTrip(copy)
}

func TestDarajaParsesSafaricomCallbackAndNormalizesPhone(t *testing.T) {
	d := &Daraja{}
	callback, err := d.ParseCallback([]byte(`{"Body":{"stkCallback":{"CheckoutRequestID":"ws_CO_123","ResultCode":0,"ResultDesc":"Success","CallbackMetadata":{"Item":[{"Name":"Amount","Value":125},{"Name":"MpesaReceiptNumber","Value":"ABC123"},{"Name":"PhoneNumber","Value":254700000000}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	if callback.AmountMinor != 12500 || callback.ReceiptNumber != "ABC123" || callback.Phone != "254700000000" {
		t.Fatalf("callback=%+v", callback)
	}
}
