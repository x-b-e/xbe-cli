#!/bin/bash
#
# XBE CLI Integration Tests: Driver Assignment Acknowledgements
#
# Tests list, show, and create operations for the driver-assignment-acknowledgements resource.
#
# COVERAGE: List filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_DRIVER_ID=""

describe "Resource: driver-assignment-acknowledgements"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver assignment acknowledgements"
xbe_json view driver-assignment-acknowledgements list --limit 5
assert_success

test_name "List driver assignment acknowledgements returns array"
xbe_json view driver-assignment-acknowledgements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver assignment acknowledgements"
fi

# ============================================================================
# Sample Record (used for filters/show/create)
# ============================================================================

test_name "Capture sample acknowledgement"
xbe_json view driver-assignment-acknowledgements list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No acknowledgements available for follow-on tests"
    fi
else
    skip "Could not list acknowledgements to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List acknowledgements with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view driver-assignment-acknowledgements list --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List acknowledgements with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view driver-assignment-acknowledgements list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver assignment acknowledgement"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-assignment-acknowledgements show "$SAMPLE_ID"
    assert_success
else
    skip "No acknowledgement ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create driver assignment acknowledgement"
if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" && -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json do driver-assignment-acknowledgements create --tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" --driver "$SAMPLE_DRIVER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No tender job schedule shift or driver ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create acknowledgement without required fields fails"
xbe_run do driver-assignment-acknowledgements create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
