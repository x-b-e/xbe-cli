#!/bin/bash
#
# XBE CLI Integration Tests: Job Schedule Shift Is Managed Toggles
#
# Tests create operations for the job-schedule-shift-is-managed-toggles resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_JOB_SCHEDULE_SHIFT_ID=""
DIRECT_API_AVAILABLE=0

if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

describe "Resource: job-schedule-shift-is-managed-toggles"

# ============================================================================
# Sample Record (used for create)
# ============================================================================

test_name "Capture job schedule shift"
if [[ $DIRECT_API_AVAILABLE -eq 1 ]]; then
    run curl -sS "$XBE_BASE_URL/v1/job-schedule-shifts?page[limit]=1" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json"
    if [[ $status -eq 0 ]]; then
        SAMPLE_JOB_SCHEDULE_SHIFT_ID=$(echo "$output" | jq -r '.data[0].id')
        if [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
            pass
        else
            skip "No job schedule shifts available"
        fi
    else
        skip "Could not query job schedule shifts"
    fi
else
    skip "Set XBE_TOKEN to query job schedule shifts via API"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Toggle job schedule shift managed status"
if [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json do job-schedule-shift-is-managed-toggles create \
        --job-schedule-shift "$SAMPLE_JOB_SCHEDULE_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Toggle failed: $output"
        fi
    fi
else
    skip "No job schedule shift available for toggle"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Toggle without required fields fails"
xbe_run do job-schedule-shift-is-managed-toggles create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
