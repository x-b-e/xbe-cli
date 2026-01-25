#!/bin/bash
#
# XBE CLI Integration Tests: Notifications
#
# Tests list/show/update operations for the notifications resource.
#
# COVERAGE: All list filters + read update
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_USER_ID=""
SAMPLE_READ=""
SAMPLE_READY=""
SAMPLE_APPROACH=""
SAMPLE_DELIVER_AT=""
SAMPLE_NOTIFICATION_TYPE=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""

UPDATE_ID=""
UPDATE_READ=""

describe "Resource: notifications"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List notifications"
xbe_json view notifications list --limit 5
assert_success

test_name "List notifications returns array"
xbe_json view notifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_USER_ID=$(echo "$output" | jq -r '.[0].user_id // empty')
    SAMPLE_READ=$(echo "$output" | jq -r '.[0].read // empty')
    SAMPLE_READY=$(echo "$output" | jq -r '.[0].is_ready_for_delivery // empty')
    SAMPLE_APPROACH=$(echo "$output" | jq -r '.[0].delivery_decision_approach // empty')
    SAMPLE_DELIVER_AT=$(echo "$output" | jq -r '.[0].deliver_at // empty')
    SAMPLE_NOTIFICATION_TYPE=$(echo "$output" | jq -r '.[0].notification_type // empty')
    SAMPLE_CREATED_AT=$(echo "$output" | jq -r '.[0].created_at // empty')
    SAMPLE_UPDATED_AT=$(echo "$output" | jq -r '.[0].updated_at // empty')
    UPDATE_ID=$(echo "$output" | jq -r '.[] | select(.notification_type == "BrokerTenderOfferedSellerNotification" or .notification_type == "CustomerTenderOfferedBuyerNotification") | .id' | head -n 1)
    UPDATE_READ=$(echo "$output" | jq -r '.[] | select(.notification_type == "BrokerTenderOfferedSellerNotification" or .notification_type == "CustomerTenderOfferedBuyerNotification") | .read' | head -n 1)
else
    fail "Failed to list notifications"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show notification"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view notifications show "$SAMPLE_ID"
    assert_success
else
    skip "No notification ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update notification read status"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    READ_TARGET="true"
    if [[ "$UPDATE_READ" == "true" ]]; then
        READ_TARGET="false"
    fi
    xbe_json do notifications update "$UPDATE_ID" --read="$READ_TARGET"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update notification: $output"
        fi
    fi
else
    skip "No notification available for update"
fi

test_name "Update notification without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do notifications update "$UPDATE_ID"
    assert_failure
else
    skip "No notification ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List notifications with --user filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view notifications list --user "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List notifications with --read filter"
if [[ -n "$SAMPLE_READ" && "$SAMPLE_READ" != "null" ]]; then
    xbe_json view notifications list --read "$SAMPLE_READ" --limit 5
    assert_success
else
    skip "No read status available for filter test"
fi

test_name "List notifications with --delivery-decision-approach filter"
if [[ -n "$SAMPLE_APPROACH" && "$SAMPLE_APPROACH" != "null" ]]; then
    xbe_json view notifications list --delivery-decision-approach "$SAMPLE_APPROACH" --limit 5
    assert_success
else
    skip "No delivery decision approach available for filter test"
fi

test_name "List notifications with --is-ready-for-delivery filter"
if [[ -n "$SAMPLE_READY" && "$SAMPLE_READY" != "null" ]]; then
    xbe_json view notifications list --is-ready-for-delivery "$SAMPLE_READY" --limit 5
    assert_success
else
    skip "No ready-for-delivery status available for filter test"
fi

test_name "List notifications with --deliver-at filter"
if [[ -n "$SAMPLE_DELIVER_AT" && "$SAMPLE_DELIVER_AT" != "null" ]]; then
    xbe_json view notifications list --deliver-at "$SAMPLE_DELIVER_AT" --limit 5
    assert_success
else
    skip "No deliver-at value available for filter test"
fi

test_name "List notifications with --deliver-at-min filter"
if [[ -n "$SAMPLE_DELIVER_AT" && "$SAMPLE_DELIVER_AT" != "null" ]]; then
    xbe_json view notifications list --deliver-at-min "$SAMPLE_DELIVER_AT" --limit 5
    assert_success
else
    skip "No deliver-at value available for deliver-at-min filter test"
fi

test_name "List notifications with --deliver-at-max filter"
if [[ -n "$SAMPLE_DELIVER_AT" && "$SAMPLE_DELIVER_AT" != "null" ]]; then
    xbe_json view notifications list --deliver-at-max "$SAMPLE_DELIVER_AT" --limit 5
    assert_success
else
    skip "No deliver-at value available for deliver-at-max filter test"
fi

test_name "List notifications with --notification-type filter"
if [[ -n "$SAMPLE_NOTIFICATION_TYPE" && "$SAMPLE_NOTIFICATION_TYPE" != "null" ]]; then
    xbe_json view notifications list --notification-type "$SAMPLE_NOTIFICATION_TYPE" --limit 5
    assert_success
else
    skip "No notification type available for filter test"
fi

test_name "List notifications with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view notifications list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at value available for filter test"
fi

test_name "List notifications with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view notifications list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at value available for filter test"
fi

test_name "List notifications with --is-created-at filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view notifications list --is-created-at true --limit 5
    assert_success
else
    skip "No created-at value available for is-created-at filter test"
fi

test_name "List notifications with --updated-at-min filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view notifications list --updated-at-min "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated-at value available for filter test"
fi

test_name "List notifications with --updated-at-max filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view notifications list --updated-at-max "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated-at value available for filter test"
fi

test_name "List notifications with --is-updated-at filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view notifications list --is-updated-at true --limit 5
    assert_success
else
    skip "No updated-at value available for is-updated-at filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
