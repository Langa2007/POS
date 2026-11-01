# ADR 003: IndexedDB for the browser terminal outbox

## Status

Accepted for Phase 3.

## Decision

The Vue terminal uses an IndexedDB outbox through `web/packages/offline`. The
outbox is keyed by the client operation UUID and stores operations until the
authenticated sync endpoint acknowledges them.

## Rationale

IndexedDB is available in supported browsers without a native runtime or a
WebAssembly SQLite download. It supports durable structured records, survives
browser restarts, and leaves the server protocol independent from the browser
storage implementation. A future Tauri shell can replace this package with a
native SQLite implementation behind the same enqueue/list/flush contract.

## Constraints

- Every operation has a client-generated UUID and is idempotent on the server.
- The server remains authoritative for master data.
- Transactions are append-only and receive per-operation acknowledgements.
- The browser never stores provider credentials or access tokens in the outbox.
