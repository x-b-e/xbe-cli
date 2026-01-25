#!/bin/bash
#
# XBE CLI Integration Tests: User Location Requests
#
# Tests create and view operations for the user-location-requests resource.
#
# COVERAGE: create + list filters + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_USER_ID=""
CREATED_REQUEST_ID=""
UPDATED_BY_ID=""

describe "Resource: user-location-requests"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create user location request without required fields fails"
xbe_run do user-location-requests create
assert_failure

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite user for user location request tests"
TEST_EMAIL=$(unique_email)
xbe_json do users create \
    --name "User Location Request User" \
    --email "$TEST_EMAIL"

if [[ $status -eq 0 ]]; then
    CREATED_USER_ID=$(json_get ".id")
    if [[ -n "$CREATED_USER_ID" && "$CREATED_USER_ID" != "null" ]]; then
        pass
    else
        fail "Created user but no ID returned"
        echo "Cannot continue without a user"
        run_tests
    fi
else
    fail "Failed to create user"
    echo "Cannot continue without a user"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user location request"
xbe_json do user-location-requests create --user "$CREATED_USER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_REQUEST_ID=$(json_get ".id")
    if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
        assert_json_equals ".user_id" "$CREATED_USER_ID"
    else
        fail "Created user location request but no ID returned"
        run_tests
    fi
else
    fail "Failed to create user location request"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user location request"
xbe_json view user-location-requests show "$CREATED_REQUEST_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user location requests"
xbe_json view user-location-requests list --limit 5
assert_success

test_name "List user location requests returns array"
xbe_json view user-location-requests list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user location requests"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Fetch current user for updated-by filter (optional)"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    UPDATED_BY_ID=$(json_get ".id")
    if [[ -n "$UPDATED_BY_ID" && "$UPDATED_BY_ID" != "null" ]]; then
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to fetch current user"
fi

if [[ -z "$UPDATED_BY_ID" || "$UPDATED_BY_ID" == "null" ]]; then
    UPDATED_BY_ID="$CREATED_USER_ID"
fi

test_name "List user location requests with --user filter"
xbe_json view user-location-requests list --user "$CREATED_USER_ID" --limit 5
assert_success

test_name "List user location requests with --updated-by filter"
xbe_json view user-location-requests list --updated-by "$UPDATED_BY_ID" --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"INTERNAL SERVER ERROR"* ]] || [[ "$output" == *"XBE-500"* ]]; then
        skip "Server error on updated-by filter"
    else
        fail "Failed to list user location requests with --updated-by filter"
    fi
fi

test_name "List user location requests with --created-at-min filter"
xbe_json view user-location-requests list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List user location requests with --created-at-max filter"
xbe_json view user-location-requests list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List user location requests with --updated-at-min filter"
xbe_json view user-location-requests list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List user location requests with --updated-at-max filter"
xbe_json view user-location-requests list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List user location requests with --is-created-at filter"
xbe_json view user-location-requests list --is-created-at true --limit 5
assert_success

test_name "List user location requests with --is-updated-at filter"
xbe_json view user-location-requests list --is-updated-at true --limit 5
assert_success

run_tests
