-- Create a custom schema for application settings
CREATE SCHEMA IF NOT EXISTS app;

-- Create a function to get the current tenant ID
CREATE OR REPLACE FUNCTION app.current_tenant_id()
RETURNS TEXT AS $$
BEGIN
    RETURN current_setting('app.tenant_id', TRUE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create a function to enforce tenant isolation in queries
CREATE OR REPLACE FUNCTION app.enforce_tenant_isolation()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.tenant_id IS NULL OR NEW.tenant_id != app.current_tenant_id() THEN
        RAISE EXCEPTION 'Invalid tenant ID';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to enforce tenant isolation
CREATE TRIGGER enforce_tenant_isolation_users
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_tenant_isolation();

CREATE TRIGGER enforce_tenant_isolation_cars
    BEFORE INSERT OR UPDATE ON cars
    FOR EACH ROW
    EXECUTE FUNCTION app.enforce_tenant_isolation();

-- Grant necessary permissions
GRANT USAGE ON SCHEMA app TO public;
GRANT EXECUTE ON FUNCTION app.current_tenant_id TO public;
GRANT EXECUTE ON FUNCTION app.enforce_tenant_isolation TO public; 