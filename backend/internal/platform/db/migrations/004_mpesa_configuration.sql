CREATE TABLE IF NOT EXISTS mpesa_configurations (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
  environment text NOT NULL CHECK (environment IN ('sandbox','production')),
  business_shortcode text NOT NULL,
  encrypted_config bytea NOT NULL,
  callback_token_hash bytea NOT NULL,
  enabled boolean NOT NULL DEFAULT false,
  c2b_registered_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE mpesa_configurations ENABLE ROW LEVEL SECURITY;
ALTER TABLE mpesa_configurations FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON mpesa_configurations;
CREATE POLICY tenant_isolation ON mpesa_configurations
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

INSERT INTO permissions(id,description) VALUES ('payments.manage','Configure payment providers') ON CONFLICT(id) DO NOTHING;
DO $$ DECLARE tenant_row record; BEGIN
  FOR tenant_row IN SELECT id FROM tenants LOOP
    PERFORM set_config('app.tenant_id',tenant_row.id::text,true);
    INSERT INTO role_permissions(tenant_id,role_id,permission_id)
    SELECT r.tenant_id,r.id,'payments.manage' FROM roles r WHERE lower(r.name) IN ('owner','admin')
    ON CONFLICT DO NOTHING;
  END LOOP;
END $$;

ALTER TABLE payment_intents ADD COLUMN IF NOT EXISTS merchant_request_id text;
ALTER TABLE payment_intents ADD COLUMN IF NOT EXISTS callback_received_at timestamptz;
ALTER TABLE payment_intents ADD COLUMN IF NOT EXISTS captured_at timestamptz;
ALTER TABLE payment_intents ADD COLUMN IF NOT EXISTS raw_provider_response jsonb;
ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_status_check;
ALTER TABLE sales ADD CONSTRAINT sales_status_check CHECK (status IN ('pending_payment','completed','voided','returned'));
ALTER TABLE sales ADD COLUMN IF NOT EXISTS payment_reference text;
CREATE UNIQUE INDEX IF NOT EXISTS sales_payment_reference_idx ON sales(tenant_id,payment_reference) WHERE payment_reference IS NOT NULL;
ALTER TABLE payment_refunds ADD COLUMN IF NOT EXISTS idempotency_key text;
ALTER TABLE payment_refunds ALTER COLUMN status SET DEFAULT 'pending';
CREATE UNIQUE INDEX IF NOT EXISTS payment_refunds_request_key_idx
  ON payment_refunds(tenant_id,idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS payment_intents_checkout_id_idx
  ON payment_intents(tenant_id, provider_reference) WHERE provider_reference IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ledger_journals_reference_idx
  ON ledger_journals(tenant_id, reference_type, reference_id);
