#!/bin/bash
#
# XBE CLI Integration Tests: Broker Settings
#
# Tests operations for the broker-settings resource.
# Note: Broker settings are organization-level settings - only update is available (no create/delete).
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_SETTING_ID=""

describe "Resource: broker-settings"

# ============================================================================
# LIST Tests - Get a broker setting ID for update tests
# ============================================================================

test_name "List broker settings"
xbe_json view broker-settings list --limit 5
assert_success

test_name "List broker settings returns array"
xbe_json view broker-settings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker settings"
fi

test_name "Get a broker setting ID for update tests"
xbe_json view broker-settings list --limit 1
if [[ $status -eq 0 ]]; then
    BROKER_SETTING_ID=$(json_get ".[0].id")
    if [[ -n "$BROKER_SETTING_ID" && "$BROKER_SETTING_ID" != "null" ]]; then
        pass
    else
        skip "No broker settings found in the system"
        run_tests
    fi
else
    fail "Failed to list broker settings"
    run_tests
fi

# ============================================================================
# UPDATE Tests - Boolean Attributes
# ============================================================================

test_name "Update broker setting is-auditing-time-card-approvals to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --is-auditing-time-card-approvals
assert_success

test_name "Update broker setting enable-recap-notifications to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --enable-recap-notifications
assert_success

test_name "Update broker setting plan-requires-project to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --plan-requires-project
assert_success

test_name "Update broker setting plan-requires-business-unit to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --plan-requires-business-unit
assert_success

test_name "Update broker setting auto-cancel-shifts-without-activity to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --auto-cancel-shifts-without-activity
assert_success

test_name "Update broker setting restrict-contact-info-visibility to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --restrict-contact-info-visibility
assert_success

test_name "Update broker setting require-explicit-rate-editing-permission to true"
xbe_json do broker-settings update "$BROKER_SETTING_ID" --require-explicit-rate-editing-permission
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker settings with --limit"
xbe_json view broker-settings list --limit 3
assert_success

test_name "List broker settings with --offset"
xbe_json view broker-settings list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do broker-settings update "$BROKER_SETTING_ID"
assert_failure

test_name "Update non-existent broker setting fails"
xbe_json do broker-settings update "99999999" --enable-recap-notifications
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
