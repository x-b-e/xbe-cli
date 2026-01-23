#!/bin/bash
#
# XBE CLI Integration Tests: Customer Settings
#
# Tests operations for the customer-settings resource.
# Note: Customer settings are organization-level settings - only update is available (no create/delete).
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CUSTOMER_SETTING_ID=""

describe "Resource: customer-settings"

# ============================================================================
# LIST Tests - Get a customer setting ID for update tests
# ============================================================================

test_name "List customer settings"
xbe_json view customer-settings list --limit 5
assert_success

test_name "List customer settings returns array"
xbe_json view customer-settings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer settings"
fi

test_name "Get a customer setting ID for update tests"
xbe_json view customer-settings list --limit 1
if [[ $status -eq 0 ]]; then
    CUSTOMER_SETTING_ID=$(json_get ".[0].id")
    if [[ -n "$CUSTOMER_SETTING_ID" && "$CUSTOMER_SETTING_ID" != "null" ]]; then
        pass
    else
        skip "No customer settings found in the system"
        run_tests
    fi
else
    fail "Failed to list customer settings"
    run_tests
fi

# ============================================================================
# UPDATE Tests - Boolean Attributes
# ============================================================================

test_name "Update customer setting is-auditing-time-card-approvals to true"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID" --is-auditing-time-card-approvals
assert_success

test_name "Update customer setting enable-recap-notifications to true"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID" --enable-recap-notifications
assert_success

test_name "Update customer setting plan-requires-project to true"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID" --plan-requires-project
assert_success

test_name "Update customer setting plan-requires-business-unit to true"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID" --plan-requires-business-unit
assert_success

test_name "Update customer setting auto-cancel-shifts-without-activity to true"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID" --auto-cancel-shifts-without-activity
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List customer settings with --limit"
xbe_json view customer-settings list --limit 3
assert_success

test_name "List customer settings with --offset"
xbe_json view customer-settings list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do customer-settings update "$CUSTOMER_SETTING_ID"
assert_failure

test_name "Update non-existent customer setting fails"
xbe_json do customer-settings update "99999999" --enable-recap-notifications
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
