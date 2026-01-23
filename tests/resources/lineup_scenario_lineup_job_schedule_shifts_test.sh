#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Lineup Job Schedule Shifts
#
# Tests CRUD operations for the lineup_scenario_lineup_job_schedule_shifts resource.
#
# COVERAGE: All filters + all create attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_LSLJSS_ID=""
SAMPLE_LINEUP_SCENARIO_ID=""
SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID=""
CREATED_LSLJSS_ID=""

describe "Resource: lineup-scenario-lineup-job-schedule-shifts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenario lineup job schedule shifts"
xbe_json view lineup-scenario-lineup-job-schedule-shifts list --limit 5
assert_success

test_name "List lineup scenario lineup job schedule shifts returns array"
xbe_json view lineup-scenario-lineup-job-schedule-shifts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list lineup scenario lineup job schedule shifts"
fi

# ============================================================================
# Prerequisites - Locate sample lineup scenario lineup job schedule shift
# ============================================================================

test_name "Locate lineup scenario lineup job schedule shift for filters"
xbe_json view lineup-scenario-lineup-job-schedule-shifts list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_LSLJSS_ID=$(json_get ".[0].id")
        SAMPLE_LINEUP_SCENARIO_ID=$(json_get ".[0].lineup_scenario_id")
        SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].lineup_job_schedule_shift_id")
        pass
    else
        if [[ -n "$XBE_TEST_LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" ]]; then
            xbe_json view lineup-scenario-lineup-job-schedule-shifts show "$XBE_TEST_LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_LSLJSS_ID=$(json_get ".id")
                SAMPLE_LINEUP_SCENARIO_ID=$(json_get ".lineup_scenario_id")
                SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID=$(json_get ".lineup_job_schedule_shift_id")
                pass
            else
                skip "Failed to load XBE_TEST_LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID"
            fi
        else
            skip "No lineup scenario lineup job schedule shifts found. Set XBE_TEST_LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID to enable filter tests."
        fi
    fi
else
    fail "Failed to list lineup scenario lineup job schedule shifts for prerequisites"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_LSLJSS_ID" && "$SAMPLE_LSLJSS_ID" != "null" ]]; then
    test_name "Show lineup scenario lineup job schedule shift"
    xbe_json view lineup-scenario-lineup-job-schedule-shifts show "$SAMPLE_LSLJSS_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show lineup scenario lineup job schedule shift"
    fi
else
    test_name "Show lineup scenario lineup job schedule shift"
    skip "No lineup scenario lineup job schedule shift available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter lineup scenario lineup job schedule shifts by lineup scenario"
if [[ -n "$SAMPLE_LINEUP_SCENARIO_ID" && "$SAMPLE_LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json view lineup-scenario-lineup-job-schedule-shifts list --lineup-scenario "$SAMPLE_LINEUP_SCENARIO_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LSLJSS_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario lineup job schedule shift"
        fi
    else
        fail "Failed to filter by lineup scenario"
    fi
else
    skip "No lineup scenario ID available for filter test"
fi

test_name "Filter lineup scenario lineup job schedule shifts by lineup job schedule shift"
if [[ -n "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view lineup-scenario-lineup-job-schedule-shifts list --lineup-job-schedule-shift "$SAMPLE_LINEUP_JOB_SCHEDULE_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$SAMPLE_LSLJSS_ID" '.[] | select(.id == $id)' > /dev/null; then
            pass
        else
            fail "Filtered list missing sample lineup scenario lineup job schedule shift"
        fi
    else
        fail "Failed to filter by lineup job schedule shift"
    fi
else
    skip "No lineup job schedule shift ID available for filter test"
fi

# ============================================================================
# CREATE Tests - Best Effort
# ============================================================================

test_name "Create lineup scenario lineup job schedule shift"
if [[ -n "$XBE_TEST_LINEUP_SCENARIO_ID" && -n "$XBE_TEST_LINEUP_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json do lineup-scenario-lineup-job-schedule-shifts create \
        --lineup-scenario "$XBE_TEST_LINEUP_SCENARIO_ID" \
        --lineup-job-schedule-shift "$XBE_TEST_LINEUP_JOB_SCHEDULE_SHIFT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_LSLJSS_ID=$(json_get ".id")
        if [[ -n "$CREATED_LSLJSS_ID" && "$CREATED_LSLJSS_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-lineup-job-schedule-shifts" "$CREATED_LSLJSS_ID"
            pass
        else
            fail "Created lineup scenario lineup job schedule shift but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario lineup job schedule shift"
    fi
else
    skip "Set XBE_TEST_LINEUP_SCENARIO_ID and XBE_TEST_LINEUP_JOB_SCHEDULE_SHIFT_ID to enable create test"
fi

# ============================================================================
# DELETE Tests - Best Effort
# ============================================================================

test_name "Delete lineup scenario lineup job schedule shift"
if [[ -n "$CREATED_LSLJSS_ID" && "$CREATED_LSLJSS_ID" != "null" ]]; then
    xbe_run do lineup-scenario-lineup-job-schedule-shifts delete "$CREATED_LSLJSS_ID" --confirm
    assert_success
else
    skip "No created lineup scenario lineup job schedule shift to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required fields fails"
xbe_run do lineup-scenario-lineup-job-schedule-shifts create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
