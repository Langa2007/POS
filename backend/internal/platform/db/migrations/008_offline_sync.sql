CREATE TABLE IF NOT EXISTS sync_operations (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  client_id text NOT NULL,
  terminal_id uuid,
  entity_type text NOT NULL,
  entity_id text NOT NULL,
  operation text NOT NULL CHECK (operation IN ('create','update','delete')),
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  client_occurred_at timestamptz NOT NULL,
  server_received_at timestamptz NOT NULL DEFAULT now(),
  status text NOT NULL CHECK (status IN ('accepted','rejected')),
  error_code text,
  error_message text,
  UNIQUE (tenant_id, client_id),
  FOREIGN KEY (tenant_id, terminal_id) REFERENCES terminals(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS sync_changes (
  cursor bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  entity_type text NOT NULL,
  entity_id text NOT NULL,
  operation text NOT NULL CHECK (operation IN ('create','update','delete')),
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, cursor)
);

CREATE INDEX IF NOT EXISTS sync_changes_tenant_cursor_idx
  ON sync_changes (tenant_id, cursor);
CREATE INDEX IF NOT EXISTS sync_operations_terminal_idx
  ON sync_operations (tenant_id, terminal_id, server_received_at);

ALTER TABLE sync_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_operations FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON sync_operations;
CREATE POLICY tenant_isolation ON sync_operations
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);

ALTER TABLE sync_changes ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_changes FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON sync_changes;
CREATE POLICY tenant_isolation ON sync_changes
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);
