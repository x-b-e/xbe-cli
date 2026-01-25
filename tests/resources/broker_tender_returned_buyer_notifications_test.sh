#!/bin/bash
#
# XBE CLI Integration Tests: Broker Tender Returned Buyer Notifications
#
# Tests list/show/update operations for broker-tender-returned-buyer-notifications.
# Notifications are read-only except for the read flag.
#
# COVERAGE: list filters + show + update
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

NOTIFICATION_ID=""
NOTIFICATION_USER_ID=""
NOTIFICATION_READ=""
NOTIFICATION_NOTIFICATION_TYPE=""
NOTIFICATION_DELIVERY_DECISION_APPROACH=""
NOTIFICATION_IS_READY_FOR_DELIVERY=""
NOTIFICATION_DELIVER_AT=""

describe "Resource: broker-tender-returned-buyer-notifications"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker tender returned buyer notifications"
xbe_json view broker-tender-returned-buyer-notifications list --limit 5
assert_success

test_name "List broker tender returned buyer notifications returns array"
xbe_json view broker-tender-returned-buyer-notifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list broker tender returned buyer notifications"
fi

# ============================================================================
# LIST Tests - Get an ID for show/update
# ============================================================================

test_name "Get notification ID for show/update tests"
xbe_json view broker-tender-returned-buyer-notifications list --limit 1
if [[ $status -eq 0 ]]; then
    NOTIFICATION_ID=$(json_get ".[0].id")
    NOTIFICATION_USER_ID=$(json_get ".[0].user_id")
    NOTIFICATION_READ=$(json_get ".[0].read")
    NOTIFICATION_DELIVER_AT=$(json_get ".[0].deliver_at")
    if [[ -n "$NOTIFICATION_ID" && "$NOTIFICATION_ID" != "null" ]]; then
        pass
    else
        skip "No broker tender returned buyer notifications found"
        run_tests
    fi
else
    fail "Failed to list broker tender returned buyer notifications"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show broker tender returned buyer notification"
xbe_json view broker-tender-returned-buyer-notifications show "$NOTIFICATION_ID"
assert_success

if [[ $status -eq 0 ]]; then
    NOTIFICATION_NOTIFICATION_TYPE=$(json_get ".notification_type")
    NOTIFICATION_DELIVERY_DECISION_APPROACH=$(json_get ".delivery_decision_approach")
    NOTIFICATION_IS_READY_FOR_DELIVERY=$(json_get ".is_ready_for_delivery")
    if [[ "$NOTIFICATION_DELIVER_AT" == "null" || -z "$NOTIFICATION_DELIVER_AT" ]]; then
        NOTIFICATION_DELIVER_AT=$(json_get ".deliver_at")
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broker tender returned buyer notifications with --user"
if [[ -n "$NOTIFICATION_USER_ID" && "$NOTIFICATION_USER_ID" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --user "$NOTIFICATION_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List broker tender returned buyer notifications with --read"
if [[ -n "$NOTIFICATION_READ" && "$NOTIFICATION_READ" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --read "$NOTIFICATION_READ" --limit 5
    assert_success
else
    skip "No read status available for filter test"
fi

test_name "List broker tender returned buyer notifications with --delivery-decision-approach"
if [[ -n "$NOTIFICATION_DELIVERY_DECISION_APPROACH" && "$NOTIFICATION_DELIVERY_DECISION_APPROACH" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --delivery-decision-approach "$NOTIFICATION_DELIVERY_DECISION_APPROACH" --limit 5
    assert_success
else
    skip "No delivery decision approach available for filter test"
fi

test_name "List broker tender returned buyer notifications with --is-ready-for-delivery"
if [[ -n "$NOTIFICATION_IS_READY_FOR_DELIVERY" && "$NOTIFICATION_IS_READY_FOR_DELIVERY" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --is-ready-for-delivery "$NOTIFICATION_IS_READY_FOR_DELIVERY" --limit 5
    assert_success
else
    skip "No ready-for-delivery value available for filter test"
fi

test_name "List broker tender returned buyer notifications with --deliver-at"
if [[ -n "$NOTIFICATION_DELIVER_AT" && "$NOTIFICATION_DELIVER_AT" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --deliver-at "$NOTIFICATION_DELIVER_AT" --limit 5
    assert_success
else
    skip "No deliver-at value available for filter test"
fi

test_name "List broker tender returned buyer notifications with --notification-type"
if [[ -n "$NOTIFICATION_NOTIFICATION_TYPE" && "$NOTIFICATION_NOTIFICATION_TYPE" != "null" ]]; then
    xbe_json view broker-tender-returned-buyer-notifications list --notification-type "$NOTIFICATION_NOTIFICATION_TYPE" --limit 5
    assert_success
else
    skip "No notification type available for filter test"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List broker tender returned buyer notifications with --limit"
xbe_json view broker-tender-returned-buyer-notifications list --limit 3
assert_success

test_name "List broker tender returned buyer notifications with --offset"
xbe_json view broker-tender-returned-buyer-notifications list --limit 3 --offset 1
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update broker tender returned buyer notification read status"
xbe_json do broker-tender-returned-buyer-notifications update "$NOTIFICATION_ID" --read
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Internal Server Error"* ]]; then
        skip "Server error updating notification (500)"
    else
        fail "Failed to update notification: $output"
    fi
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do broker-tender-returned-buyer-notifications update "$NOTIFICATION_ID"
assert_failure

test_name "Update non-existent broker tender returned buyer notification fails"
xbe_json do broker-tender-returned-buyer-notifications update "99999999" --read
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
