DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'pos_app') THEN
    CREATE ROLE pos_app LOGIN PASSWORD 'pos-app-local-only';
  END IF;
END $$;
GRANT CONNECT, CREATE ON DATABASE pos TO pos_app;
GRANT USAGE, CREATE ON SCHEMA public TO pos_app;
