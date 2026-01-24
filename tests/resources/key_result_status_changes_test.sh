#!/bin/bash
#
# XBE CLI Integration Tests: Key Result Status Changes
#
# Tests list/show operations for key-result-status-changes.
#
# COVERAGE: List filters (key-result, status) + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

KEY_RESULT_ID="${XBE_TEST_KEY_RESULT_STATUS_CHANGE_KEY_RESULT_ID:-}"
SAMPLE_ID=""

describe "Resource: key-result-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List key result status changes"
xbe_json view key-result-status-changes list --limit 5
assert_success

test_name "List key result status changes returns array"
xbe_json view key-result-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list key result status changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List key result status changes with --status filter"
xbe_json view key-result-status-changes list --status green --limit 5
assert_success

test_name "List key result status changes with --key-result filter"
if [[ -n "$KEY_RESULT_ID" && "$KEY_RESULT_ID" != "null" ]]; then
    xbe_json view key-result-status-changes list --key-result "$KEY_RESULT_ID" --limit 5
    assert_success
else
    skip "No key result ID available. Set XBE_TEST_KEY_RESULT_STATUS_CHANGE_KEY_RESULT_ID to enable key result filter testing."
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample key result status change"
xbe_json view key-result-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No key result status changes available for show test"
    fi
else
    skip "Could not list key result status changes to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show key result status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view key-result-status-changes show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show key result status change: $output"
        fi
    fi
else
    skip "No key result status change ID available for show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
