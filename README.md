# Universal POS Platform

Universal POS is a modular, offline-aware point-of-sale platform for businesses
that sell products, services, rooms, meals, or memberships. It uses one shared
backend and a Vue frontend that can be shaped with business-specific modules.

The project is designed as a **modular monolith**: shared capabilities such as
catalog, inventory, sales, payments, customers, tenancy, and synchronization
are reused across business types, while vertical modules add their own
workflows and screens.

## Current status

The frontend currently has four business modules wired into the running POS:

- **Retail** — weighed goods, promotions, loyalty, perishables, and shelf labels
- **Pharmacy** — prescriptions, FEFO batch tracking, controlled drugs, and
  clinical dispensing labels
- **Hardware and building supplies** — multi-unit sales, cut-to-length
  materials, trade credit, and delivery
- **Restaurant and bar** — table service, kitchen orders, menu modifiers, and
  split bills
- **Bookshop** — ISBN catalog, special and school orders, consignments, and
  publisher returns
- **Salon and spa** — appointments, service packages, stylist commissions, and
  client loyalty

Additional verticals are declared in `web/modules/` and are at different stages
of implementation:

- Restaurant
- Hospitality / hotel
- Hardware and building supplies
- Bookshop
- Dealership
- Fuel station
- Garage / auto service
- Membership
- Salon and spa

The module manifests describe the intended capabilities, routes, permissions,
and UI slots. The active frontend loader enables Retail, Pharmacy, Hardware, Restaurant,
Bookshop, and Salon. The other modules should be treated as work in progress until
they are added to the loader and their views are complete.

## Architecture

```text
Vue 3 POS shell
  ├── shared checkout, back office, authentication, and offline state
  ├── dynamic vertical modules
  │   ├── retail
  │   └── pharmacy
  └── module slots for checkout and other shared screens
           │
           ▼
Go API
  ├── tenancy and authentication
  ├── catalog, inventory, sales, payments, and supply
  ├── fiscal and M-Pesa integrations
  └── vertical domain modules
           │
           ▼
PostgreSQL + object storage
```

The frontend is a pnpm workspace under `web/`. Modules use the shared
`@pos/module-sdk` contract and can contribute routes, navigation, translations,
capabilities, and slot components without modifying the core checkout.

The platform supports an online API-backed mode and local cached/offline
operation. Pending operations are synchronized when connectivity returns.

## Technology

- **Frontend:** Vue 3, Vite, Vue Router, Pinia, Vue I18n
- **Backend:** Go
- **Database:** PostgreSQL
- **Object storage:** S3-compatible storage, SeaweedFS locally
- **Payments:** M-Pesa integration
- **Workspace tooling:** pnpm, Corepack, Task
- **Testing:** Vitest and Go tests
- **Deployment support:** Docker Compose

## Prerequisites

Install:

- Go
- Docker Engine and Docker Compose
- Node.js with Corepack enabled
- `go-task`

The commands below are for local development. Customer deployments are
intended to use a hosted service or a packaged desktop/on-premise installer;
customers should not need Go, Docker, or pnpm.

## Quick start

From the repository root:

```bash
task bootstrap
task bootstrap-owner
task dev
```

`task bootstrap` installs frontend dependencies, starts local support services,
and applies database migrations. `task bootstrap-owner` creates the initial
tenant and owner from the configured environment. `task dev` starts the Go API
and Vue development server.

The frontend is available at:

```text
http://localhost:5173
```

The API runs on port `8080` by default. Vite proxies `/api` and `/health` to
the API during development.

For detailed environment, database, onboarding, and deployment instructions,
see [INSTALLATION.md](./INSTALLATION.md).

## Frontend commands

Run these from the repository root:

```bash
# Start the frontend only
corepack pnpm dev

# Run frontend tests
corepack pnpm test

# Build the frontend
corepack pnpm build
```

The root shortcuts are also available:

```bash
corepack pnpm dev
corepack pnpm test
corepack pnpm build
```

## Repository layout

```text
backend/             Go API, domain modules, adapters, and database access
contracts/           OpenAPI and synchronization schemas
web/apps/pos/        Main Vue POS application
web/modules/         Business-specific frontend modules
web/packages/        Shared frontend SDK, UI, and offline packages
agents/print-agent/  Receipt and printer integration service
deploy/              Database and object-storage initialization
presets/             Example business capability presets
scripts/             Environment and development helpers
docs/                Architecture decision records and setup documentation
```

## Adding a business vertical

Create a module under `web/modules/<id>/` containing its manifest and module
entry point. A module can provide:

- capabilities
- authenticated routes
- navigation entries
- translations
- checkout or workspace slot contributions
- business-specific views and components

Then register it in
[`web/apps/pos/src/moduleLoader.js`](./web/apps/pos/src/moduleLoader.js) and
ensure its backend permissions, API contracts, and tenant capabilities are
available. Prefer capability checks over hard-coded business-type checks.

## Project documentation

- [Installation and local onboarding](./INSTALLATION.md)
- [Repository structure](./STRUCTURE.md)
- [System blueprint](./POSSYSTEM.md)
- [Development phases](./PHASES.md)

## License

No license has been declared in this repository yet.
