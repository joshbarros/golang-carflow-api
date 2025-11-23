#!/bin/bash

# ==========================================
# CarFlow SaaS - Authentication Test Script
# ==========================================

set -e

API_URL="${API_URL:-http://localhost:8080}"

echo "🧪 Testing CarFlow Authentication Endpoints"
echo "API URL: $API_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# ==========================================
# 1. Test Health Check
# ==========================================
echo -e "${BLUE}1. Testing Health Check${NC}"
curl -s "$API_URL/healthz" | jq '.'
echo ""

# ==========================================
# 2. Test Registration
# ==========================================
echo -e "${BLUE}2. Testing Registration${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "dealership_name": "Test Motors Inc",
    "owner_email": "owner@testmotors.com",
    "owner_password": "securepassword123",
    "owner_first_name": "John",
    "owner_last_name": "Doe",
    "owner_phone": "+1234567890",
    "company_address": "123 Main St, City, State 12345",
    "plan": "professional"
  }')

echo "$REGISTER_RESPONSE" | jq '.'
echo ""

# Extract token from registration response
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Registration failed! Token not found.${NC}"
  echo "Response:"
  echo "$REGISTER_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✅ Registration successful!${NC}"
echo "Token: $TOKEN"
echo ""

# ==========================================
# 3. Test Get Current User
# ==========================================
echo -e "${BLUE}3. Testing Get Current User (Protected Route)${NC}"
curl -s -X GET "$API_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

# ==========================================
# 4. Test Token Refresh
# ==========================================
echo -e "${BLUE}4. Testing Token Refresh${NC}"
REFRESH_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/refresh" \
  -H "Authorization: Bearer $TOKEN")

echo "$REFRESH_RESPONSE" | jq '.'
NEW_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.token')
echo ""

if [ "$NEW_TOKEN" != "null" ] && [ -n "$NEW_TOKEN" ]; then
  echo -e "${GREEN}✅ Token refreshed successfully!${NC}"
  TOKEN="$NEW_TOKEN"
else
  echo -e "${RED}⚠️  Token refresh failed, continuing with original token${NC}"
fi
echo ""

# ==========================================
# 5. Test Login
# ==========================================
echo -e "${BLUE}5. Testing Login with registered credentials${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "owner@testmotors.com",
    "password": "securepassword123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'
echo ""

LOGIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$LOGIN_TOKEN" == "null" ] || [ -z "$LOGIN_TOKEN" ]; then
  echo -e "${RED}❌ Login failed!${NC}"
else
  echo -e "${GREEN}✅ Login successful!${NC}"
  TOKEN="$LOGIN_TOKEN"
fi
echo ""

# ==========================================
# 6. Test Protected Car Endpoint
# ==========================================
echo -e "${BLUE}6. Testing Protected Car Endpoint (List Cars)${NC}"
curl -s -X GET "$API_URL/api/v1/cars" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

# ==========================================
# 7. Test Demo Login (from seed data)
# ==========================================
echo -e "${BLUE}7. Testing Demo Login (from seed data)${NC}"
DEMO_LOGIN=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "owner@demo-dealership.com",
    "password": "password123"
  }')

echo "$DEMO_LOGIN" | jq '.'
DEMO_TOKEN=$(echo "$DEMO_LOGIN" | jq -r '.token')
echo ""

if [ "$DEMO_TOKEN" != "null" ] && [ -n "$DEMO_TOKEN" ]; then
  echo -e "${GREEN}✅ Demo login successful!${NC}"
  echo "Demo Token: $DEMO_TOKEN"
else
  echo -e "${RED}⚠️  Demo login failed (database may not be seeded)${NC}"
fi
echo ""

# ==========================================
# 8. Test Invalid Credentials
# ==========================================
echo -e "${BLUE}8. Testing Invalid Credentials (should fail)${NC}"
INVALID_LOGIN=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@example.com",
    "password": "wrongpassword"
  }')

echo "$INVALID_LOGIN" | jq '.'
echo ""

# ==========================================
# 9. Test Unauthorized Access
# ==========================================
echo -e "${BLUE}9. Testing Unauthorized Access (should fail)${NC}"
curl -s -X GET "$API_URL/api/v1/auth/me" | jq '.'
echo ""

# ==========================================
# Summary
# ==========================================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ All authentication tests completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "💡 Tips:"
echo "   - Use the token in the Authorization header: 'Bearer \$TOKEN'"
echo "   - Tokens expire after 24 hours (configurable)"
echo "   - Use /api/v1/auth/refresh to get a new token"
echo ""
echo "📝 Saved tokens:"
echo "   Registration Token: $TOKEN"
if [ "$DEMO_TOKEN" != "null" ] && [ -n "$DEMO_TOKEN" ]; then
  echo "   Demo Token:         $DEMO_TOKEN"
fi
echo ""
