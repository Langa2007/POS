CREATE TABLE IF NOT EXISTS salon_service_packages (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    visits integer NOT NULL CHECK (visits > 0),
    duration_minutes integer NOT NULL CHECK (duration_minutes > 0),
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS salon_commission_periods (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    staff_id uuid NOT NULL,
    period_start date NOT NULL,
    period_end date NOT NULL,
    service_count integer NOT NULL DEFAULT 0,
    sales_minor bigint NOT NULL DEFAULT 0 CHECK (sales_minor >= 0),
    commission_minor bigint NOT NULL DEFAULT 0 CHECK (commission_minor >= 0),
    UNIQUE (tenant_id, staff_id, period_start, period_end)
);

ALTER TABLE salon_service_packages ENABLE ROW LEVEL SECURITY;
ALTER TABLE salon_commission_periods ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS salon_service_packages_tenant_isolation ON salon_service_packages;
CREATE POLICY salon_service_packages_tenant_isolation ON salon_service_packages
    USING (tenant_id::text = current_setting('app.tenant_id', true));
DROP POLICY IF EXISTS salon_commission_periods_tenant_isolation ON salon_commission_periods;
CREATE POLICY salon_commission_periods_tenant_isolation ON salon_commission_periods
    USING (tenant_id::text = current_setting('app.tenant_id', true));
