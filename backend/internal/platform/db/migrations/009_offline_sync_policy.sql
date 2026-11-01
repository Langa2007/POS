ALTER TABLE sync_operations ADD COLUMN IF NOT EXISTS operation_hash bytea;
ALTER TABLE sales ADD COLUMN IF NOT EXISTS receipt_number text;
CREATE UNIQUE INDEX IF NOT EXISTS sales_offline_receipt_number_idx ON sales(tenant_id,receipt_number) WHERE receipt_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS offline_sync_policies (
  tenant_id uuid PRIMARY KEY REFERENCES tenants(id),
  max_offline_hours integer NOT NULL DEFAULT 24 CHECK (max_offline_hours BETWEEN 1 AND 168),
  max_sale_minor bigint NOT NULL DEFAULT 500000 CHECK (max_sale_minor > 0),
  allow_negative_stock boolean NOT NULL DEFAULT false,
  updated_at timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE offline_sync_policies ENABLE ROW LEVEL SECURITY;
ALTER TABLE offline_sync_policies FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON offline_sync_policies;
CREATE POLICY tenant_isolation ON offline_sync_policies
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

INSERT INTO permissions(id,description) VALUES
  ('sync.policy.read','Read offline sync policy'),
  ('sync.policy.manage','Configure offline sync policy')
ON CONFLICT(id) DO NOTHING;
DO $$ DECLARE tenant_row record; BEGIN
  FOR tenant_row IN SELECT id FROM tenants LOOP
    PERFORM set_config('app.tenant_id',tenant_row.id::text,true);
    INSERT INTO offline_sync_policies(tenant_id) VALUES(tenant_row.id) ON CONFLICT DO NOTHING;
    INSERT INTO role_permissions(tenant_id,role_id,permission_id)
    SELECT r.tenant_id,r.id,p.permission_id
    FROM roles r CROSS JOIN (VALUES ('sync.policy.read'),('sync.policy.manage')) p(permission_id)
    WHERE lower(r.name) IN ('owner','admin') ON CONFLICT DO NOTHING;
  END LOOP;
END $$;
