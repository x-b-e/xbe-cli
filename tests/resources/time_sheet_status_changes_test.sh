#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Status Changes
#
# Tests view operations for the time_sheet_status_changes resource.
# Time sheet status changes record status transitions with timestamps and comments.
#
# COVERAGE: List + show + filters (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: time-sheet-status-changes (view-only)"

SAMPLE_ID=""
SAMPLE_TIME_SHEET_ID=""
SAMPLE_STATUS=""

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet status changes"
xbe_json view time-sheet-status-changes list --limit 5
assert_success

test_name "List time sheet status changes returns array"
xbe_json view time-sheet-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time sheet status changes"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample time sheet status change"
xbe_json view time-sheet-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No time sheet status changes available for follow-on tests"
    fi
else
    skip "Could not list time sheet status changes to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List time sheet status changes with --time-sheet filter"
if [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
    xbe_json view time-sheet-status-changes list --time-sheet "$SAMPLE_TIME_SHEET_ID" --limit 5
    assert_success
else
    skip "No sample time sheet ID available"
fi

test_name "List time sheet status changes with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view time-sheet-status-changes list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No sample status available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheet status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-sheet-status-changes show "$SAMPLE_ID"
    assert_success
else
    skip "No time sheet status change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
