package printagent

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type ReceiptPrinter interface {
	Print(context.Context, []byte) error
}

// TCPPrinter sends raw ESC/POS bytes to a network printer (commonly port 9100).
type TCPPrinter struct {
	Address string
	Timeout time.Duration
}

func (p TCPPrinter) Print(ctx context.Context, data []byte) error {
	if strings.TrimSpace(p.Address) == "" {
		return errors.New("printer address is required")
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", p.Address)
	if err != nil {
		return err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	for len(data) > 0 {
		n, e := conn.Write(data)
		if e != nil {
			return e
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

type LocalServer struct {
	Token         string
	AllowedOrigin string
	Printer       ReceiptPrinter
}

func (s LocalServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	mux.HandleFunc("OPTIONS /print", func(w http.ResponseWriter, r *http.Request) {
		if !s.originAllowed(w, r) {
			http.Error(w, "origin denied", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /print", func(w http.ResponseWriter, r *http.Request) {
		if !s.originAllowed(w, r) {
			http.Error(w, "origin denied", http.StatusForbidden)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if s.Token == "" || len(token) != len(s.Token) || subtle.ConstantTimeCompare([]byte(token), []byte(s.Token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if s.Printer == nil {
			http.Error(w, "printer unavailable", http.StatusServiceUnavailable)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		var receipt Receipt
		if json.NewDecoder(r.Body).Decode(&receipt) != nil {
			http.Error(w, "invalid receipt", http.StatusBadRequest)
			return
		}
		payload, err := EncodeReceipt(receipt)
		if err != nil {
			http.Error(w, "invalid receipt", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 7*time.Second)
		defer cancel()
		if err = s.Printer.Print(ctx, payload); err != nil {
			http.Error(w, "printer write failed", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"queued"}`))
	})
	return mux
}

func (s LocalServer) originAllowed(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if s.AllowedOrigin == "" || origin != s.AllowedOrigin {
		return false
	}
	w.Header().Set("Access-Control-Allow-Origin", s.AllowedOrigin)
	w.Header().Set("Vary", "Origin")
	return true
}
