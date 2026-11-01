package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	printagent "github.com/Langa2007/POS/agents/print-agent"
)

func main() {
	address := env("POS_PRINT_LISTEN", "127.0.0.1:8765")
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		log.Fatalf("invalid POS_PRINT_LISTEN: %v", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		log.Fatal("POS_PRINT_LISTEN must use a loopback IP address")
	}
	token := os.Getenv("POS_PRINT_TOKEN")
	if len(token) < 32 {
		log.Fatal("POS_PRINT_TOKEN must contain at least 32 characters")
	}
	origin := strings.TrimSpace(os.Getenv("POS_PRINT_ALLOWED_ORIGIN"))
	if origin == "" {
		log.Fatal("POS_PRINT_ALLOWED_ORIGIN is required")
	}
	printer := printagent.TCPPrinter{Address: os.Getenv("POS_PRINTER_ADDRESS"), Timeout: 5 * time.Second}
	srv := &http.Server{Addr: address, Handler: (printagent.LocalServer{Token: token, AllowedOrigin: origin, Printer: printer}).Handler(), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 8 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()
	log.Printf("local receipt printer listening at %s", address)
	if err = srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
