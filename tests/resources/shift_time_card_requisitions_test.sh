#!/bin/bash
#
# XBE CLI Integration Tests: Shift Time Card Requisitions
#
# Tests list/show/create operations for shift-time-card-requisitions.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_CREATED_BY_ID=""

describe "Resource: shift-time-card-requisitions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List shift time card requisitions"
xbe_json view shift-time-card-requisitions list --limit 5
assert_success

test_name "List shift time card requisitions returns array"
xbe_json view shift-time-card-requisitions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list shift time card requisitions"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample requisition"
xbe_json view shift-time-card-requisitions list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No requisitions available for follow-on tests"
    fi
else
    skip "Could not list requisitions to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List requisitions with --tender-job-schedule-shift filter"
TARGET_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
if [[ -z "$TARGET_SHIFT_ID" && -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    TARGET_SHIFT_ID="$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID"
fi
if [[ -n "$TARGET_SHIFT_ID" ]]; then
    xbe_json view shift-time-card-requisitions list --tender-job-schedule-shift "$TARGET_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "List requisitions with --time-card filter"
TARGET_TIME_CARD_ID="${XBE_TEST_TIME_CARD_ID:-}"
if [[ -z "$TARGET_TIME_CARD_ID" && -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    TARGET_TIME_CARD_ID="$SAMPLE_TIME_CARD_ID"
fi
if [[ -n "$TARGET_TIME_CARD_ID" ]]; then
    xbe_json view shift-time-card-requisitions list --time-card "$TARGET_TIME_CARD_ID" --limit 5
    assert_success
else
    skip "No time card ID available"
fi

test_name "List requisitions with --broker filter"
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
if [[ -z "$BROKER_ID" ]]; then
    xbe_json view brokers list --limit 1
    if [[ $status -eq 0 ]]; then
        BROKER_ID=$(json_get ".[0].id")
    fi
fi
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List requisitions with --trucker filter"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
if [[ -z "$TRUCKER_ID" ]]; then
    xbe_json view truckers list --limit 1
    if [[ $status -eq 0 ]]; then
        TRUCKER_ID=$(json_get ".[0].id")
    fi
fi
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List requisitions with --job-production-plan filter"
JPP_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_ID:-}"
if [[ -z "$JPP_ID" ]]; then
    xbe_json view job-production-plans list --limit 1
    if [[ $status -eq 0 ]]; then
        JPP_ID=$(json_get ".[0].id")
    fi
fi
if [[ -n "$JPP_ID" && "$JPP_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions list --job-production-plan "$JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List requisitions with --driver filter"
DRIVER_ID="${XBE_TEST_DRIVER_ID:-}"
if [[ -z "$DRIVER_ID" ]]; then
    xbe_json view users list --limit 1
    if [[ $status -eq 0 ]]; then
        DRIVER_ID=$(json_get ".[0].id")
    fi
fi
if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions list --driver "$DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List requisitions with --created-by filter"
CREATED_BY_ID="${XBE_TEST_CREATED_BY_ID:-}"
if [[ -z "$CREATED_BY_ID" && -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    CREATED_BY_ID="$SAMPLE_CREATED_BY_ID"
fi
if [[ -z "$CREATED_BY_ID" ]]; then
    xbe_json view users list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".[0].id")
    fi
fi
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List requisitions with --status=open filter"
xbe_json view shift-time-card-requisitions list --status open --limit 5
assert_success

test_name "List requisitions with --status=closed filter"
xbe_json view shift-time-card-requisitions list --status closed --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show shift time card requisition"
SHOW_ID="${XBE_TEST_SHIFT_TIME_CARD_REQUISITION_ID:-$SAMPLE_ID}"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view shift-time-card-requisitions show "$SHOW_ID"
    assert_success
else
    skip "No requisition ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requisition requires tender job schedule shift"
xbe_run do shift-time-card-requisitions create
assert_failure

test_name "Create shift time card requisition with invalid-when-hourly override"
TARGET_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID}"
if [[ -n "$TARGET_SHIFT_ID" && "$TARGET_SHIFT_ID" != "null" ]]; then
    xbe_json do shift-time-card-requisitions create \
        --tender-job-schedule-shift "$TARGET_SHIFT_ID" \
        --invalid-when-hourly=false
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create shift time card requisition: $output"
        fi
    fi
else
    skip "Missing tender job schedule shift ID (set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID)"
fi

run_tests
