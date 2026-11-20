CREATE TABLE IF NOT EXISTS vertical_operation_records (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    vertical_id text NOT NULL CHECK (vertical_id IN ('fuelstation', 'garage')),
    resource text NOT NULL CHECK (resource IN ('pumps', 'shifts', 'dips', 'deliveries', 'vehicles', 'jobs', 'appointments', 'reminders')),
    data jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((vertical_id = 'fuelstation' AND resource IN ('pumps', 'shifts', 'dips', 'deliveries')) OR
           (vertical_id = 'garage' AND resource IN ('vehicles', 'jobs', 'appointments', 'reminders')))
);

CREATE INDEX IF NOT EXISTS vertical_operation_records_lookup
    ON vertical_operation_records (tenant_id, vertical_id, resource, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS garage_vehicle_registration_unique
    ON vertical_operation_records (tenant_id, upper(data->>'registration'))
    WHERE vertical_id='garage' AND resource='vehicles';

ALTER TABLE vertical_operation_records
    DROP CONSTRAINT IF EXISTS garage_technician_appointment_no_overlap;
ALTER TABLE vertical_operation_records
    ADD CONSTRAINT garage_technician_appointment_no_overlap
    EXCLUDE USING gist (
        tenant_id WITH =,
        (data->>'technician') WITH =,
        tstzrange((data->>'start')::timestamptz, (data->>'end')::timestamptz, '[)') WITH &&
    ) WHERE (
        vertical_id='garage' AND resource='appointments'
        AND COALESCE(data->>'technician', '') <> ''
        AND COALESCE(data->>'status', 'Booked') NOT IN ('Cancelled', 'Completed', 'No show')
    );

ALTER TABLE vertical_operation_records ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS vertical_operation_records_tenant_isolation ON vertical_operation_records;
CREATE POLICY vertical_operation_records_tenant_isolation ON vertical_operation_records
    USING (tenant_id::text = current_setting('app.tenant_id', true))
    WITH CHECK (tenant_id::text = current_setting('app.tenant_id', true));

INSERT INTO permissions(id, description) VALUES
    ('fuelstation.read', 'View fuel station operations'),
    ('fuelstation.manage', 'Manage fuel station records'),
    ('fuelstation.reconcile', 'Reconcile fuel meter shifts'),
    ('garage.read', 'View garage workshop records'),
    ('garage.manage', 'Manage garage workshop records')
ON CONFLICT (id) DO NOTHING;

INSERT INTO role_permissions(tenant_id, role_id, permission_id)
SELECT r.tenant_id, r.id, p.id
FROM roles r CROSS JOIN permissions p
WHERE lower(r.name) IN ('owner', 'admin')
  AND p.id IN ('fuelstation.read', 'fuelstation.manage', 'fuelstation.reconcile', 'garage.read', 'garage.manage')
ON CONFLICT DO NOTHING;
