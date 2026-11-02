CREATE TABLE IF NOT EXISTS branch_inventory (
  tenant_id uuid NOT NULL,
  branch_id uuid NOT NULL,
  item_id uuid NOT NULL,
  quantity_on_hand bigint NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id,branch_id,item_id),
  FOREIGN KEY (tenant_id,branch_id) REFERENCES branches(tenant_id,id),
  FOREIGN KEY (tenant_id,item_id) REFERENCES catalog_items(tenant_id,id)
);
ALTER TABLE stock_movements ADD COLUMN IF NOT EXISTS branch_id uuid;
DO $$ BEGIN
  ALTER TABLE stock_movements ADD CONSTRAINT stock_movements_branch_fk FOREIGN KEY (tenant_id,branch_id) REFERENCES branches(tenant_id,id);
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
UPDATE stock_movements sm SET branch_id=COALESCE(
  (SELECT s.branch_id FROM sales s WHERE s.tenant_id=sm.tenant_id AND s.id=sm.reference_id),
  (SELECT gr.branch_id FROM goods_receipts gr WHERE gr.tenant_id=sm.tenant_id AND gr.id=sm.reference_id),
  (SELECT sr.branch_id FROM supplier_returns sr WHERE sr.tenant_id=sm.tenant_id AND sr.id=sm.reference_id),
  (SELECT st.from_branch_id FROM stock_transfers st WHERE st.tenant_id=sm.tenant_id AND st.id=sm.reference_id),
  (SELECT b.id FROM branches b WHERE b.tenant_id=sm.tenant_id ORDER BY b.created_at,b.id LIMIT 1)
) WHERE sm.branch_id IS NULL;
INSERT INTO branch_inventory(tenant_id,branch_id,item_id,quantity_on_hand)
SELECT ci.tenant_id,b.id,ci.id,ci.stock_on_hand
FROM catalog_items ci
CROSS JOIN LATERAL (SELECT id FROM branches WHERE tenant_id=ci.tenant_id ORDER BY created_at,id LIMIT 1) b
ON CONFLICT (tenant_id,branch_id,item_id) DO NOTHING;
DO $$ DECLARE t text; BEGIN
  t := 'branch_inventory';
  EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
  EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
  EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
END $$;
CREATE INDEX IF NOT EXISTS branch_inventory_item_idx ON branch_inventory(tenant_id,item_id,branch_id);
