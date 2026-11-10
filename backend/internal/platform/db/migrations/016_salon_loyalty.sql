CREATE TABLE IF NOT EXISTS salon_loyalty_members (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name text NOT NULL,
    phone text NOT NULL,
    membership_no text NOT NULL,
    tier text NOT NULL DEFAULT 'Bronze'
        CHECK (tier IN ('Bronze', 'Silver', 'Gold')),
    points integer NOT NULL DEFAULT 0 CHECK (points >= 0),
    last_visit timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, membership_no),
    UNIQUE (tenant_id, phone)
);

ALTER TABLE salon_loyalty_members ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS salon_loyalty_members_tenant_isolation ON salon_loyalty_members;
CREATE POLICY salon_loyalty_members_tenant_isolation
    ON salon_loyalty_members
    USING (tenant_id::text = current_setting('app.tenant_id', true));
