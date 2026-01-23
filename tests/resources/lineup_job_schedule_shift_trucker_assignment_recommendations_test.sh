#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Job Schedule Shift Trucker Assignment Recommendations
#
# Tests list, show, and create operations for
# lineup-job-schedule-shift-trucker-assignment-recommendations.
#
# COVERAGE: List filters + show + create
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
LINEUP_SHIFT_ID="${XBE_TEST_LINEUP_JOB_SCHEDULE_SHIFT_ID:-}"

describe "Resource: lineup-job-schedule-shift-trucker-assignment-recommendations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup job schedule shift trucker assignment recommendations"
xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -z "$LINEUP_SHIFT_ID" || "$LINEUP_SHIFT_ID" == "null" ]]; then
            LINEUP_SHIFT_ID=$(json_get ".[0].lineup_job_schedule_shift_id")
        fi
    fi
else
    fail "Failed to list recommendations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List recommendations with --lineup-job-schedule-shift filter"
if [[ -n "$LINEUP_SHIFT_ID" && "$LINEUP_SHIFT_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --lineup-job-schedule-shift "$LINEUP_SHIFT_ID" --limit 5
    assert_success
else
    skip "No lineup job schedule shift ID available"
fi

test_name "List recommendations with --created-at-min filter"
xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recommendations with --created-at-max filter"
xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recommendations with --updated-at-min filter"
xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recommendations with --updated-at-max filter"
xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup job schedule shift trucker assignment recommendation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shift-trucker-assignment-recommendations show "$SAMPLE_ID"
    assert_success
else
    skip "No recommendation ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create recommendation requires --lineup-job-schedule-shift"
xbe_run do lineup-job-schedule-shift-trucker-assignment-recommendations create
assert_failure

test_name "Create recommendation for lineup job schedule shift"
if [[ -n "$LINEUP_SHIFT_ID" && "$LINEUP_SHIFT_ID" != "null" ]]; then
    xbe_json do lineup-job-schedule-shift-trucker-assignment-recommendations create --lineup-job-schedule-shift "$LINEUP_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"no candidate truckers"* ]] || [[ "$output" == *"model file not found"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Unable to generate recommendations with available lineup job schedule shift"
        else
            fail "Failed to create recommendation"
        fi
    fi
else
    skip "No lineup job schedule shift ID available. Set XBE_TEST_LINEUP_JOB_SCHEDULE_SHIFT_ID to enable create testing."
fi

run_tests
