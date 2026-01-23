#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Approvals
#
# Tests list, show, and create operations for the time-sheet-approvals resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TIME_SHEET_ID=""
CREATE_TIME_SHEET_ID=""
LIST_SUPPORTED="true"

describe "Resource: time-sheet-approvals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet approvals"
xbe_json view time-sheet-approvals list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing time sheet approvals"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List time sheet approvals returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-approvals list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list time sheet approvals"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample time sheet approval"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-approvals list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No time sheet approvals available for follow-on tests"
        fi
    else
        skip "Could not list time sheet approvals to capture sample"
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
    xbe_json view labor-requirements list --limit 1
    if [[ $status -eq 0 ]]; then
        LABOR_REQUIREMENT_ID=$(json_get ".[0].id")
        if [[ -n "$LABOR_REQUIREMENT_ID" && "$LABOR_REQUIREMENT_ID" != "null" ]]; then
            xbe_json view labor-requirements show "$LABOR_REQUIREMENT_ID"
            if [[ $status -eq 0 ]]; then
                CREATE_TIME_SHEET_ID=$(json_get ".time_sheet_id")
            fi
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet approval"
if [[ -n "$CREATE_TIME_SHEET_ID" && "$CREATE_TIME_SHEET_ID" != "null" ]]; then
    xbe_json do time-sheet-approvals create \
        --time-sheet "$CREATE_TIME_SHEET_ID" \
        --comment "CLI test approval"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status must be valid"* ]] || \
           [[ "$output" == *"must have a duration"* ]] || \
           [[ "$output" == *"cost code allocation must be present"* ]] || \
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

test_name "Show time sheet approval"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-sheet-approvals show "$SAMPLE_ID"
    assert_success
else
    skip "No approval ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create approval without time sheet fails"
xbe_run do time-sheet-approvals create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
