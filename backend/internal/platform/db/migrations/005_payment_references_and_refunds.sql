-- Additive follow-up: 004 may already be recorded in existing installations.
ALTER TABLE sales ADD COLUMN IF NOT EXISTS payment_reference text;
CREATE UNIQUE INDEX IF NOT EXISTS sales_payment_reference_idx
  ON sales(tenant_id,payment_reference) WHERE payment_reference IS NOT NULL;

ALTER TABLE payment_refunds ADD COLUMN IF NOT EXISTS idempotency_key text;
ALTER TABLE payment_refunds ALTER COLUMN status SET DEFAULT 'pending';
CREATE UNIQUE INDEX IF NOT EXISTS payment_refunds_request_key_idx
  ON payment_refunds(tenant_id,idempotency_key) WHERE idempotency_key IS NOT NULL;
