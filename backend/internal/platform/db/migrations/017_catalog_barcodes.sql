CREATE TABLE IF NOT EXISTS catalog_item_barcodes (
  tenant_id uuid NOT NULL,
  item_id uuid NOT NULL,
  barcode text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, barcode),
  FOREIGN KEY (tenant_id, item_id) REFERENCES catalog_items(tenant_id, id) ON DELETE CASCADE
);

INSERT INTO catalog_item_barcodes(tenant_id,item_id,barcode)
SELECT tenant_id,id,barcode
FROM catalog_items
WHERE barcode IS NOT NULL AND barcode <> ''
ON CONFLICT (tenant_id,barcode) DO NOTHING;

CREATE INDEX IF NOT EXISTS catalog_item_barcodes_item_idx
  ON catalog_item_barcodes(tenant_id,item_id);

ALTER TABLE catalog_item_barcodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE catalog_item_barcodes FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON catalog_item_barcodes;
CREATE POLICY tenant_isolation ON catalog_item_barcodes
  USING (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid)
  WITH CHECK (tenant_id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);
