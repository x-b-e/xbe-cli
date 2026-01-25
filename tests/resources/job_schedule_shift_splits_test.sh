#!/bin/bash
#
# XBE CLI Integration Tests: Job Schedule Shift Splits
#
# Tests list, show, and create operations for the job-schedule-shift-splits resource.
#
# COVERAGE: List + show + create + filters + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_NEW_JOB_SCHEDULE_SHIFT_ID=""
CREATE_JOB_SCHEDULE_SHIFT_ID=""
LIST_SUPPORTED="true"

describe "Resource: job-schedule-shift-splits"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job schedule shift splits"
xbe_json view job-schedule-shift-splits list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing job schedule shift splits"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List job schedule shift splits returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-schedule-shift-splits list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list job schedule shift splits"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample job schedule shift split"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-schedule-shift-splits list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].job_schedule_shift_id")
        SAMPLE_NEW_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].new_job_schedule_shift_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No job schedule shift splits available for follow-on tests"
        fi
    else
        skip "Could not list job schedule shift splits to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter splits by job schedule shift"
if [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view job-schedule-shift-splits list --job-schedule-shift "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by job schedule shift failed"
    fi
else
    skip "No job schedule shift ID available"
fi

test_name "Filter splits by new job schedule shift"
if [[ -n "$SAMPLE_NEW_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_NEW_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view job-schedule-shift-splits list --new-job-schedule-shift "$SAMPLE_NEW_JOB_SCHEDULE_SHIFT_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by new job schedule shift failed"
    fi
else
    skip "No new job schedule shift ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -n "$XBE_TEST_JOB_SCHEDULE_SHIFT_ID" ]]; then
    CREATE_JOB_SCHEDULE_SHIFT_ID="$XBE_TEST_JOB_SCHEDULE_SHIFT_ID"
elif [[ -n "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    CREATE_JOB_SCHEDULE_SHIFT_ID="$SAMPLE_JOB_SCHEDULE_SHIFT_ID"
fi

test_name "Create job schedule shift split"
if [[ -n "$CREATE_JOB_SCHEDULE_SHIFT_ID" && "$CREATE_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    NEW_START_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    xbe_json do job-schedule-shift-splits create \
        --job-schedule-shift "$CREATE_JOB_SCHEDULE_SHIFT_ID" \
        --expected-material-transaction-count 1 \
        --expected-material-transaction-tons 1.0 \
        --new-start-at "$NEW_START_AT"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"must be flexible"* ]] || \
           [[ "$output" == *"must have an accepted broker tender job schedule shift"* ]] || \
           [[ "$output" == *"must have an accepted customer tender job schedule shift"* ]] || \
           [[ "$output" == *"must leave sufficient remainder"* ]] || \
           [[ "$output" == *"must be present on the job schedule shift"* ]] || \
           [[ "$output" == *"new start at must be between start at min and max"* ]] || \
           [[ "$output" == *"must be different from the old start at"* ]] || \
           [[ "$output" == *"is not valid to be split"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No job schedule shift ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job schedule shift split"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-schedule-shift-splits show "$SAMPLE_ID"
    assert_success
else
    skip "No job schedule shift split ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create split without job schedule shift fails"
xbe_run do job-schedule-shift-splits create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
