ALTER TABLE tax_rates ADD COLUMN IF NOT EXISTS category_id uuid;
DO $$ BEGIN
  ALTER TABLE tax_rates ADD CONSTRAINT tax_rates_category_fk FOREIGN KEY (tenant_id, category_id) REFERENCES catalog_categories(tenant_id, id);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS goods_receipt_lines (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, goods_receipt_id uuid NOT NULL,
  purchase_order_line_id uuid NOT NULL, item_id uuid NOT NULL,
  quantity bigint NOT NULL CHECK (quantity > 0), unit_cost_minor bigint NOT NULL CHECK (unit_cost_minor >= 0),
  UNIQUE (tenant_id,id),
  FOREIGN KEY (tenant_id,goods_receipt_id) REFERENCES goods_receipts(tenant_id,id),
  FOREIGN KEY (tenant_id,purchase_order_line_id) REFERENCES purchase_order_lines(tenant_id,id),
  FOREIGN KEY (tenant_id,item_id) REFERENCES catalog_items(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS stock_transfer_lines (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, transfer_id uuid NOT NULL, item_id uuid NOT NULL,
  shipped_quantity bigint NOT NULL CHECK (shipped_quantity > 0), received_quantity bigint CHECK (received_quantity >= 0 AND received_quantity <= shipped_quantity),
  UNIQUE (tenant_id,id), UNIQUE (tenant_id,transfer_id,item_id),
  FOREIGN KEY (tenant_id,transfer_id) REFERENCES stock_transfers(tenant_id,id),
  FOREIGN KEY (tenant_id,item_id) REFERENCES catalog_items(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS supplier_invoices (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, supplier_id uuid NOT NULL, goods_receipt_id uuid,
  invoice_number text NOT NULL, currency char(3) NOT NULL DEFAULT 'KES', total_minor bigint NOT NULL CHECK (total_minor >= 0),
  due_at date, status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','partially_paid','paid','void')),
  created_at timestamptz NOT NULL DEFAULT now(), UNIQUE (tenant_id,id), UNIQUE (tenant_id,supplier_id,invoice_number),
  FOREIGN KEY (tenant_id,supplier_id) REFERENCES suppliers(tenant_id,id),
  FOREIGN KEY (tenant_id,goods_receipt_id) REFERENCES goods_receipts(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS supplier_payments (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, invoice_id uuid NOT NULL, idempotency_key text NOT NULL,
  amount_minor bigint NOT NULL CHECK (amount_minor > 0), method text NOT NULL, reference text,
  paid_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  UNIQUE (tenant_id,id), UNIQUE (tenant_id,idempotency_key),
  FOREIGN KEY (tenant_id,invoice_id) REFERENCES supplier_invoices(tenant_id,id),
  FOREIGN KEY (tenant_id,created_by) REFERENCES users(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS supplier_returns (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, supplier_id uuid NOT NULL, branch_id uuid NOT NULL,
  return_number text NOT NULL, status text NOT NULL DEFAULT 'posted' CHECK (status IN ('draft','posted','cancelled')),
  total_minor bigint NOT NULL CHECK (total_minor >= 0), created_by uuid NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id,id), UNIQUE (tenant_id,return_number),
  FOREIGN KEY (tenant_id,supplier_id) REFERENCES suppliers(tenant_id,id),
  FOREIGN KEY (tenant_id,branch_id) REFERENCES branches(tenant_id,id),
  FOREIGN KEY (tenant_id,created_by) REFERENCES users(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS supplier_return_lines (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, return_id uuid NOT NULL, item_id uuid NOT NULL,
  quantity bigint NOT NULL CHECK (quantity > 0), unit_cost_minor bigint NOT NULL CHECK (unit_cost_minor >= 0),
  UNIQUE (tenant_id,id), FOREIGN KEY (tenant_id,return_id) REFERENCES supplier_returns(tenant_id,id),
  FOREIGN KEY (tenant_id,item_id) REFERENCES catalog_items(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS inventory_cost_layers (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, item_id uuid NOT NULL, branch_id uuid NOT NULL,
  source_type text NOT NULL, source_id uuid NOT NULL, quantity_remaining bigint NOT NULL CHECK (quantity_remaining >= 0),
  unit_cost_minor bigint NOT NULL CHECK (unit_cost_minor >= 0), created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id,id), FOREIGN KEY (tenant_id,item_id) REFERENCES catalog_items(tenant_id,id),
  FOREIGN KEY (tenant_id,branch_id) REFERENCES branches(tenant_id,id)
);
CREATE TABLE IF NOT EXISTS sale_tax_lines (
  id uuid PRIMARY KEY, tenant_id uuid NOT NULL, sale_id uuid NOT NULL, tax_rate_id uuid,
  tax_code text NOT NULL, net_minor bigint NOT NULL CHECK (net_minor >= 0), tax_minor bigint NOT NULL CHECK (tax_minor >= 0),
  gross_minor bigint NOT NULL CHECK (gross_minor = net_minor + tax_minor),
  FOREIGN KEY (tenant_id,sale_id) REFERENCES sales(tenant_id,id),
  FOREIGN KEY (tenant_id,tax_rate_id) REFERENCES tax_rates(tenant_id,id)
);
CREATE INDEX IF NOT EXISTS supplier_invoices_due_idx ON supplier_invoices(tenant_id,due_at) WHERE status IN ('open','partially_paid');
CREATE INDEX IF NOT EXISTS inventory_cost_layers_fifo_idx ON inventory_cost_layers(tenant_id,branch_id,item_id,created_at,id) WHERE quantity_remaining > 0;
ALTER TABLE fiscal_submissions ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
CREATE INDEX IF NOT EXISTS fiscal_pending_idx ON fiscal_submissions(tenant_id,created_at) WHERE status IN ('pending','failed');
DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['goods_receipt_lines','stock_transfer_lines','supplier_invoices','supplier_payments','supplier_returns','supplier_return_lines','inventory_cost_layers','sale_tax_lines'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;

INSERT INTO permissions(id,description) VALUES
 ('tax.read','View tax configuration and calculations'),('tax.manage','Manage tax rates'),
 ('purchasing.read','View suppliers and purchase workflows'),('purchasing.manage','Create suppliers and purchase orders'),
 ('purchasing.approve','Approve purchase orders'),('purchasing.receive','Receive purchase orders'),('purchasing.return','Post supplier returns'),
 ('payables.read','View supplier invoices'),('payables.manage','Record supplier invoices'),('payables.pay','Record supplier payments'),
 ('inventory.read','View stock and transfers'),('inventory.count','Create stock counts'),('inventory.adjust','Approve stock adjustments'),('inventory.transfer','Ship and receive stock transfers'),
 ('fiscal.submit','Queue and retry fiscal documents') ON CONFLICT(id) DO NOTHING;
INSERT INTO role_permissions(tenant_id,role_id,permission_id)
 SELECT r.tenant_id,r.id,p.id FROM roles r CROSS JOIN permissions p
 WHERE r.name IN ('owner','admin') AND p.id IN ('tax.read','tax.manage','purchasing.read','purchasing.manage','purchasing.approve','purchasing.receive','purchasing.return','payables.read','payables.manage','payables.pay','inventory.read','inventory.count','inventory.adjust','inventory.transfer','fiscal.submit')
 ON CONFLICT DO NOTHING;
