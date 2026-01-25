#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Planned Event Time Schedules
#
# Tests view and create behavior for project_transport_plan_planned_event_time_schedules.
#
# COVERAGE: List + list filters + show + create + required flag failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-planned-event-time-schedules"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""
SAMPLE_ORDER_ID=""
SAMPLE_SUCCESS=""

PLAN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
TRANSPORT_ORDER_ID="${XBE_TEST_TRANSPORT_ORDER_ID:-}"
PLAN_DATA_JSON="${XBE_TEST_PLAN_DATA_JSON:-}"

CREATED_ID=""

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan planned event time schedules"
xbe_json view project-transport-plan-planned-event-time-schedules list --limit 5
assert_success

test_name "List project transport plan planned event time schedules returns array"
xbe_json view project-transport-plan-planned-event-time-schedules list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan planned event time schedules"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample schedule"
xbe_json view project-transport-plan-planned-event-time-schedules list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
    SAMPLE_ORDER_ID=$(json_get ".[0].transport_order_id")
    SAMPLE_SUCCESS=$(json_get ".[0].success")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No schedules available for follow-on tests"
    fi
else
    skip "Could not list schedules to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List schedules with --project-transport-plan filter"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-planned-event-time-schedules list \
        --project-transport-plan "$SAMPLE_PLAN_ID" \
        --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List schedules with --transport-order filter"
if [[ -n "$SAMPLE_ORDER_ID" && "$SAMPLE_ORDER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-planned-event-time-schedules list \
        --transport-order "$SAMPLE_ORDER_ID" \
        --limit 5
    assert_success
else
    skip "No transport order ID available"
fi

test_name "List schedules with --success filter"
if [[ -n "$SAMPLE_SUCCESS" && "$SAMPLE_SUCCESS" != "null" ]]; then
    xbe_json view project-transport-plan-planned-event-time-schedules list \
        --success "$SAMPLE_SUCCESS" \
        --limit 5
    assert_success
else
    skip "No success value available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show schedule details"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-planned-event-time-schedules show "$SAMPLE_ID"
    assert_success
else
    skip "No schedule ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create schedule without required flags fails"
xbe_run do project-transport-plan-planned-event-time-schedules create
assert_failure

if [[ -n "$PLAN_ID" ]]; then
    test_name "Create schedule using project transport plan"
    xbe_json do project-transport-plan-planned-event-time-schedules create \
        --project-transport-plan "$PLAN_ID" \
        --respect-provided-event-times

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Created schedule but no ID returned"
        fi
    else
        fail "Failed to create schedule"
    fi
elif [[ -n "$TRANSPORT_ORDER_ID" && -n "$PLAN_DATA_JSON" ]]; then
    test_name "Create schedule using transport order and plan data"
    xbe_json do project-transport-plan-planned-event-time-schedules create \
        --transport-order "$TRANSPORT_ORDER_ID" \
        --plan-data "$PLAN_DATA_JSON" \
        --respect-provided-event-times

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Created schedule but no ID returned"
        fi
    else
        fail "Failed to create schedule"
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID or XBE_TEST_TRANSPORT_ORDER_ID with XBE_TEST_PLAN_DATA_JSON to run create tests"
fi

# ============================================================================
# SHOW Created
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Show created schedule"
    xbe_json view project-transport-plan-planned-event-time-schedules show "$CREATED_ID"
    assert_success
else
    skip "No schedule created; skipping created schedule show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
