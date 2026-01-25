#!/bin/bash
#
# XBE CLI Integration Tests: KeepTruckin Users
#
# Tests list, show, and update operations for the keep-truckin-users resource.
#
# COVERAGE: List filters + show + update
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

KEEP_TRUCKIN_USER_ID=""
BROKER_ID=""
USER_ID=""
ROLE=""
ACTIVE=""
SKIP_SHOW=0
SKIP_UPDATE=0
UPDATE_KEEP_TRUCKIN_USER_ID=""
UPDATE_USER_ID=""

describe "Resource: keep-truckin-users"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List KeepTruckin users"
xbe_json view keep-truckin-users list --limit 5
assert_success

test_name "List KeepTruckin users returns array"
xbe_json view keep-truckin-users list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list KeepTruckin users"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample KeepTruckin user"
xbe_json view keep-truckin-users list --limit 1
if [[ $status -eq 0 ]]; then
    KEEP_TRUCKIN_USER_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    USER_ID=$(json_get ".[0].user_id")
    ROLE=$(json_get ".[0].role")
    ACTIVE=$(json_get ".[0].active")
    if [[ -n "$KEEP_TRUCKIN_USER_ID" && "$KEEP_TRUCKIN_USER_ID" != "null" ]]; then
        pass
    else
        SKIP_SHOW=1
        SKIP_UPDATE=1
        skip "No KeepTruckin users available"
    fi
else
    SKIP_SHOW=1
    SKIP_UPDATE=1
    fail "Failed to list KeepTruckin users"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List KeepTruckin users with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view keep-truckin-users list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List KeepTruckin users with --user filter"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view keep-truckin-users list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List KeepTruckin users with --role filter"
if [[ -n "$ROLE" && "$ROLE" != "null" ]]; then
    xbe_json view keep-truckin-users list --role "$ROLE" --limit 5
    assert_success
else
    skip "No role available"
fi

test_name "List KeepTruckin users with --active filter"
if [[ -n "$ACTIVE" && "$ACTIVE" != "null" ]]; then
    xbe_json view keep-truckin-users list --active "$ACTIVE" --limit 5
    assert_success
else
    xbe_json view keep-truckin-users list --active true --limit 5
    assert_success
fi

test_name "List KeepTruckin users with --has-user filter"
xbe_json view keep-truckin-users list --has-user true --limit 5
assert_success

test_name "List KeepTruckin users with --user-set-at-min filter"
xbe_json view keep-truckin-users list --user-set-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List KeepTruckin users with --user-set-at-max filter"
xbe_json view keep-truckin-users list --user-set-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List KeepTruckin users with --is-user-set-at filter"
xbe_json view keep-truckin-users list --is-user-set-at true --limit 5
assert_success

test_name "List KeepTruckin users with --assigned-at-min filter"
xbe_json view keep-truckin-users list --assigned-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List KeepTruckin users with --assigned-at-max filter"
xbe_json view keep-truckin-users list --assigned-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show KeepTruckin user"
if [[ $SKIP_SHOW -eq 0 && -n "$KEEP_TRUCKIN_USER_ID" && "$KEEP_TRUCKIN_USER_ID" != "null" ]]; then
    xbe_json view keep-truckin-users show "$KEEP_TRUCKIN_USER_ID"
    assert_success
else
    skip "No KeepTruckin user ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Find KeepTruckin user with user ID for update"
xbe_json view keep-truckin-users list --limit 20
if [[ $status -eq 0 ]]; then
    UPDATE_KEEP_TRUCKIN_USER_ID=$(json_get 'map(select(.user_id != null and .user_id != "")) | .[0].id')
    UPDATE_USER_ID=$(json_get 'map(select(.user_id != null and .user_id != "")) | .[0].user_id')
    if [[ -n "$UPDATE_KEEP_TRUCKIN_USER_ID" && "$UPDATE_KEEP_TRUCKIN_USER_ID" != "null" && -n "$UPDATE_USER_ID" && "$UPDATE_USER_ID" != "null" ]]; then
        pass
    else
        SKIP_UPDATE=1
        skip "No KeepTruckin user with linked user found"
    fi
else
    SKIP_UPDATE=1
    fail "Failed to list KeepTruckin users for update"
fi

test_name "Update KeepTruckin user assignment"
if [[ $SKIP_UPDATE -eq 0 ]]; then
    xbe_json do keep-truckin-users update "$UPDATE_KEEP_TRUCKIN_USER_ID" --user "$UPDATE_USER_ID"
    assert_success
else
    skip "No KeepTruckin user available for update"
fi

test_name "Update KeepTruckin user without fields fails"
if [[ $SKIP_UPDATE -eq 0 ]]; then
    xbe_json do keep-truckin-users update "$UPDATE_KEEP_TRUCKIN_USER_ID"
    assert_failure
else
    skip "No KeepTruckin user available for update"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
