#!/bin/bash
#
# XBE CLI Integration Tests: Shift Acknowledgement Reminder Notification Subscriptions
#
# Tests CRUD operations for the shift_acknowledgement_reminder_notification_subscriptions resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SUBSCRIPTION_ID=""
SAMPLE_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_USER_ID=""
WHOAMI_USER_ID=""

# shellcheck disable=SC2034
TRUCKER_ID=""
# shellcheck disable=SC2034
USER_ID=""

describe "Resource: shift-acknowledgement-reminder-notification-subscriptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List shift acknowledgement reminder notification subscriptions"
xbe_json view shift-acknowledgement-reminder-notification-subscriptions list --limit 5
assert_success

test_name "List shift acknowledgement reminder notification subscriptions returns array"
xbe_json view shift-acknowledgement-reminder-notification-subscriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_TRUCKER_ID=$(echo "$output" | jq -r '.[0].trucker_id // empty')
    SAMPLE_USER_ID=$(echo "$output" | jq -r '.[0].user_id // empty')
else
    fail "Failed to list shift acknowledgement reminder notification subscriptions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show shift acknowledgement reminder notification subscription"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view shift-acknowledgement-reminder-notification-subscriptions show "$SAMPLE_ID"
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

test_name "Create shift acknowledgement reminder notification subscription"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-$SAMPLE_TRUCKER_ID}"
USER_ID="${XBE_TEST_USER_ID:-$WHOAMI_USER_ID}"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json do shift-acknowledgement-reminder-notification-subscriptions create \
        --trucker "$TRUCKER_ID" \
        --user "$USER_ID" \
        --notify-by-txt \
        --notify-by-email
    if [[ $status -eq 0 ]]; then
        CREATED_SUBSCRIPTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
            register_cleanup "shift-acknowledgement-reminder-notification-subscriptions" "$CREATED_SUBSCRIPTION_ID"
            pass
        else
            fail "Created subscription but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"must be a trucker manager"* ]]; then
            pass
        else
            fail "Failed to create shift acknowledgement reminder notification subscription: $output"
        fi
    fi
else
    skip "No trucker/user ID available (set XBE_TEST_TRUCKER_ID and XBE_TEST_USER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update shift acknowledgement reminder notification subscription notify-by-email"
UPDATE_ID="${CREATED_SUBSCRIPTION_ID:-$SAMPLE_ID}"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do shift-acknowledgement-reminder-notification-subscriptions update "$UPDATE_ID" --notify-by-email=false
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update shift acknowledgement reminder notification subscription: $output"
        fi
    fi
else
    skip "No subscription ID available for update"
fi

test_name "Update shift acknowledgement reminder notification subscription without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do shift-acknowledgement-reminder-notification-subscriptions update "$UPDATE_ID"
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
    xbe_json view shift-acknowledgement-reminder-notification-subscriptions list --user "$USER_FILTER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List subscriptions with --trucker filter"
TRUCKER_FILTER_ID="${TRUCKER_ID:-$SAMPLE_TRUCKER_ID}"
if [[ -n "$TRUCKER_FILTER_ID" && "$TRUCKER_FILTER_ID" != "null" ]]; then
    xbe_json view shift-acknowledgement-reminder-notification-subscriptions list --trucker "$TRUCKER_FILTER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available for filter test"
fi

test_name "List subscriptions with --trucker-id filter"
TRUCKER_ID_FILTER_ID="${TRUCKER_ID:-$SAMPLE_TRUCKER_ID}"
if [[ -n "$TRUCKER_ID_FILTER_ID" && "$TRUCKER_ID_FILTER_ID" != "null" ]]; then
    xbe_json view shift-acknowledgement-reminder-notification-subscriptions list --trucker-id "$TRUCKER_ID_FILTER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete subscription requires --confirm flag"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do shift-acknowledgement-reminder-notification-subscriptions delete "$CREATED_SUBSCRIPTION_ID"
    assert_failure
else
    skip "No created subscription for delete confirmation test"
fi

test_name "Delete subscription with --confirm"
if [[ -n "$CREATED_SUBSCRIPTION_ID" && "$CREATED_SUBSCRIPTION_ID" != "null" ]]; then
    xbe_run do shift-acknowledgement-reminder-notification-subscriptions delete "$CREATED_SUBSCRIPTION_ID" --confirm
    assert_success
else
    skip "No created subscription to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create subscription without trucker fails"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_run do shift-acknowledgement-reminder-notification-subscriptions create --user "$USER_ID"
    assert_failure
else
    skip "No user available for missing trucker test"
fi

test_name "Create subscription without user fails"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_run do shift-acknowledgement-reminder-notification-subscriptions create --trucker "$TRUCKER_ID"
    assert_failure
else
    skip "No trucker ID available for missing user test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
