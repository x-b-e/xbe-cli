#!/bin/bash
#
# XBE CLI Integration Tests: Tender Status Changes
#
# Tests list/show operations for tender-status-changes.
#
# COVERAGE: List filters (tender, status) + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TENDER_ID="${XBE_TEST_TENDER_STATUS_CHANGE_TENDER_ID:-}"
SAMPLE_ID=""

describe "Resource: tender-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender status changes"
xbe_json view tender-status-changes list --limit 5
assert_success

test_name "List tender status changes returns array"
xbe_json view tender-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tender status changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List tender status changes with --status filter"
xbe_json view tender-status-changes list --status accepted --limit 5
assert_success

test_name "List tender status changes with --tender filter"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
    xbe_json view tender-status-changes list --tender "$TENDER_ID" --limit 5
    assert_success
else
    skip "No tender ID available. Set XBE_TEST_TENDER_STATUS_CHANGE_TENDER_ID to enable tender filter testing."
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample tender status change"
xbe_json view tender-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No tender status changes available for show test"
    fi
else
    skip "Could not list tender status changes to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-status-changes show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show tender status change: $output"
        fi
    fi
else
    skip "No tender status change ID available for show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
