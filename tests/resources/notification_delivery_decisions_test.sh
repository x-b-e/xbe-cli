#!/bin/bash
#
# XBE CLI Integration Tests: Notification Delivery Decisions
#
# Tests view operations for notification delivery decisions.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FIRST_DECISION_ID=""
NOTIFICATION_ID=""
NOTIFICATION_USER_ID=""

describe "Resource: notification-delivery-decisions (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List notification delivery decisions"
xbe_json view notification-delivery-decisions list --limit 5
assert_success

test_name "List notification delivery decisions returns array"
xbe_json view notification-delivery-decisions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list notification delivery decisions"
fi

# Capture IDs for downstream tests
xbe_json view notification-delivery-decisions list --limit 5
if [[ $status -eq 0 ]]; then
    FIRST_DECISION_ID=$(json_get ".[0].id")
    NOTIFICATION_ID=$(json_get ".[0].notification_id")
    NOTIFICATION_USER_ID=$(json_get ".[0].notification_user_id")
else
    FIRST_DECISION_ID=""
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show notification delivery decision"
if [[ -n "$FIRST_DECISION_ID" && "$FIRST_DECISION_ID" != "null" ]]; then
    xbe_json view notification-delivery-decisions show "$FIRST_DECISION_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Failed to show notification delivery decision"
    fi
else
    skip "No notification delivery decision ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List decisions with --notification filter"
if [[ -n "$NOTIFICATION_ID" && "$NOTIFICATION_ID" != "null" ]]; then
    xbe_json view notification-delivery-decisions list --notification "$NOTIFICATION_ID" --limit 5
    assert_success
else
    skip "No notification ID available for filter test"
fi

test_name "List decisions with --notification-user filter"
if [[ -n "$NOTIFICATION_USER_ID" && "$NOTIFICATION_USER_ID" != "null" ]]; then
    xbe_json view notification-delivery-decisions list --notification-user "$NOTIFICATION_USER_ID" --limit 5
    assert_success
else
    skip "No notification user ID available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
