CREATE TABLE IF NOT EXISTS payment_intents (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid,
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  provider text NOT NULL,
  method text NOT NULL CHECK (method IN ('cash','mpesa_stk','mpesa_c2b','credit')),
  phone text,
  status text NOT NULL CHECK (status IN ('pending','authorized','captured','failed','cancelled','refunded')),
  provider_reference text,
  idempotency_key text NOT NULL,
  raw_payload jsonb,
  failure_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, idempotency_key),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id)
);
CREATE INDEX IF NOT EXISTS payment_intents_status_idx ON payment_intents(tenant_id,status,created_at);
CREATE TABLE IF NOT EXISTS payment_tenders (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  intent_id uuid,
  method text NOT NULL CHECK (method IN ('cash','mpesa_stk','mpesa_c2b','credit')),
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  provider_reference text,
  idempotency_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id),
  FOREIGN KEY (tenant_id, intent_id) REFERENCES payment_intents(tenant_id, id)
);
CREATE UNIQUE INDEX IF NOT EXISTS payment_tenders_idempotency_idx ON payment_tenders(tenant_id,idempotency_key);
CREATE TABLE IF NOT EXISTS payment_callbacks (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  provider text NOT NULL,
  callback_key text NOT NULL,
  payload jsonb NOT NULL,
  processed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, provider, callback_key)
);
CREATE TABLE IF NOT EXISTS provider_credentials (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  provider text NOT NULL,
  environment text NOT NULL CHECK (environment IN ('sandbox','production')),
  encrypted_secret bytea NOT NULL,
  key_version integer NOT NULL DEFAULT 1,
  rotated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, provider, environment)
);
CREATE TABLE IF NOT EXISTS unallocated_payments (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  provider text NOT NULL,
  provider_reference text NOT NULL,
  amount_minor bigint NOT NULL,
  currency char(3) NOT NULL DEFAULT 'KES',
  account_reference text,
  payload jsonb NOT NULL,
  status text NOT NULL DEFAULT 'unallocated' CHECK (status IN ('unallocated','allocated','rejected')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, provider, provider_reference)
);
CREATE TABLE IF NOT EXISTS reconciliation_records (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  provider text NOT NULL,
  provider_reference text NOT NULL,
  expected_minor bigint,
  actual_minor bigint,
  variance_minor bigint NOT NULL,
  status text NOT NULL CHECK (status IN ('matched','variance','unresolved')),
  statement_date date,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS ledger_accounts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  code text NOT NULL,
  name text NOT NULL,
  account_type text NOT NULL CHECK (account_type IN ('asset','liability','equity','revenue','expense')),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, code)
);
CREATE TABLE IF NOT EXISTS ledger_journals (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  reference_type text NOT NULL,
  reference_id text NOT NULL,
  description text NOT NULL,
  posted_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id)
);
CREATE TABLE IF NOT EXISTS ledger_lines (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  journal_id uuid NOT NULL,
  account_id uuid NOT NULL,
  debit_minor bigint NOT NULL DEFAULT 0 CHECK (debit_minor >= 0),
  credit_minor bigint NOT NULL DEFAULT 0 CHECK (credit_minor >= 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  CHECK ((debit_minor = 0) <> (credit_minor = 0)),
  FOREIGN KEY (tenant_id, journal_id) REFERENCES ledger_journals(tenant_id, id),
  FOREIGN KEY (tenant_id, account_id) REFERENCES ledger_accounts(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS credit_accounts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  customer_name text NOT NULL,
  limit_minor bigint NOT NULL CHECK (limit_minor >= 0),
  balance_minor bigint NOT NULL DEFAULT 0 CHECK (balance_minor >= 0),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id)
);
CREATE TABLE IF NOT EXISTS payment_refunds (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  tender_id uuid,
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  reason text NOT NULL,
  status text NOT NULL DEFAULT 'completed' CHECK (status IN ('pending','completed','rejected')),
  provider_reference text,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id),
  FOREIGN KEY (tenant_id, tender_id) REFERENCES payment_tenders(tenant_id, id)
);
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['payment_intents','payment_tenders','payment_callbacks','provider_credentials','unallocated_payments','reconciliation_records','ledger_accounts','ledger_journals','ledger_lines','credit_accounts','payment_refunds'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;
