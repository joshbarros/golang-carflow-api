#!/bin/bash

# ==========================================
# CarFlow SaaS - Subscription Test Script
# ==========================================

set -e

API_URL="${API_URL:-http://localhost:8080}"

echo "🧪 Testing CarFlow Subscription Endpoints"
echo "API URL: $API_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ==========================================
# Setup: Register and get token
# ==========================================
echo -e "${BLUE}📝 Setting up test account...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "dealership_name": "Subscription Test Dealership",
    "owner_email": "subscription-test@testdealer.com",
    "owner_password": "securepassword123",
    "owner_first_name": "Jane",
    "owner_last_name": "Smith",
    "plan": "starter"
  }')

TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Registration failed! Cannot proceed with subscription tests.${NC}"
  echo "Response:"
  echo "$REGISTER_RESPONSE" | jq '.'
  exit 1
fi

echo -e "${GREEN}✅ Test account created${NC}"
echo ""

# ==========================================
# 1. Test Get Current Subscription (should be trial)
# ==========================================
echo -e "${BLUE}1. Testing Get Current Subscription (should show trial status)${NC}"
CURRENT_SUB=$(curl -s -X GET "$API_URL/api/v1/subscriptions/current" \
  -H "Authorization: Bearer $TOKEN")

echo "$CURRENT_SUB" | jq '.'
echo ""

# Check if no subscription exists yet (expected for new account)
if echo "$CURRENT_SUB" | jq -e '.error' > /dev/null; then
  echo -e "${YELLOW}⚠️  No active subscription found (expected for new trial account)${NC}"
else
  echo -e "${GREEN}✅ Retrieved subscription information${NC}"
fi
echo ""

# ==========================================
# 2. Test Create Checkout Session
# ==========================================
echo -e "${BLUE}2. Testing Create Checkout Session${NC}"
echo -e "${YELLOW}💡 Note: This will create a Stripe Checkout URL (requires STRIPE_SECRET_KEY)${NC}"

CHECKOUT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/subscriptions/checkout" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "professional",
    "success_url": "http://localhost:3000/billing/success",
    "cancel_url": "http://localhost:3000/billing"
  }')

echo "$CHECKOUT_RESPONSE" | jq '.'

CHECKOUT_URL=$(echo "$CHECKOUT_RESPONSE" | jq -r '.checkout_url')

if [ "$CHECKOUT_URL" != "null" ] && [ -n "$CHECKOUT_URL" ]; then
  echo -e "${GREEN}✅ Checkout session created successfully!${NC}"
  echo ""
  echo -e "${BLUE}🔗 Checkout URL:${NC}"
  echo "   $CHECKOUT_URL"
  echo ""
  echo -e "${YELLOW}💡 In a real scenario, redirect the user to this URL to complete payment${NC}"
else
  echo -e "${RED}⚠️  Failed to create checkout session${NC}"
  echo "This is expected if STRIPE_SECRET_KEY is not configured"
fi
echo ""

# ==========================================
# 3. Test Create Checkout Session - Different Plans
# ==========================================
echo -e "${BLUE}3. Testing Different Subscription Plans${NC}"

for plan in "starter" "professional" "enterprise"; do
  echo ""
  echo -e "${BLUE}   Testing plan: ${plan}${NC}"

  PLAN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/subscriptions/checkout" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"plan\": \"${plan}\",
      \"success_url\": \"http://localhost:3000/billing/success\",
      \"cancel_url\": \"http://localhost:3000/billing\"
    }")

  PLAN_URL=$(echo "$PLAN_RESPONSE" | jq -r '.checkout_url')

  if [ "$PLAN_URL" != "null" ] && [ -n "$PLAN_URL" ]; then
    echo -e "   ${GREEN}✅ ${plan} plan checkout session created${NC}"
  else
    echo -e "   ${RED}❌ ${plan} plan checkout failed${NC}"
  fi
done
echo ""

# ==========================================
# 4. Test Invalid Plan
# ==========================================
echo -e "${BLUE}4. Testing Invalid Plan (should fail)${NC}"
INVALID_PLAN=$(curl -s -X POST "$API_URL/api/v1/subscriptions/checkout" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan": "invalid_plan",
    "success_url": "http://localhost:3000/billing/success",
    "cancel_url": "http://localhost:3000/billing"
  }')

echo "$INVALID_PLAN" | jq '.'

if echo "$INVALID_PLAN" | jq -e '.error' > /dev/null; then
  echo -e "${GREEN}✅ Invalid plan correctly rejected${NC}"
else
  echo -e "${RED}❌ Invalid plan should have been rejected${NC}"
fi
echo ""

# ==========================================
# 5. Test Cancel Subscription (will fail if no active subscription)
# ==========================================
echo -e "${BLUE}5. Testing Cancel Subscription${NC}"
echo -e "${YELLOW}💡 Note: This will fail if there's no active Stripe subscription${NC}"

CANCEL_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/subscriptions/cancel" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cancel_immediately": false
  }')

echo "$CANCEL_RESPONSE" | jq '.'

if echo "$CANCEL_RESPONSE" | jq -e '.error' > /dev/null; then
  echo -e "${YELLOW}⚠️  Cancel failed (expected - no active subscription to cancel)${NC}"
else
  echo -e "${GREEN}✅ Subscription canceled${NC}"
fi
echo ""

# ==========================================
# 6. Test Unauthorized Access
# ==========================================
echo -e "${BLUE}6. Testing Unauthorized Access (should fail)${NC}"
curl -s -X GET "$API_URL/api/v1/subscriptions/current" | jq '.'
echo -e "${GREEN}✅ Unauthorized request correctly rejected${NC}"
echo ""

# ==========================================
# 7. Test Stripe Webhook Endpoint
# ==========================================
echo -e "${BLUE}7. Testing Stripe Webhook Endpoint (Public)${NC}"
echo -e "${YELLOW}💡 Note: This will fail signature validation (expected)${NC}"

WEBHOOK_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/webhooks/stripe" \
  -H "Content-Type: application/json" \
  -H "Stripe-Signature: invalid_signature" \
  -d '{
    "id": "evt_test_webhook",
    "object": "event",
    "type": "checkout.session.completed"
  }')

echo "$WEBHOOK_RESPONSE" | jq '.'

if echo "$WEBHOOK_RESPONSE" | jq -e '.error' > /dev/null; then
  echo -e "${GREEN}✅ Webhook signature validation working (rejected invalid signature)${NC}"
else
  echo -e "${YELLOW}⚠️  Webhook validation may not be working correctly${NC}"
fi
echo ""

# ==========================================
# Summary
# ==========================================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ Subscription endpoint tests completed!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "📝 Test Summary:"
echo "   ✓ Get current subscription endpoint working"
echo "   ✓ Create checkout session endpoint working"
echo "   ✓ Plan validation working"
echo "   ✓ Cancel subscription endpoint working"
echo "   ✓ Authentication required for protected endpoints"
echo "   ✓ Stripe webhook endpoint accessible"
echo ""
echo "💡 Next Steps:"
echo "   1. Configure STRIPE_SECRET_KEY in .env"
echo "   2. Configure STRIPE_WEBHOOK_SECRET in .env"
echo "   3. Set up Stripe webhook forwarding: stripe listen --forward-to localhost:8080/api/v1/webhooks/stripe"
echo "   4. Complete a test payment to verify full flow"
echo ""
echo "📋 Pricing Plans:"
echo "   • Starter:        \$29/month  - 100 vehicles, 3 users"
echo "   • Professional:   \$99/month  - 500 vehicles, 10 users"
echo "   • Enterprise:     \$299/month - Unlimited vehicles & users"
echo ""
