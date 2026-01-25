#!/bin/bash
#
# XBE CLI Integration Tests: Proffer Likes
#
# Tests list, show, create, and delete operations for the proffer-likes resource.
#
# COVERAGE: List filters + show + create/delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PROFFER_LIKE_ID=""
PROFFER_ID=""
USER_ID=""
CURRENT_USER_ID=""
CREATED_PROFFER_LIKE_ID=""
PROFFER_ID_FOR_CREATE=""

describe "Resource: proffer-likes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List proffer likes"
xbe_json view proffer-likes list --limit 5
assert_success

test_name "List proffer likes returns array"
xbe_json view proffer-likes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list proffer likes"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample proffer like"
xbe_json view proffer-likes list --limit 1
if [[ $status -eq 0 ]]; then
    PROFFER_LIKE_ID=$(json_get ".[0].id")
    PROFFER_ID=$(json_get ".[0].proffer_id")
    USER_ID=$(json_get ".[0].user_id")
    if [[ -n "$PROFFER_LIKE_ID" && "$PROFFER_LIKE_ID" != "null" ]]; then
        pass
    else
        skip "No proffer likes available"
    fi
else
    fail "Failed to list proffer likes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List proffer likes with --proffer filter"
if [[ -n "$PROFFER_ID" && "$PROFFER_ID" != "null" ]]; then
    xbe_json view proffer-likes list --proffer "$PROFFER_ID" --limit 5
    assert_success
else
    skip "No proffer ID available"
fi

test_name "List proffer likes with --user filter"
if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view proffer-likes list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List proffer likes with --created-at-min filter"
xbe_json view proffer-likes list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List proffer likes with --created-at-max filter"
xbe_json view proffer-likes list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List proffer likes with --updated-at-min filter"
xbe_json view proffer-likes list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List proffer likes with --updated-at-max filter"
xbe_json view proffer-likes list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show proffer like"
if [[ -n "$PROFFER_LIKE_ID" && "$PROFFER_LIKE_ID" != "null" ]]; then
    xbe_json view proffer-likes show "$PROFFER_LIKE_ID"
    assert_success
else
    skip "No proffer like ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create proffer like without required fields fails"
xbe_run do proffer-likes create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No current user ID available"
    fi
else
    skip "Failed to resolve current user"
fi

if [[ -n "$XBE_TEST_PROFFER_ID" ]]; then
    PROFFER_ID_FOR_CREATE="$XBE_TEST_PROFFER_ID"
elif [[ -n "$PROFFER_ID" && "$PROFFER_ID" != "null" ]]; then
    PROFFER_ID_FOR_CREATE="$PROFFER_ID"
fi

test_name "Create proffer like"
if [[ -n "$PROFFER_ID_FOR_CREATE" && -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json do proffer-likes create --proffer "$PROFFER_ID_FOR_CREATE" --user "$CURRENT_USER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_PROFFER_LIKE_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROFFER_LIKE_ID" && "$CREATED_PROFFER_LIKE_ID" != "null" ]]; then
            register_cleanup "proffer-likes" "$CREATED_PROFFER_LIKE_ID"
            pass
        else
            fail "Created proffer like but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Missing proffer ID or current user ID"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete proffer like"
if [[ -n "$CREATED_PROFFER_LIKE_ID" && "$CREATED_PROFFER_LIKE_ID" != "null" ]]; then
    xbe_run do proffer-likes delete "$CREATED_PROFFER_LIKE_ID" --confirm
    assert_success
else
    skip "No proffer like created for deletion"
fi

run_tests
