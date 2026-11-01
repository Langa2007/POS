ALTER TABLE payment_refunds ADD COLUMN IF NOT EXISTS requested_by uuid;
ALTER TABLE payment_refunds ADD COLUMN IF NOT EXISTS approved_by uuid;
ALTER TABLE payment_refunds ADD COLUMN IF NOT EXISTS approved_at timestamptz;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='payment_refunds_requested_by_fk') THEN
    ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_requested_by_fk
      FOREIGN KEY (tenant_id, requested_by) REFERENCES users(tenant_id, id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='payment_refunds_approved_by_fk') THEN
    ALTER TABLE payment_refunds ADD CONSTRAINT payment_refunds_approved_by_fk
      FOREIGN KEY (tenant_id, approved_by) REFERENCES users(tenant_id, id);
  END IF;
END $$;

INSERT INTO permissions(id,description) VALUES
  ('payments.refund.approve','Approve refunds requested by another user'),
  ('payments.reconcile','Import and review provider statements')
ON CONFLICT(id) DO NOTHING;

DO $$ DECLARE tenant_row record; BEGIN
  FOR tenant_row IN SELECT id FROM tenants LOOP
    PERFORM set_config('app.tenant_id',tenant_row.id::text,true);
    INSERT INTO role_permissions(tenant_id,role_id,permission_id)
    SELECT r.tenant_id,r.id,p.permission_id
    FROM roles r CROSS JOIN (VALUES ('payments.refund.approve'),('payments.reconcile')) p(permission_id)
    WHERE lower(r.name) IN ('owner','admin')
    ON CONFLICT DO NOTHING;
  END LOOP;
END $$;
