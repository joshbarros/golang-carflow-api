-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation_policy ON users;
DROP POLICY IF EXISTS tenant_isolation_policy ON cars;

-- Disable Row Level Security
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE cars DISABLE ROW LEVEL SECURITY;

-- Drop tenant_id column from cars table
DROP INDEX IF EXISTS idx_cars_tenant_id;
ALTER TABLE cars DROP COLUMN IF EXISTS tenant_id;

-- Drop tenant_id column from users table
DROP INDEX IF EXISTS idx_users_tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;

-- Drop tenants table and related objects
DROP INDEX IF EXISTS idx_tenants_custom_domain;
DROP INDEX IF EXISTS idx_tenants_status;
DROP TABLE IF EXISTS tenants;

-- Drop current_tenant_id function
DROP FUNCTION IF EXISTS current_tenant_id(); 