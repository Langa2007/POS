CREATE TABLE IF NOT EXISTS payment_allocation_requests (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  idempotency_key text NOT NULL,
  total_minor bigint NOT NULL CHECK (total_minor > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, idempotency_key),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id)
);
ALTER TABLE payment_allocation_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_allocation_requests FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON payment_allocation_requests;
CREATE POLICY tenant_isolation ON payment_allocation_requests
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);
