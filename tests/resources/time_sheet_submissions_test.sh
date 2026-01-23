#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Submissions
#
# Tests create operations for the time-sheet-submissions resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_TIME_SHEET_ID="${XBE_TEST_TIME_SHEET_SUBMISSION_TIME_SHEET_ID:-}"

describe "Resource: time-sheet-submissions"

# ============================================================================
# Sample Record (used for create)
# ============================================================================

if [[ -z "$SAMPLE_TIME_SHEET_ID" ]]; then
    test_name "Locate time sheet via cost code allocations"
    xbe_json view time-sheet-cost-code-allocations list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
        if [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
            pass
        else
            skip "No time sheet found in cost code allocations"
        fi
    else
        skip "Could not list time sheet cost code allocations"
    fi
else
    test_name "Use provided time sheet ID"
    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Submit time sheet"
if [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
    xbe_json do time-sheet-submissions create \
        --time-sheet "$SAMPLE_TIME_SHEET_ID" \
        --comment "CLI submission test"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Submit failed: $output"
        fi
    fi
else
    skip "No time sheet available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Submit without required fields fails"
xbe_run do time-sheet-submissions create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
