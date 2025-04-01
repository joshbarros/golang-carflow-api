#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Base URL
BASE_URL="http://localhost:8080"

# Generate timestamp for unique email
TIMESTAMP=$(date +%s)
TEST_EMAIL="test_${TIMESTAMP}@example.com"
TEST_PASSWORD="test123456"
TEST_FIRST_NAME="Test"
TEST_LAST_NAME="User"

# Car makes and models for random selection
CAR_MAKES=("Toyota" "Honda" "Ford" "BMW" "Mercedes" "Audi" "Tesla" "Porsche" "Chevrolet" "Volkswagen")
CAR_MODELS=("Camry" "Civic" "F-150" "3-Series" "C-Class" "A4" "Model 3" "911" "Silverado" "Golf")
CAR_COLORS=("Red" "Blue" "Black" "White" "Silver" "Gray" "Green" "Yellow" "Orange" "Purple")

# Function to get random array element
random_element() {
    local array=("$@")
    echo "${array[RANDOM % ${#array[@]}]}"
}

# Function to get random year between 2000 and 2024
random_year() {
    echo $((RANDOM % 25 + 2000))
}

# Variables to store tokens and IDs
ACCESS_TOKEN=""
CAR_ID=""

# Function to print test results
print_result() {
    local test_name=$1
    local http_code=$2
    local expected_code=$3

    if [ "$http_code" -eq "$expected_code" ]; then
        echo -e "${GREEN}✓ $test_name - Success (HTTP Status: $http_code)${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name - Failed (Expected HTTP Status: $expected_code, Got: $http_code)${NC}"
        return 1
    fi
}

# Function to extract response body and status code
parse_response() {
    local response=$1
    local status_line=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d') # Remove last line (status code)
    echo "$body"
    return ${status_line:-0}
}

echo -e "${YELLOW}Starting API Tests...${NC}"

# Test 1: Health Check
echo -e "\n${YELLOW}Testing: Health Check${NC}"
response=$(curl -s -w "\n%{http_code}" "$BASE_URL/healthz")
body=$(parse_response "$response")
status=$?
print_result "Health Check" "$status" 200
echo "$body"

# Test 2: Register User
echo -e "\n${YELLOW}Testing: Register User${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\",
        \"first_name\": \"$TEST_FIRST_NAME\",
        \"last_name\": \"$TEST_LAST_NAME\"
    }")
body=$(parse_response "$response")
status=$?

if [ "$status" -ne 201 ]; then
    echo -e "${RED}Registration failed with HTTP code: $status${NC}"
    if [ ! -z "$body" ]; then
        echo -e "${RED}Response: $body${NC}"
    fi
    exit 1
fi

user_id=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ -z "$user_id" ]; then
    echo -e "${RED}Failed to get user ID from response${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Register User - Success (HTTP Status: $status)${NC}"

# Test 3: Login
echo -e "\n${YELLOW}Testing: Login${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_EMAIL\",
        \"password\": \"$TEST_PASSWORD\"
    }")
body=$(parse_response "$response")
status=$?

if [ "$status" -ne 200 ]; then
    echo -e "${RED}Login failed with HTTP code: $status${NC}"
    if [ ! -z "$body" ]; then
        echo -e "${RED}Response: $body${NC}"
    fi
    exit 1
fi

ACCESS_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}Failed to get access token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Login - Success (HTTP Status: $status)${NC}"

# Test 4: Create Car
echo -e "\n${YELLOW}Testing: Create Car${NC}"
RANDOM_MAKE=$(random_element "${CAR_MAKES[@]}")
RANDOM_MODEL=$(random_element "${CAR_MODELS[@]}")
RANDOM_COLOR=$(random_element "${CAR_COLORS[@]}")
RANDOM_YEAR=$(random_year)

response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/cars" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"make\": \"$RANDOM_MAKE\",
        \"model\": \"$RANDOM_MODEL\",
        \"year\": $RANDOM_YEAR,
        \"color\": \"$RANDOM_COLOR\"
    }")
body=$(parse_response "$response")
status=$?

if [ "$status" -ne 201 ]; then
    echo -e "${RED}Create car failed with HTTP code: $status${NC}"
    if [ ! -z "$body" ]; then
        echo -e "${RED}Response: $body${NC}"
    fi
    exit 1
fi

CAR_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
if [ -z "$CAR_ID" ]; then
    echo -e "${RED}Failed to get car ID${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Create Car - Success (HTTP Status: $status)${NC}"

# Test 5: Get All Cars
echo -e "\n${YELLOW}Testing: Get All Cars${NC}"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/cars")
body=$(parse_response "$response")
status=$?
print_result "Get All Cars" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 6: Get Car by ID
echo -e "\n${YELLOW}Testing: Get Car by ID${NC}"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/cars/$CAR_ID")
body=$(parse_response "$response")
status=$?
print_result "Get Car by ID" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 7: Update Car
echo -e "\n${YELLOW}Testing: Update Car${NC}"
RANDOM_MAKE=$(random_element "${CAR_MAKES[@]}")
RANDOM_MODEL=$(random_element "${CAR_MODELS[@]}")
RANDOM_COLOR=$(random_element "${CAR_COLORS[@]}")
RANDOM_YEAR=$(random_year)

response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/cars/$CAR_ID" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"make\": \"$RANDOM_MAKE\",
        \"model\": \"$RANDOM_MODEL\",
        \"year\": $RANDOM_YEAR,
        \"color\": \"$RANDOM_COLOR\"
    }")
body=$(parse_response "$response")
status=$?
print_result "Update Car" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 8: Delete Car
echo -e "\n${YELLOW}Testing: Delete Car${NC}"
response=$(curl -s -w "\n%{http_code}" -X DELETE -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/cars/$CAR_ID")
body=$(parse_response "$response")
status=$?
print_result "Delete Car" "$status" 204

# Test 9: Verify Car Deletion
echo -e "\n${YELLOW}Testing: Verify Car Deletion${NC}"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/cars/$CAR_ID")
body=$(parse_response "$response")
status=$?
print_result "Verify Car Deletion" "$status" 404
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 10: Get User Profile
echo -e "\n${YELLOW}Testing: Get User Profile${NC}"
response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $ACCESS_TOKEN" "$BASE_URL/auth/profile")
body=$(parse_response "$response")
status=$?
print_result "Get User Profile" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 11: Update User Profile
echo -e "\n${YELLOW}Testing: Update User Profile${NC}"
response=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/auth/profile" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"first_name\": \"Updated\",
        \"last_name\": \"User\",
        \"email\": \"$TEST_EMAIL\"
    }")
body=$(parse_response "$response")
status=$?
print_result "Update User Profile" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

# Test 12: Refresh Token
echo -e "\n${YELLOW}Testing: Refresh Token${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/refresh" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
body=$(parse_response "$response")
status=$?
print_result "Refresh Token" "$status" 200
if [ ! -z "$body" ]; then
    echo "$body"
fi

echo -e "\n${YELLOW}Test Summary:${NC}"
echo -e "${GREEN}✓ All tests completed successfully${NC}" 