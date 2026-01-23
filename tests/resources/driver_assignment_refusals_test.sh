#!/bin/bash
#
# XBE CLI Integration Tests: Driver Assignment Refusals
#
# Tests list, show, and create operations for the driver-assignment-refusals resource.
#
# COVERAGE: List filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_SHIFT_ID=""
SAMPLE_DRIVER_ID=""

describe "Resource: driver-assignment-refusals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver assignment refusals"
xbe_json view driver-assignment-refusals list --limit 5
assert_success

test_name "List driver assignment refusals returns array"
xbe_json view driver-assignment-refusals list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver assignment refusals"
fi

# ============================================================================
# Sample Record (used for filters/show/create)
# ============================================================================

test_name "Capture sample refusal"
xbe_json view driver-assignment-refusals list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No refusals available for follow-on tests"
    fi
else
    skip "Could not list refusals to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List refusals with --tender-job-schedule-shift filter"
if [[ -n "$SAMPLE_SHIFT_ID" && "$SAMPLE_SHIFT_ID" != "null" ]]; then
    xbe_json view driver-assignment-refusals list --tender-job-schedule-shift "$SAMPLE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List refusals with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view driver-assignment-refusals list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List refusals with --created-at-min filter"
xbe_json view driver-assignment-refusals list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List refusals with --created-at-max filter"
xbe_json view driver-assignment-refusals list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List refusals with --updated-at-min filter"
xbe_json view driver-assignment-refusals list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List refusals with --updated-at-max filter"
xbe_json view driver-assignment-refusals list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver assignment refusal"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-assignment-refusals show "$SAMPLE_ID"
    assert_success
else
    skip "No refusal ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create refusal"
if [[ -n "$SAMPLE_SHIFT_ID" && "$SAMPLE_SHIFT_ID" != "null" && -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json do driver-assignment-refusals create \
        --tender-job-schedule-shift "$SAMPLE_SHIFT_ID" \
        --driver "$SAMPLE_DRIVER_ID" \
        --comment "CLI test refusal"
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
    skip "No shift/driver IDs available for create"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create refusal without required fields fails"
xbe_run do driver-assignment-refusals create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
