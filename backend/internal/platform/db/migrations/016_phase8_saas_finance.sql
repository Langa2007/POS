CREATE TABLE IF NOT EXISTS bank_statements (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    provider text NOT NULL,
    account_reference text NOT NULL,
    statement_date date NOT NULL,
    imported_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, provider, account_reference, statement_date)
);

CREATE TABLE IF NOT EXISTS bank_statement_lines (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    statement_id uuid NOT NULL,
    external_reference text,
    value_date date NOT NULL,
    amount_minor bigint NOT NULL,
    fee_minor bigint NOT NULL DEFAULT 0 CHECK (fee_minor >= 0),
    description text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'unmatched'
        CHECK (status IN ('unmatched','matched','exception')),
    UNIQUE (tenant_id, statement_id, id),
    FOREIGN KEY (tenant_id, statement_id) REFERENCES bank_statements(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS reconciliation_matches (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    statement_line_id uuid NOT NULL,
    target_type text NOT NULL CHECK (target_type IN ('sale','invoice','payment')),
    target_id uuid NOT NULL,
    matched_minor bigint NOT NULL CHECK (matched_minor > 0),
    confidence numeric(5,4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    automatic boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, statement_line_id, target_type, target_id),
    FOREIGN KEY (tenant_id, statement_line_id) REFERENCES bank_statement_lines(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS payouts (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    beneficiary_type text NOT NULL CHECK (beneficiary_type IN ('supplier','employee','customer')),
    beneficiary_id uuid,
    provider text NOT NULL,
    amount_minor bigint NOT NULL CHECK (amount_minor > 0),
    idempotency_key text NOT NULL,
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','approved','rejected','submitted','completed','failed')),
    requested_by uuid NOT NULL,
    approved_by uuid,
    provider_reference text,
    created_at timestamptz NOT NULL DEFAULT now(),
    approved_at timestamptz,
    UNIQUE (tenant_id, idempotency_key),
    CHECK (approved_by IS NULL OR approved_by <> requested_by)
);

CREATE TABLE IF NOT EXISTS saas_plans (
    id text PRIMARY KEY,
    name text NOT NULL,
    monthly_minor bigint NOT NULL CHECK (monthly_minor >= 0),
    capabilities jsonb NOT NULL DEFAULT '[]'::jsonb,
    included_units bigint NOT NULL DEFAULT 0 CHECK (included_units >= 0),
    overage_minor bigint NOT NULL DEFAULT 0 CHECK (overage_minor >= 0),
    active boolean NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS tenant_subscriptions (
    tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
    plan_id text NOT NULL REFERENCES saas_plans(id),
    status text NOT NULL DEFAULT 'active'
        CHECK (status IN ('trialing','active','past_due','suspended','cancelled')),
    period_start date NOT NULL,
    period_end date NOT NULL,
    grace_until date,
    CHECK (period_end > period_start)
);

CREATE TABLE IF NOT EXISTS tenant_usage (
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    period_start date NOT NULL,
    metric text NOT NULL,
    units bigint NOT NULL DEFAULT 0 CHECK (units >= 0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, period_start, metric)
);

CREATE TABLE IF NOT EXISTS tenant_branding (
    tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
    logo_url text,
    primary_color text NOT NULL DEFAULT '#1f2937',
    receipt_footer text NOT NULL DEFAULT '',
    custom_domain text UNIQUE,
    updated_at timestamptz NOT NULL DEFAULT now()
);

DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['bank_statements','bank_statement_lines','reconciliation_matches','payouts','tenant_subscriptions','tenant_usage','tenant_branding'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;

INSERT INTO saas_plans(id, name, monthly_minor, capabilities, included_units, overage_minor)
VALUES
    ('starter', 'Starter', 250000, '["catalog","sales","payments"]'::jsonb, 1000, 0),
    ('growth', 'Growth', 750000, '["catalog","sales","payments","multi_branch","analytics"]'::jsonb, 10000, 25),
    ('enterprise', 'Enterprise', 0, '["*"]'::jsonb, 1000000000, 0)
ON CONFLICT (id) DO NOTHING;
