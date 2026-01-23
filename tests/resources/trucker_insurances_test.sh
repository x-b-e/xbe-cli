#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Insurances
#
# Tests operations for the trucker-insurances resource.
# Trucker insurances require a trucker relationship and have specific phone number
# validation. This test focuses on list operations and error cases.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: trucker-insurances"

# ============================================================================
# Prerequisites - Create resources for filtering tests
# ============================================================================

test_name "Create prerequisite broker for trucker insurance tests"
BROKER_NAME=$(unique_name "InsuranceTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker"
    fi
fi

test_name "Create prerequisite trucker for trucker insurance tests"
TRUCKER_NAME=$(unique_name "InsuranceTrucker")

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "100 Insurance Test Lane"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
    fi
else
    fail "Failed to create trucker"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucker insurances"
xbe_json view trucker-insurances list --limit 5
assert_success

test_name "List trucker insurances returns array"
xbe_json view trucker-insurances list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker insurances"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

# Note: --trucker filter is not allowed by the API, skipping filter tests

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List trucker insurances with --limit"
xbe_json view trucker-insurances list --limit 3
assert_success

test_name "List trucker insurances with --offset"
xbe_json view trucker-insurances list --limit 3 --offset 1
assert_success

test_name "List trucker insurances with pagination (limit + offset)"
xbe_json view trucker-insurances list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trucker insurance without trucker fails"
xbe_json do trucker-insurances create --company-name "Test"
assert_failure

test_name "Update without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do trucker-insurances update "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
