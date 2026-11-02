CREATE TABLE IF NOT EXISTS tax_rates (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  code text NOT NULL,
  name text NOT NULL,
  basis_points bigint NOT NULL CHECK (basis_points BETWEEN 0 AND 10000),
  exempt boolean NOT NULL DEFAULT false,
  effective_from date NOT NULL,
  effective_to date,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, code, effective_from),
  CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE TABLE IF NOT EXISTS suppliers (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  name text NOT NULL,
  tax_number text,
  phone text,
  email text,
  payment_terms_days integer NOT NULL DEFAULT 0 CHECK (payment_terms_days >= 0),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id)
);

CREATE TABLE IF NOT EXISTS purchase_orders (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  supplier_id uuid NOT NULL,
  branch_id uuid NOT NULL,
  order_number text NOT NULL,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','partially_received','received','cancelled')),
  currency char(3) NOT NULL DEFAULT 'KES',
  expected_at date,
  total_minor bigint NOT NULL DEFAULT 0 CHECK (total_minor >= 0),
  created_by uuid NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, order_number),
  FOREIGN KEY (tenant_id, supplier_id) REFERENCES suppliers(tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, created_by) REFERENCES users(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS purchase_order_lines (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  purchase_order_id uuid NOT NULL,
  item_id uuid NOT NULL,
  ordered_quantity bigint NOT NULL CHECK (ordered_quantity > 0),
  received_quantity bigint NOT NULL DEFAULT 0 CHECK (received_quantity >= 0 AND received_quantity <= ordered_quantity),
  unit_cost_minor bigint NOT NULL CHECK (unit_cost_minor >= 0),
  UNIQUE (tenant_id, id),
  FOREIGN KEY (tenant_id, purchase_order_id) REFERENCES purchase_orders(tenant_id, id),
  FOREIGN KEY (tenant_id, item_id) REFERENCES catalog_items(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS goods_receipts (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  purchase_order_id uuid NOT NULL,
  branch_id uuid NOT NULL,
  receipt_number text NOT NULL,
  status text NOT NULL DEFAULT 'posted' CHECK (status IN ('draft','posted','cancelled')),
  received_at timestamptz NOT NULL DEFAULT now(),
  received_by uuid NOT NULL,
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, receipt_number),
  FOREIGN KEY (tenant_id, purchase_order_id) REFERENCES purchase_orders(tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, received_by) REFERENCES users(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS stock_takes (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  branch_id uuid NOT NULL,
  status text NOT NULL DEFAULT 'counting' CHECK (status IN ('counting','pending_approval','posted','cancelled')),
  blind boolean NOT NULL DEFAULT true,
  started_at timestamptz NOT NULL DEFAULT now(),
  approved_by uuid,
  approved_at timestamptz,
  UNIQUE (tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, approved_by) REFERENCES users(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS stock_take_lines (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  stock_take_id uuid NOT NULL,
  item_id uuid NOT NULL,
  expected_quantity bigint NOT NULL,
  counted_quantity bigint NOT NULL CHECK (counted_quantity >= 0),
  variance_quantity bigint GENERATED ALWAYS AS (counted_quantity - expected_quantity) STORED,
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, stock_take_id, item_id),
  FOREIGN KEY (tenant_id, stock_take_id) REFERENCES stock_takes(tenant_id, id),
  FOREIGN KEY (tenant_id, item_id) REFERENCES catalog_items(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS stock_transfers (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  from_branch_id uuid NOT NULL,
  to_branch_id uuid NOT NULL,
  status text NOT NULL DEFAULT 'in_transit' CHECK (status IN ('in_transit','received','cancelled')),
  shipped_at timestamptz,
  received_at timestamptz,
  created_by uuid NOT NULL,
  UNIQUE (tenant_id, id),
  CHECK (from_branch_id <> to_branch_id),
  FOREIGN KEY (tenant_id, from_branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, to_branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, created_by) REFERENCES users(tenant_id, id)
);

CREATE TABLE IF NOT EXISTS fiscal_submissions (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  document_type text NOT NULL CHECK (document_type IN ('receipt','invoice','credit_note')),
  document_id uuid NOT NULL,
  idempotency_key text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','submitted','failed')),
  control_number text,
  qr_code text,
  signature text,
  last_error text,
  attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  submitted_at timestamptz,
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, idempotency_key),
  UNIQUE (tenant_id, control_number)
);

DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['tax_rates','suppliers','purchase_orders','purchase_order_lines','goods_receipts','stock_takes','stock_take_lines','stock_transfers','fiscal_submissions'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;
