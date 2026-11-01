# POS local receipt printer

This small service sends ESC/POS receipts to a TCP thermal printer (often port 9100). It listens on loopback only and requires a bearer token for print requests.

```sh
export POS_PRINT_TOKEN="$(openssl rand -hex 32)"
export POS_PRINT_ALLOWED_ORIGIN="http://localhost:5173" # use the deployed POS origin in production
export POS_PRINTER_ADDRESS="192.168.1.50:9100"
go run ./agents/print-agent/cmd/print-agent
```

The service defaults to `127.0.0.1:8765`. `POS_PRINT_LISTEN` may select another loopback IP and port; non-loopback binds are rejected. `/health` is unauthenticated for local health checks; `/print` requires the token and exact configured browser origin. Keep the token private and configure a distinct random token for each terminal. The service does not provide TLS; loopback binding is required. Do not expose the listener through a reverse proxy or LAN interface.

The first driver is raw TCP ESC/POS. USB, Bluetooth, Android, drawer kick, scales, and printer-specific character encoding still need platform drivers and physical-device validation. The browser print dialog remains the fallback.
