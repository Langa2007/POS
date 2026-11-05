CREATE TABLE IF NOT EXISTS vertical_modules (
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id text NOT NULL,
    version text NOT NULL,
    enabled boolean NOT NULL DEFAULT false,
    enabled_at timestamptz,
    disabled_at timestamptz,
    PRIMARY KEY (tenant_id, module_id)
);

INSERT INTO capability_definitions (id, name, description, requires, conflicts, settings) VALUES
('cut_to_length', 'Cut to Length', 'Track cut lengths and offcuts', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('contract_pricing', 'Contract Pricing', 'Customer and tier pricing contracts', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('delivery', 'Delivery Dispatch', 'Quotes, deliveries and dispatch tracking', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('isbn_catalog', 'ISBN Catalog', 'Validate and manage ISBN catalog records', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('consignment', 'Consignment Stock', 'Track supplier-owned book stock and settlements', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('publisher_returns', 'Publisher Returns', 'Return consignment stock to publishers', '["consignment"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('special_orders', 'Special Orders', 'Book special orders and notifications', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('second_hand', 'Second Hand', 'Track used book condition and pricing', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('school_orders', 'School Orders', 'Manage school bulk orders', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('rate_plans', 'Rate Plans', 'Seasonal hospitality rates', '["room_booking"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('housekeeping', 'Housekeeping', 'Room readiness and housekeeping board', '["room_booking"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('night_audit', 'Night Audit', 'Idempotent daily hospitality audit', '["guest_folio"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('group_billing', 'Group Billing', 'Corporate and group folios', '["guest_folio"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('warranty', 'Warranty', 'Warranty lookup and claims', '["serial_tracking"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('trade_ins', 'Trade-ins', 'Trade-in valuation and inventory intake', '["serial_tracking"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('installments', 'Installments', 'Payment plans and late fees', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('document_packs', 'Document Packs', 'Dealership document packs', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('appointments', 'Appointments', 'Conflict-safe service appointments', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('commissions', 'Commissions', 'Staff commission calculations', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('packages', 'Packages', 'Salon packages and redemption', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('job_cards', 'Job Cards', 'Garage work orders', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('parts_and_labor', 'Parts and Labor', 'Separate garage parts and labor lines', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('service_reminders', 'Service Reminders', 'Vehicle service reminders', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('pump_integration', 'Pump Integration', 'Fuel pump readings and sales', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('tank_dips', 'Tank Dips', 'Fuel tank dip readings', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('meter_shifts', 'Meter Shifts', 'Fuel meter shift reconciliation', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('pump_reconciliation', 'Pump Reconciliation', 'Pump variance reconciliation', '["pump_integration"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('recurring_billing', 'Recurring Billing', 'Recurring membership billing', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('membership_plans', 'Membership Plans', 'Membership plan catalog', '[]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('attendance', 'Attendance', 'Membership attendance tracking', '["membership_plans"]'::jsonb, '[]'::jsonb, '[]'::jsonb),
('idempotent_billing', 'Idempotent Billing', 'Prevent duplicate billing runs', '["recurring_billing"]'::jsonb, '[]'::jsonb, '[]'::jsonb)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE vertical_modules ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS vertical_modules_tenant_isolation ON vertical_modules;
CREATE POLICY vertical_modules_tenant_isolation ON vertical_modules
    USING (tenant_id::text = current_setting('app.tenant_id', true));

CREATE TABLE IF NOT EXISTS hardware_uom_conversions (
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    item_id uuid NOT NULL REFERENCES catalog_items(id) ON DELETE CASCADE,
    from_uom text NOT NULL,
    to_uom text NOT NULL,
    multiplier numeric(20,8) NOT NULL CHECK (multiplier > 0),
    PRIMARY KEY (tenant_id, item_id, from_uom, to_uom)
);

CREATE TABLE IF NOT EXISTS hardware_quotes (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id uuid,
    status text NOT NULL CHECK (status IN ('draft','accepted','invoiced','cancelled')),
    lines jsonb NOT NULL DEFAULT '[]'::jsonb,
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bookshop_books (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    isbn13 text,
    isbn10 text,
    title text NOT NULL,
    authors jsonb NOT NULL DEFAULT '[]'::jsonb,
    publisher text,
    consignment boolean NOT NULL DEFAULT false,
    supplier_terms numeric(8,5),
    UNIQUE (tenant_id, isbn13),
    UNIQUE (tenant_id, isbn10)
);

CREATE TABLE IF NOT EXISTS hospitality_rooms (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_number text NOT NULL,
    room_type text NOT NULL,
    active boolean NOT NULL DEFAULT true,
    UNIQUE (tenant_id, room_number)
);

CREATE TABLE IF NOT EXISTS hospitality_reservations (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_id uuid NOT NULL REFERENCES hospitality_rooms(id) ON DELETE RESTRICT,
    guest_id uuid,
    check_in date NOT NULL,
    check_out date NOT NULL,
    status text NOT NULL CHECK (status IN ('hold','confirmed','checked_in','checked_out','cancelled','no_show')),
    deposit_minor bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (check_out > check_in)
);

CREATE INDEX IF NOT EXISTS hospitality_reservations_dates
    ON hospitality_reservations (tenant_id, room_id, check_in, check_out)
    WHERE status IN ('hold','confirmed','checked_in');

CREATE TABLE IF NOT EXISTS dealership_assets (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    serial text NOT NULL,
    asset_type text NOT NULL,
    warranty_until date,
    status text NOT NULL DEFAULT 'available'
        CHECK (status IN ('available','reserved','sold','returned','repossessed')),
    UNIQUE (tenant_id, serial)
);

CREATE TABLE IF NOT EXISTS service_appointments (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_type text NOT NULL,
    staff_id uuid,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'booked'
        CHECK (status IN ('booked','completed','cancelled','no_show')),
    CHECK (ends_at > starts_at)
);

CREATE EXTENSION IF NOT EXISTS btree_gist;
ALTER TABLE hospitality_reservations
    DROP CONSTRAINT IF EXISTS hospitality_reservations_no_overlap;
ALTER TABLE hospitality_reservations
    ADD CONSTRAINT hospitality_reservations_no_overlap
    EXCLUDE USING gist (
        tenant_id WITH =,
        room_id WITH =,
        daterange(check_in, check_out, '[)') WITH &&
    ) WHERE (status IN ('hold','confirmed','checked_in'));

ALTER TABLE hardware_uom_conversions ENABLE ROW LEVEL SECURITY;
ALTER TABLE hardware_quotes ENABLE ROW LEVEL SECURITY;
ALTER TABLE bookshop_books ENABLE ROW LEVEL SECURITY;
ALTER TABLE hospitality_rooms ENABLE ROW LEVEL SECURITY;
ALTER TABLE hospitality_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE dealership_assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE service_appointments ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS hardware_uom_tenant_isolation ON hardware_uom_conversions;
CREATE POLICY hardware_uom_tenant_isolation ON hardware_uom_conversions USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS hardware_quotes_tenant_isolation ON hardware_quotes;
CREATE POLICY hardware_quotes_tenant_isolation ON hardware_quotes USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS bookshop_books_tenant_isolation ON bookshop_books;
CREATE POLICY bookshop_books_tenant_isolation ON bookshop_books USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS hospitality_rooms_tenant_isolation ON hospitality_rooms;
CREATE POLICY hospitality_rooms_tenant_isolation ON hospitality_rooms USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS hospitality_reservations_tenant_isolation ON hospitality_reservations;
CREATE POLICY hospitality_reservations_tenant_isolation ON hospitality_reservations USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS dealership_assets_tenant_isolation ON dealership_assets;
CREATE POLICY dealership_assets_tenant_isolation ON dealership_assets USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS service_appointments_tenant_isolation ON service_appointments;
CREATE POLICY service_appointments_tenant_isolation ON service_appointments USING (tenant_id::text = current_setting('app.tenant_id', true));
