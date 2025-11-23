-- Seed data for development/testing

-- Insert demo tenant
INSERT INTO tenants (
    id,
    name,
    slug,
    subscription_status,
    subscription_plan,
    trial_ends_at,
    max_vehicles,
    max_users,
    max_api_calls_per_month,
    owner_email
) VALUES (
    '550e8400-e29b-41d4-a716-446655440000', -- Fixed UUID for testing
    'Demo Dealership',
    'demo-dealership',
    'trial',
    'professional',
    CURRENT_TIMESTAMP + INTERVAL '14 days',
    1000,
    10,
    100000,
    'owner@demo-dealership.com'
);

-- Insert demo owner user
-- Password: "password123" (hashed with bcrypt cost 10)
INSERT INTO users (
    id,
    tenant_id,
    email,
    password_hash,
    first_name,
    last_name,
    role
) VALUES (
    '660e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'owner@demo-dealership.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password123
    'John',
    'Doe',
    'owner'
);

-- Insert demo admin user
INSERT INTO users (
    id,
    tenant_id,
    email,
    password_hash,
    first_name,
    last_name,
    role
) VALUES (
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@demo-dealership.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', -- password123
    'Jane',
    'Smith',
    'admin'
);

-- Insert some demo cars
INSERT INTO cars (id, tenant_id, make, model, year, color, vin, stock_number, price, mileage, status, description, created_by) VALUES
('car-001', '550e8400-e29b-41d4-a716-446655440000', 'Toyota', 'Camry', 2023, 'Silver', '1HGCM82633A123456', 'STK-001', 28999.00, 12500, 'available', 'Excellent condition, one owner', '660e8400-e29b-41d4-a716-446655440000'),
('car-002', '550e8400-e29b-41d4-a716-446655440000', 'Honda', 'Accord', 2022, 'Black', '1HGCM82633A123457', 'STK-002', 26499.00, 18200, 'available', 'Clean title, well maintained', '660e8400-e29b-41d4-a716-446655440000'),
('car-003', '550e8400-e29b-41d4-a716-446655440000', 'Ford', 'F-150', 2024, 'Blue', '1FTFW1E50PFA12345', 'STK-003', 45999.00, 5000, 'available', 'Brand new, loaded with features', '660e8400-e29b-41d4-a716-446655440000'),
('car-004', '550e8400-e29b-41d4-a716-446655440000', 'Tesla', 'Model 3', 2023, 'White', '5YJ3E1EA5KF123456', 'STK-004', 42999.00, 8000, 'reserved', 'Full self-driving capability', '660e8400-e29b-41d4-a716-446655440000'),
('car-005', '550e8400-e29b-41d4-a716-446655440000', 'Chevrolet', 'Silverado', 2023, 'Red', '1GCVKPEC5DZ123456', 'STK-005', 38999.00, 15000, 'sold', 'Previous rental fleet vehicle', '660e8400-e29b-41d4-a716-446655440000');

-- Initialize usage tracking for current month
INSERT INTO usage_tracking (tenant_id, date, api_calls, vehicle_count, user_count) VALUES
('550e8400-e29b-41d4-a716-446655440000', CURRENT_DATE, 0, 5, 2);
