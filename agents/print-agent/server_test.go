package printagent

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type capturePrinter struct {
	payload []byte
	err     error
}

func (p *capturePrinter) Print(_ context.Context, data []byte) error {
	p.payload = append([]byte(nil), data...)
	return p.err
}

func TestLocalServerPrintRequiresTokenAndAllowedOrigin(t *testing.T) {
	printer := &capturePrinter{}
	server := httptest.NewServer((LocalServer{Token: "test-token", AllowedOrigin: "https://pos.example", Printer: printer}).Handler())
	defer server.Close()
	request := func(origin, token string) *http.Response {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/print", strings.NewReader(`{"header":"Shop","lines":["Item 1"],"total":"1 KES"}`))
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}
	resp := request("https://pos.example", "bad")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("bad token status = %d", resp.StatusCode)
	}
	resp.Body.Close()
	resp = request("https://evil.example", "test-token")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("bad origin status = %d", resp.StatusCode)
	}
	resp.Body.Close()
	resp = request("https://pos.example", "test-token")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted || !strings.Contains(string(body), "queued") {
		t.Fatalf("print response = %d %s", resp.StatusCode, body)
	}
	if len(printer.payload) == 0 {
		t.Fatal("printer received no ESC/POS data")
	}
}

func TestEncodeReceiptBoundsAndStripsPrinterCommands(t *testing.T) {
	data, err := EncodeReceipt(Receipt{Header: "Shop\x1b@", Lines: []string{"Line\x1dV"}, Total: "10"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "\x1b@Shop") || strings.Contains(string(data), "Line\x1dV") {
		t.Fatalf("printer control bytes were not stripped: %q", data)
	}
	if _, err = EncodeReceipt(Receipt{Header: "Shop", Total: "10", Lines: []string{strings.Repeat("x", 513)}}); err == nil {
		t.Fatal("oversized line was accepted")
	}
}
