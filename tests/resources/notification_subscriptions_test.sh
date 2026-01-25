#!/bin/bash
#
# XBE CLI Integration Tests: Notification Subscriptions
#
# Tests list and show operations for the notification_subscriptions resource.
# Notification subscriptions define which users receive specific notification types.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBSCRIPTION_ID=""
USER_ID=""
SKIP_SHOW=0
SKIP_USER_FILTER=0

describe "Resource: notification-subscriptions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List notification subscriptions"
xbe_json view notification-subscriptions list --limit 5
assert_success

test_name "List notification subscriptions returns array"
xbe_json view notification-subscriptions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list notification subscriptions"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample notification subscription"
xbe_json view notification-subscriptions list --limit 1
if [[ $status -eq 0 ]]; then
    SUBSCRIPTION_ID=$(json_get ".[0].id")
    USER_ID=$(json_get ".[0].user_id")
    if [[ -n "$SUBSCRIPTION_ID" && "$SUBSCRIPTION_ID" != "null" ]]; then
        pass
    else
        SKIP_SHOW=1
        SKIP_USER_FILTER=1
        skip "No notification subscriptions available"
    fi
else
    SKIP_SHOW=1
    SKIP_USER_FILTER=1
    fail "Failed to list notification subscriptions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List notification subscriptions with --user filter"
if [[ $SKIP_USER_FILTER -eq 0 && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view notification-subscriptions list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List notification subscriptions with --created-at-min filter"
xbe_json view notification-subscriptions list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List notification subscriptions with --created-at-max filter"
xbe_json view notification-subscriptions list --created-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List notification subscriptions with --updated-at-min filter"
xbe_json view notification-subscriptions list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List notification subscriptions with --updated-at-max filter"
xbe_json view notification-subscriptions list --updated-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List notification subscriptions with --is-created-at filter"
xbe_json view notification-subscriptions list --is-created-at true --limit 5
assert_success

test_name "List notification subscriptions with --is-updated-at filter"
xbe_json view notification-subscriptions list --is-updated-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show notification subscription"
if [[ $SKIP_SHOW -eq 0 && -n "$SUBSCRIPTION_ID" && "$SUBSCRIPTION_ID" != "null" ]]; then
    xbe_json view notification-subscriptions show "$SUBSCRIPTION_ID"
    assert_success
else
    skip "No notification subscription ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
