#!/bin/bash

# Session Timeout Testing Script
# This script helps verify that session timeout is working correctly

echo "=========================================="
echo "Gophish Session Timeout Testing Script"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
GOPHISH_URL="${GOPHISH_URL:-http://localhost:3333}"
TEST_TIMEOUT="${TEST_TIMEOUT:-1}"  # 1 minute for testing (change to 15 for production)
COOKIE_FILE="test_session_cookies.txt"

echo "Configuration:"
echo "  Gophish URL: $GOPHISH_URL"
echo "  Test Timeout: $TEST_TIMEOUT minutes"
echo ""

# Check if URL is accessible
echo -n "Checking if Gophish is accessible... "
if curl -s -o /dev/null -w "%{http_code}" "$GOPHISH_URL/login" | grep -q "200\|302"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "Error: Cannot reach Gophish at $GOPHISH_URL"
    echo "Please set GOPHISH_URL environment variable if using a different URL"
    exit 1
fi

echo ""
echo "=========================================="
echo "Test 1: Login and Verify Session Cookie"
echo "=========================================="
echo ""
echo "Please login manually and provide your credentials:"
echo "  URL: $GOPHISH_URL/login"
echo ""
echo "After logging in, press Enter to continue..."
read

# Check if cookie file exists (user should have logged in via browser)
if [ ! -f "$COOKIE_FILE" ]; then
    echo ""
    echo -e "${YELLOW}Note:${NC} For automated testing, you can:"
    echo "  1. Login via browser"
    echo "  2. Export cookies using browser extension"
    echo "  3. Save to $COOKIE_FILE"
    echo ""
    echo "Or test manually:"
    echo "  1. Login to $GOPHISH_URL"
    echo "  2. Wait $TEST_TIMEOUT minutes without activity"
    echo "  3. Try to access a protected page"
    echo "  4. Verify you are redirected to /login"
    exit 0
fi

echo ""
echo "=========================================="
echo "Test 2: Verify Session Persistence"
echo "=========================================="
echo ""

# Test accessing protected endpoint with cookie
echo -n "Testing session cookie validity... "
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -b "$COOKIE_FILE" "$GOPHISH_URL/campaigns")
if [ "$RESPONSE" = "200" ]; then
    echo -e "${GREEN}✓ Session is valid${NC}"
else
    echo -e "${RED}✗ Session is invalid (HTTP $RESPONSE)${NC}"
    echo "  This might be expected if session has expired"
fi

echo ""
echo "=========================================="
echo "Test 3: Wait for Timeout"
echo "=========================================="
echo ""
echo -e "${YELLOW}Waiting $TEST_TIMEOUT minutes for session timeout...${NC}"
echo "  (You can modify TEST_TIMEOUT environment variable)"
echo ""

# Convert minutes to seconds
WAIT_SECONDS=$((TEST_TIMEOUT * 60))

# Countdown
for i in $(seq $WAIT_SECONDS -1 1); do
    minutes=$((i / 60))
    seconds=$((i % 60))
    printf "\r  Time remaining: %02d:%02d" $minutes $seconds
    sleep 1
done
echo ""

echo ""
echo "=========================================="
echo "Test 4: Verify Session Expiration"
echo "=========================================="
echo ""

# Test accessing protected endpoint after timeout
echo -n "Testing session after timeout... "
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -L -b "$COOKIE_FILE" "$GOPHISH_URL/campaigns")
if [ "$RESPONSE" = "200" ]; then
    echo -e "${RED}✗ Session is still valid (should have expired)${NC}"
    echo "  This indicates the timeout may not be working correctly"
    exit 1
elif [ "$RESPONSE" = "302" ] || [ "$RESPONSE" = "401" ] || [ "$RESPONSE" = "403" ]; then
    echo -e "${GREEN}✓ Session expired correctly (HTTP $RESPONSE - redirected/unauthorized)${NC}"
else
    echo -e "${YELLOW}? Unexpected response (HTTP $RESPONSE)${NC}"
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo "Manual Testing Steps:"
echo "  1. Login to $GOPHISH_URL"
echo "  2. Open browser DevTools (F12)"
echo "  3. Go to Application → Cookies"
echo "  4. Verify 'gophish' cookie exists with Max-Age=900 (15 minutes)"
echo "  5. Wait 15 minutes without activity"
echo "  6. Try accessing /campaigns or any protected page"
echo "  7. Verify you are redirected to /login"
echo ""
echo "For faster testing, set environment variable:"
echo "  export SESSION_TIMEOUT_MINUTES=1"
echo "  (Then restart Gophish)"
echo ""

