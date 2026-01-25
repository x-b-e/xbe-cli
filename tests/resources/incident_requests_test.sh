#!/bin/bash
#
# XBE CLI Integration Tests: Incident Requests
#
# Tests CRUD operations and list filters for the incident-requests resource.
#
# COVERAGE: Create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REQUEST_ID=""
CURRENT_USER_ID=""
SHIFT_ID=""

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_TIME_VALUE_TYPE=""
SAMPLE_START_AT=""
SAMPLE_END_AT=""
SAMPLE_SHIFT_ID=""
SAMPLE_ASSIGNEE_ID=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_CUSTOMER_ID=""
SAMPLE_BROKER_ID=""

START_AT="2025-01-01T08:00:00Z"
END_AT="2025-01-01T09:00:00Z"
UPDATED_START_AT="2025-01-02T08:00:00Z"
UPDATED_END_AT="2025-01-02T09:00:00Z"

describe "Resource: incident-requests"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List incident requests"
xbe_json view incident-requests list --limit 5
assert_success

test_name "List incident requests returns array"
xbe_json view incident-requests list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list incident requests"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample incident request"
xbe_json view incident-requests list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_TIME_VALUE_TYPE=$(json_get ".[0].time_value_type")
    SAMPLE_START_AT=$(json_get ".[0].start_at")
    SAMPLE_END_AT=$(json_get ".[0].end_at")
    SAMPLE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
    SAMPLE_ASSIGNEE_ID=$(json_get ".[0].assignee_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    SAMPLE_CUSTOMER_ID=$(json_get ".[0].customer_id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No incident requests available for sample-based tests"
    fi
else
    skip "Could not list incident requests to capture sample"
fi

# ============================================================================
# Prerequisites - tender job schedule shift and current user
# ============================================================================

test_name "Capture tender job schedule shift for create tests"
SHIFT_ID="$SAMPLE_SHIFT_ID"
if [[ -z "$SHIFT_ID" || "$SHIFT_ID" == "null" ]]; then
    xbe_json view tender-job-schedule-shifts list --limit 1
    if [[ $status -eq 0 ]]; then
        SHIFT_ID=$(json_get ".[0].id")
    fi
fi
if [[ -z "$SHIFT_ID" || "$SHIFT_ID" == "null" ]]; then
    if [[ -n "$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID" ]]; then
        SHIFT_ID="$XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID"
        echo "    Using XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID: $SHIFT_ID"
        pass
    else
        skip "No tender job schedule shift available"
    fi
else
    pass
fi

test_name "Fetch current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned"
    fi
else
    fail "Failed to fetch current user"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create incident request requires --start-at"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    xbe_json do incident-requests create --tender-job-schedule-shift "$SHIFT_ID"
    assert_failure
else
    skip "No tender job schedule shift ID available"
fi

test_name "Create incident request requires --tender-job-schedule-shift"
xbe_json do incident-requests create --start-at "$START_AT"
assert_failure

test_name "Create incident request"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    xbe_json do incident-requests create \
        --start-at "$START_AT" \
        --end-at "$END_AT" \
        --description "Incident request test" \
        --time-value-type credited_time \
        --is-down-time \
        --assignee "$CURRENT_USER_ID" \
        --created-by "$CURRENT_USER_ID" \
        --tender-job-schedule-shift "$SHIFT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_REQUEST_ID=$(json_get ".id")
        if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
            register_cleanup "incident-requests" "$CREATED_REQUEST_ID"
            pass
        else
            fail "Created incident request but no ID returned"
        fi
    else
        fail "Failed to create incident request"
    fi
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update incident request attributes"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do incident-requests update "$CREATED_REQUEST_ID" \
        --start-at "$UPDATED_START_AT" \
        --end-at "$UPDATED_END_AT" \
        --description "Incident request updated" \
        --time-value-type deducted_time \
        --is-down-time=false
    assert_success
else
    skip "No incident request ID available"
fi

test_name "Update incident request without any fields fails"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do incident-requests update "$CREATED_REQUEST_ID"
    assert_failure
else
    skip "No incident request ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show incident request"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json view incident-requests show "$CREATED_REQUEST_ID"
    assert_success
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view incident-requests show "$SAMPLE_ID"
    assert_success
else
    skip "No incident request ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter: --tender-job-schedule-shift"
FILTER_SHIFT_ID="${SHIFT_ID:-$SAMPLE_SHIFT_ID}"
if [[ -n "$FILTER_SHIFT_ID" && "$FILTER_SHIFT_ID" != "null" ]]; then
    xbe_json view incident-requests list --tender-job-schedule-shift "$FILTER_SHIFT_ID" --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

test_name "Filter: --broker"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view incident-requests list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter: --customer"
if [[ -n "$SAMPLE_CUSTOMER_ID" && "$SAMPLE_CUSTOMER_ID" != "null" ]]; then
    xbe_json view incident-requests list --customer "$SAMPLE_CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "Filter: --assignee"
if [[ -n "$SAMPLE_ASSIGNEE_ID" && "$SAMPLE_ASSIGNEE_ID" != "null" ]]; then
    xbe_json view incident-requests list --assignee "$SAMPLE_ASSIGNEE_ID" --limit 5
    assert_success
else
    skip "No assignee ID available"
fi

test_name "Filter: --created-by"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view incident-requests list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "Filter: --start-at-min"
xbe_json view incident-requests list --start-at-min "${SAMPLE_START_AT:-$START_AT}" --limit 5
assert_success

test_name "Filter: --start-at-max"
xbe_json view incident-requests list --start-at-max "${SAMPLE_START_AT:-$START_AT}" --limit 5
assert_success

test_name "Filter: --end-at-min"
xbe_json view incident-requests list --end-at-min "${SAMPLE_END_AT:-$END_AT}" --limit 5
assert_success

test_name "Filter: --end-at-max"
xbe_json view incident-requests list --end-at-max "${SAMPLE_END_AT:-$END_AT}" --limit 5
assert_success

test_name "Filter: --status"
xbe_json view incident-requests list --status "${SAMPLE_STATUS:-submitted}" --limit 5
assert_success

test_name "Filter: --time-value-type"
xbe_json view incident-requests list --time-value-type "${SAMPLE_TIME_VALUE_TYPE:-credited_time}" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete incident request requires --confirm flag"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do incident-requests delete "$CREATED_REQUEST_ID"
    assert_failure
else
    skip "No incident request ID available"
fi

test_name "Delete incident request with --confirm"
if [[ -n "$CREATED_REQUEST_ID" && "$CREATED_REQUEST_ID" != "null" ]]; then
    xbe_json do incident-requests delete "$CREATED_REQUEST_ID" --confirm
    assert_success
else
    skip "No incident request ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
