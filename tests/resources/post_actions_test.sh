#!/bin/bash
#
# XBE CLI Integration Tests: Post Actions
#
# Tests view operations for post actions.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FIRST_ACTION_ID=""
ACTION_TOKEN=""

describe "Resource: post-actions (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List post actions"
xbe_json view post-actions list --limit 5
assert_success

test_name "List post actions returns array"
xbe_json view post-actions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list post actions"
fi

# Capture IDs for downstream tests
xbe_json view post-actions list --limit 5
if [[ $status -eq 0 ]]; then
    FIRST_ACTION_ID=$(json_get ".[0].id")
    ACTION_TOKEN=$(json_get ".[0].token")
else
    FIRST_ACTION_ID=""
    ACTION_TOKEN=""
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show post action"
if [[ -n "$FIRST_ACTION_ID" && "$FIRST_ACTION_ID" != "null" ]]; then
    xbe_json view post-actions show "$FIRST_ACTION_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Failed to show post action"
    fi
else
    skip "No post action ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List post actions with --action-token filter"
if [[ -n "$ACTION_TOKEN" && "$ACTION_TOKEN" != "null" ]]; then
    xbe_json view post-actions list --action-token "$ACTION_TOKEN" --limit 5
    assert_success
else
    skip "No post action token available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
