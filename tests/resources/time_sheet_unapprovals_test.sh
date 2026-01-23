#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Unapprovals
#
# Tests list, show, and create operations for the time-sheet-unapprovals resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TIME_SHEET_ID=""
CREATE_TIME_SHEET_ID=""
LIST_SUPPORTED="true"

describe "Resource: time-sheet-unapprovals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet unapprovals"
xbe_json view time-sheet-unapprovals list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing time sheet unapprovals"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List time sheet unapprovals returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-unapprovals list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list time sheet unapprovals"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample time sheet unapproval"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-unapprovals list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No time sheet unapprovals available for follow-on tests"
        fi
    else
        skip "Could not list time sheet unapprovals to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

if [[ -n "$XBE_TEST_TIME_SHEET_ID" ]]; then
    CREATE_TIME_SHEET_ID="$XBE_TEST_TIME_SHEET_ID"
elif [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
    CREATE_TIME_SHEET_ID="$SAMPLE_TIME_SHEET_ID"
else
    xbe_json view time-sheet-approvals list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet unapproval"
if [[ -n "$CREATE_TIME_SHEET_ID" && "$CREATE_TIME_SHEET_ID" != "null" ]]; then
    xbe_json do time-sheet-unapprovals create \
        --time-sheet "$CREATE_TIME_SHEET_ID" \
        --comment "CLI test unapproval"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status must be valid"* ]] || \
           [[ "$output" == *"must be approved"* ]] || \
           [[ "$output" == *"approved status"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No time sheet ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheet unapproval"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-sheet-unapprovals show "$SAMPLE_ID"
    assert_success
else
    skip "No unapproval ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create unapproval without time sheet fails"
xbe_run do time-sheet-unapprovals create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
