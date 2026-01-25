#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Trips Adjustments
#
# Tests CRUD operations for the driver-day-trips-adjustments resource.
# Requires a tender-job-schedule-shift ID; if unavailable, CRUD tests are skipped.
#
# COVERAGE: All list filters + create/update attributes (when prerequisites exist)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ADJUSTMENT_ID=""
CREATED_DRIVER_DAY_ID=""
CREATED_TRUCKER_ID=""
CREATED_BROKER_ID=""
CREATED_CREATED_BY_ID=""
TENDER_JOB_SCHEDULE_SHIFT_ID=""

describe "Resource: driver-day-trips-adjustments"

# ============================================================================
# Resolve a tender job schedule shift ID
# ============================================================================

test_name "Resolve tender job schedule shift ID"
if [[ -n "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
    TENDER_JOB_SCHEDULE_SHIFT_ID="$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID"
    echo "    Using XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID: $TENDER_JOB_SCHEDULE_SHIFT_ID"
    pass
else
    xbe_json view shift-feedbacks list --limit 1
    if [[ $status -eq 0 ]]; then
        TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
        if [[ -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" && "$TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
            echo "    Using tender-job-schedule-shift from shift feedbacks: $TENDER_JOB_SCHEDULE_SHIFT_ID"
            pass
        else
            skip "No tender-job-schedule-shift available; set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID to enable CRUD tests"
        fi
    else
        skip "Failed to list shift feedbacks; set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID to enable CRUD tests"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver day trips adjustments"
xbe_json view driver-day-trips-adjustments list --limit 5
assert_success

test_name "List driver day trips adjustments returns array"
xbe_json view driver-day-trips-adjustments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver day trips adjustments"
fi

# ============================================================================
# CREATE Tests - Error cases
# ============================================================================

test_name "Create adjustment requires --tender-job-schedule-shift"
xbe_run do driver-day-trips-adjustments create --old-trips-attributes '[{\"note\":\"missing shift\"}]'
assert_failure

test_name "Create adjustment requires --old-trips-attributes"
xbe_run do driver-day-trips-adjustments create --tender-job-schedule-shift "nonexistent"
assert_failure

# ============================================================================
# CREATE Tests - Success (if prerequisites available)
# ============================================================================

if [[ -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" && "$TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    test_name "Create driver day trips adjustment"
    OLD_TRIPS_JSON='[{"note":"cli-test old trip"}]'
    xbe_json do driver-day-trips-adjustments create \
        --tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --old-trips-attributes "$OLD_TRIPS_JSON" \
        --description "CLI test adjustment" \
        --status editing

    if [[ $status -eq 0 ]]; then
        CREATED_ADJUSTMENT_ID=$(json_get ".id")
        CREATED_DRIVER_DAY_ID=$(json_get ".driver_day_id")
        CREATED_TRUCKER_ID=$(json_get ".trucker_id")
        CREATED_BROKER_ID=$(json_get ".broker_id")
        CREATED_CREATED_BY_ID=$(json_get ".created_by_id")
        if [[ -n "$CREATED_ADJUSTMENT_ID" && "$CREATED_ADJUSTMENT_ID" != "null" ]]; then
            register_cleanup "driver-day-trips-adjustments" "$CREATED_ADJUSTMENT_ID"
            pass
        else
            fail "Created adjustment but no ID returned"
        fi
    else
        fail "Failed to create driver day trips adjustment"
    fi
else
    skip "No tender job schedule shift ID available; skipping create/update/delete tests"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$CREATED_ADJUSTMENT_ID" && "$CREATED_ADJUSTMENT_ID" != "null" ]]; then
    test_name "Show driver day trips adjustment"
    xbe_json view driver-day-trips-adjustments show "$CREATED_ADJUSTMENT_ID"
    assert_success
else
    skip "No adjustment ID available for show test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update adjustment without fields fails"
xbe_run do driver-day-trips-adjustments update "nonexistent"
assert_failure

if [[ -n "$CREATED_ADJUSTMENT_ID" && "$CREATED_ADJUSTMENT_ID" != "null" ]]; then
    test_name "Update driver day trips adjustment"
    NEW_TRIPS_JSON='[{"note":"cli-test new trip"}]'
    xbe_json do driver-day-trips-adjustments update "$CREATED_ADJUSTMENT_ID" \
        --description "Updated adjustment" \
        --status editing \
        --new-trips-attributes "$NEW_TRIPS_JSON"
    assert_success
else
    skip "No adjustment ID available for update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

FILTER_STATUS="editing"
FILTER_DRIVER_DAY_ID="${CREATED_DRIVER_DAY_ID:-${XBE_TEST_DRIVER_DAY_ID:-1}}"
FILTER_TENDER_JOB_SCHEDULE_SHIFT_ID="${TENDER_JOB_SCHEDULE_SHIFT_ID:-1}"
FILTER_TRUCKER_ID="${CREATED_TRUCKER_ID:-${XBE_TEST_TRUCKER_ID:-1}}"
FILTER_BROKER_ID="${CREATED_BROKER_ID:-${XBE_TEST_BROKER_ID:-1}}"
FILTER_CREATED_BY_ID="${CREATED_CREATED_BY_ID:-1}"

test_name "List adjustments with --status filter"
xbe_json view driver-day-trips-adjustments list --status "$FILTER_STATUS" --limit 10
assert_success

test_name "List adjustments with --driver-day filter"
xbe_json view driver-day-trips-adjustments list --driver-day "$FILTER_DRIVER_DAY_ID" --limit 10
assert_success

test_name "List adjustments with --tender-job-schedule-shift filter"
xbe_json view driver-day-trips-adjustments list --tender-job-schedule-shift "$FILTER_TENDER_JOB_SCHEDULE_SHIFT_ID" --limit 10
assert_success

test_name "List adjustments with --trucker filter"
xbe_json view driver-day-trips-adjustments list --trucker "$FILTER_TRUCKER_ID" --limit 10
assert_success

test_name "List adjustments with --broker filter"
xbe_json view driver-day-trips-adjustments list --broker "$FILTER_BROKER_ID" --limit 10
assert_success

test_name "List adjustments with --created-by filter"
xbe_json view driver-day-trips-adjustments list --created-by "$FILTER_CREATED_BY_ID" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete adjustment requires --confirm flag"
xbe_run do driver-day-trips-adjustments delete "nonexistent"
assert_failure

if [[ -n "$CREATED_ADJUSTMENT_ID" && "$CREATED_ADJUSTMENT_ID" != "null" ]]; then
    test_name "Delete driver day trips adjustment"
    xbe_json do driver-day-trips-adjustments delete "$CREATED_ADJUSTMENT_ID" --confirm
    assert_success
else
    skip "No adjustment ID available for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
