-- Drop triggers
DROP TRIGGER IF EXISTS enforce_tenant_isolation_users ON users;
DROP TRIGGER IF EXISTS enforce_tenant_isolation_cars ON cars;

-- Drop functions
DROP FUNCTION IF EXISTS app.enforce_tenant_isolation();
DROP FUNCTION IF EXISTS app.current_tenant_id();

-- Drop schema (only if empty)
DROP SCHEMA IF EXISTS app; 