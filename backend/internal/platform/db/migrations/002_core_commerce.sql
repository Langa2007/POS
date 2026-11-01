CREATE TABLE IF NOT EXISTS catalog_categories (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  name text NOT NULL,
  parent_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, name),
  FOREIGN KEY (tenant_id, parent_id) REFERENCES catalog_categories(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS catalog_items (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  category_id uuid,
  name text NOT NULL,
  sku text NOT NULL,
  barcode text,
  kind text NOT NULL DEFAULT 'stocked' CHECK (kind IN ('stocked','service','non_inventory')),
  price_minor bigint NOT NULL CHECK (price_minor >= 0),
  cost_minor bigint NOT NULL DEFAULT 0 CHECK (cost_minor >= 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  stock_on_hand bigint NOT NULL DEFAULT 0,
  reorder_level bigint NOT NULL DEFAULT 0 CHECK (reorder_level >= 0),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, sku),
  UNIQUE (tenant_id, barcode),
  FOREIGN KEY (tenant_id, category_id) REFERENCES catalog_categories(tenant_id, id)
);
CREATE INDEX IF NOT EXISTS catalog_items_search_idx ON catalog_items (tenant_id, lower(name), sku, barcode);
CREATE TABLE IF NOT EXISTS shifts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  branch_id uuid NOT NULL,
  user_id uuid NOT NULL,
  opening_float_minor bigint NOT NULL CHECK (opening_float_minor >= 0),
  closing_cash_minor bigint,
  variance_minor bigint,
  status text NOT NULL CHECK (status IN ('open','closed')),
  opened_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz,
  UNIQUE (tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id)
);
CREATE UNIQUE INDEX IF NOT EXISTS one_open_shift_per_user ON shifts (tenant_id, user_id) WHERE status = 'open';
CREATE TABLE IF NOT EXISTS sales (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  branch_id uuid NOT NULL,
  cashier_id uuid NOT NULL,
  shift_id uuid NOT NULL,
  customer_name text,
  total_minor bigint NOT NULL CHECK (total_minor >= 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  status text NOT NULL CHECK (status IN ('completed','voided','returned')),
  idempotency_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, idempotency_key),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, cashier_id) REFERENCES users(tenant_id, id),
  FOREIGN KEY (tenant_id, shift_id) REFERENCES shifts(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS sale_lines (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  item_id uuid NOT NULL,
  quantity bigint NOT NULL CHECK (quantity > 0),
  unit_price_minor bigint NOT NULL CHECK (unit_price_minor >= 0),
  discount_minor bigint NOT NULL DEFAULT 0 CHECK (discount_minor >= 0),
  line_total_minor bigint NOT NULL CHECK (line_total_minor >= 0),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id),
  FOREIGN KEY (tenant_id, item_id) REFERENCES catalog_items(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS stock_movements (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  item_id uuid NOT NULL,
  quantity_delta bigint NOT NULL,
  reason text NOT NULL,
  reference_id uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, item_id) REFERENCES catalog_items(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS cash_tenders (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  amount_minor bigint NOT NULL CHECK (amount_minor >= 0),
  received_minor bigint NOT NULL CHECK (received_minor >= amount_minor),
  change_minor bigint NOT NULL CHECK (change_minor >= 0),
  currency char(3) NOT NULL DEFAULT 'KES',
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS sale_returns (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  sale_id uuid NOT NULL,
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  reason text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, sale_id) REFERENCES sales(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS report_exports (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  report_type text NOT NULL,
  requested_by uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['catalog_categories','catalog_items','shifts','sales','sale_lines','stock_movements','cash_tenders','sale_returns','report_exports'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;
