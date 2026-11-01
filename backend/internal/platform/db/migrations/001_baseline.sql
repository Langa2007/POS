CREATE TABLE IF NOT EXISTS tenants (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS businesses (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id)
);
CREATE TABLE IF NOT EXISTS branches (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  business_id uuid NOT NULL,
  name text NOT NULL,
  timezone text NOT NULL DEFAULT 'Africa/Nairobi',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  FOREIGN KEY (tenant_id, business_id) REFERENCES businesses(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS terminals (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  branch_id uuid NOT NULL,
  name text NOT NULL,
  token_hash text UNIQUE,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  email text NOT NULL,
  password_hash text NOT NULL,
  pin_hash text,
  pin_failed_count integer NOT NULL DEFAULT 0,
  pin_locked_until timestamptz,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, email)
);
CREATE TABLE IF NOT EXISTS roles (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  name text NOT NULL,
  UNIQUE (tenant_id, id),
  UNIQUE (tenant_id, name)
);
CREATE TABLE IF NOT EXISTS permissions (
  id text PRIMARY KEY,
  description text NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS user_roles (
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  user_id uuid NOT NULL,
  role_id uuid NOT NULL,
  PRIMARY KEY (tenant_id, user_id, role_id),
  FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id),
  FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS user_branches (
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  user_id uuid NOT NULL,
  branch_id uuid NOT NULL,
  PRIMARY KEY (tenant_id, user_id, branch_id),
  FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS role_permissions (
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  role_id uuid NOT NULL,
  permission_id text NOT NULL REFERENCES permissions(id),
  PRIMARY KEY (tenant_id, role_id, permission_id),
  FOREIGN KEY (tenant_id, role_id) REFERENCES roles(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  user_id uuid NOT NULL,
  branch_id uuid,
  terminal_id uuid,
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, user_id) REFERENCES users(tenant_id, id),
  FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
  FOREIGN KEY (tenant_id, terminal_id) REFERENCES terminals(tenant_id, id),
  CHECK ((branch_id IS NULL) = (terminal_id IS NULL))
);
CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx ON refresh_tokens (tenant_id,user_id);
CREATE TABLE IF NOT EXISTS tenant_modules (
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  module_id text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  enabled_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (tenant_id, module_id)
);
CREATE TABLE IF NOT EXISTS tenant_capabilities (
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  capability_id text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  PRIMARY KEY (tenant_id, capability_id)
);
CREATE TABLE IF NOT EXISTS audit_log (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  actor_id uuid,
  action text NOT NULL,
  entity_type text NOT NULL,
  entity_id text NOT NULL,
  before_data jsonb,
  after_data jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  FOREIGN KEY (tenant_id, actor_id) REFERENCES users(tenant_id, id)
);
CREATE TABLE IF NOT EXISTS outbox (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL REFERENCES tenants(id),
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  delivered_at timestamptz,
  attempts integer NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS outbox_pending_idx ON outbox (occurred_at) WHERE delivered_at IS NULL;

DO $$ DECLARE t text; BEGIN
  FOREACH t IN ARRAY ARRAY['businesses','branches','terminals','users','roles','user_roles','user_branches','role_permissions','refresh_tokens','tenant_modules','tenant_capabilities','audit_log','outbox'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %I', t);
    EXECUTE format('CREATE POLICY tenant_isolation ON %I USING (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = NULLIF(current_setting(''app.tenant_id'', true), '''')::uuid)', t);
  END LOOP;
END $$;
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON tenants;
CREATE POLICY tenant_isolation ON tenants USING (id = NULLIF(current_setting('app.tenant_id', true), '')::uuid) WITH CHECK (id = NULLIF(current_setting('app.tenant_id', true), '')::uuid);
