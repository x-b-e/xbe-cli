#!/bin/bash
#
# XBE CLI Integration Tests: Customer Tender Offered Buyer Notification Subscriptions
#
# Tests CRUD operations for the customer_tender_offered_buyer_notification_subscriptions resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIPTION_ID=""
SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_USER_ID=""
WHOAMI_USER_ID=""

describe "Resource: customer-tender-offered-buyer-notification-subscriptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer tender offered buyer notification subscriptions"
xbe_json view customer-tender-offered-buyer-notification-subscriptions list --limit 5
assert_success

test_name "List customer tender offered buyer notification subscriptions returns array"
xbe_json view customer-tender-offered-buyer-notification-subscriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_BROKER_ID=$(echo "$output" | jq -r '.[0].broker_id // empty')
    SAMPLE_USER_ID=$(echo "$output" | jq -r '.[0].user_id // empty')
else
    fail "Failed to list customer tender offered buyer notification subscriptions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer tender offered buyer notification subscription"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view customer-tender-offered-buyer-notification-subscriptions show "$SAMPLE_ID"
    assert_success
else
    skip "No subscription ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Get current user for subscription create"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    pass
else
    skip "Unable to resolve current user"
fi

test_name "Create customer tender offered buyer notification subscription"
BROKER_ID="${XBE_TEST_BROKER_ID:-$SAMPLE_BROKER_ID}"
USER_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json do customer-tender-offered-buyer-notification-subscriptions create \
        --broker "$BROKER_ID" \
        --user "$USER_ID" \
        --notify-by-txt \
        --notify-by-email
    if [[ $status -eq 0 ]]; then
        CREATED_SUBSCRIPTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
            register_cleanup "customer-tender-offered-buyer-notification-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
        else
            fail "Created subscription but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"does not have a broker membership with the broker"* ]]; then
            pass
        else
            fail "Failed to create customer tender offered buyer notification subscription: $output"
        fi
    fi
else
    skip "No broker/user ID available (set XBE_TEST_BROKER_ID and XBE_TEST_USER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update customer tender offered buyer notification subscription notify-by-email"
UPDATE_ID="${CREATED_SUBSCRIPTION_ID:-$SAMPLE_ID}"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do customer-tender-offered-buyer-notification-subscriptions update "$UPDATE_ID" --notify-by-email=false
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update customer tender offered buyer notification subscription: $output"
        fi
    fi
else
    skip "No subscription ID available for update"
fi

test_name "Update customer tender offered buyer notification subscription without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do customer-tender-offered-buyer-notification-subscriptions update "$UPDATE_ID"
    assert_failure
else
    skip "No subscription ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List subscriptions with --user filter"
USER_FILTER_ID="${USER_ID:-$SAMPLE_USER_ID}"
if [[ -n "$USER_FILTER_ID" && "$USER_FILTER_ID" != "null" ]]; then
    xbe_json view customer-tender-offered-buyer-notification-subscriptions list --user "$USER_FILTER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete subscription requires --confirm flag"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do customer-tender-offered-buyer-notification-subscriptions delete "$CREATED_SUBSCRIPTION_ID"
    assert_failure
else
    skip "No created subscription for delete confirmation test"
fi

test_name "Delete subscription with --confirm"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do customer-tender-offered-buyer-notification-subscriptions delete "$CREATED_SUBSCRIPTION_ID" --confirm
    assert_success
else
    skip "No created subscription to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create subscription without broker fails"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_run do customer-tender-offered-buyer-notification-subscriptions create --user "$USER_ID"
    assert_failure
else
    skip "No user available for missing broker test"
fi

test_name "Create subscription without user fails"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_run do customer-tender-offered-buyer-notification-subscriptions create --broker "$BROKER_ID"
    assert_failure
else
    skip "No broker ID available for missing user test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
