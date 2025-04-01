-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    plan VARCHAR(20) NOT NULL CHECK (plan IN ('basic', 'pro', 'enterprise')),
    features JSONB NOT NULL DEFAULT '[]',
    limits JSONB NOT NULL,
    custom_domain VARCHAR(255) UNIQUE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on custom_domain
CREATE INDEX idx_tenants_custom_domain ON tenants(custom_domain);

-- Create index on status
CREATE INDEX idx_tenants_status ON tenants(status);

-- Add tenant_id column to users table
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);

-- Add tenant_id column to cars table
ALTER TABLE cars ADD COLUMN tenant_id UUID REFERENCES tenants(id);
CREATE INDEX idx_cars_tenant_id ON cars(tenant_id);

-- Enable Row Level Security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE cars ENABLE ROW LEVEL SECURITY;

-- Create RLS policies
CREATE POLICY tenant_isolation_policy ON users
    FOR ALL
    TO authenticated
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_policy ON cars
    FOR ALL
    TO authenticated
    USING (tenant_id = current_tenant_id());

-- Create function to get current tenant ID
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
BEGIN
    RETURN current_setting('app.tenant_id')::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql; 