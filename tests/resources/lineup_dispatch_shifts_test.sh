#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Dispatch Shifts
#
# Tests list/show for lineup-dispatch-shifts resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_LINEUP_DISPATCH_ID=""
SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID=""

describe "Resource: lineup-dispatch-shifts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup dispatch shifts"
xbe_json view lineup-dispatch-shifts list --limit 5
assert_success

test_name "List lineup dispatch shifts returns array"
xbe_json view lineup-dispatch-shifts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup dispatch shifts"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample lineup dispatch shift"
xbe_json view lineup-dispatch-shifts list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_LINEUP_DISPATCH_ID=$(json_get ".[0].lineup_dispatch_id")
    SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].lineup_job_schedule_shift_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No lineup dispatch shifts available for follow-on tests"
    fi
else
    skip "Could not list lineup dispatch shifts to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup dispatch shifts with --lineup-dispatch filter"
if [[ -n "$SAMPLE_LINEUP_DISPATCH_ID" && "$SAMPLE_LINEUP_DISPATCH_ID" != "null" ]]; then
    xbe_json view lineup-dispatch-shifts list --lineup-dispatch "$SAMPLE_LINEUP_DISPATCH_ID" --limit 5
    assert_success
else
    skip "No lineup dispatch ID available"
fi

test_name "List lineup dispatch shifts with --lineup-job-schedule-shift filter"
if [[ -n "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view lineup-dispatch-shifts list --lineup-job-schedule-shift "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No lineup job schedule shift ID available"
fi

test_name "List lineup dispatch shifts with --created-at-min filter"
xbe_json view lineup-dispatch-shifts list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatch shifts with --created-at-max filter"
xbe_json view lineup-dispatch-shifts list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatch shifts with --updated-at-min filter"
xbe_json view lineup-dispatch-shifts list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List lineup dispatch shifts with --updated-at-max filter"
xbe_json view lineup-dispatch-shifts list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup dispatch shift"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-dispatch-shifts show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup dispatch shift ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
