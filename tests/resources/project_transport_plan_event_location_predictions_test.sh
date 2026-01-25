#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Event Location Predictions
#
# Tests view/create/delete behavior for project_transport_plan_event_location_predictions.
#
# COVERAGE: List + list filters + show + create + delete guard + required flag failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-event-location-predictions"

SAMPLE_ID=""
SAMPLE_EVENT_ID=""
SAMPLE_TRANSPORT_ORDER_ID=""

CREATED_ID=""

EVENT_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_EVENT_ID:-}"
TRANSPORT_ORDER_ID="${XBE_TEST_TRANSPORT_ORDER_ID:-}"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan event location predictions"
xbe_json view project-transport-plan-event-location-predictions list --limit 5
assert_success

test_name "List project transport plan event location predictions returns array"
xbe_json view project-transport-plan-event-location-predictions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan event location predictions"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample prediction"
xbe_json view project-transport-plan-event-location-predictions list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_EVENT_ID=$(json_get ".[0].project_transport_plan_event_id")
    SAMPLE_TRANSPORT_ORDER_ID=$(json_get ".[0].transport_order_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No predictions available for follow-on tests"
    fi
else
    skip "Could not list predictions to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List predictions with --project-transport-plan-event filter"
if [[ -n "$SAMPLE_EVENT_ID" && "$SAMPLE_EVENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-location-predictions list \
        --project-transport-plan-event "$SAMPLE_EVENT_ID" \
        --limit 5
    assert_success
else
    skip "No project transport plan event ID available"
fi

test_name "List predictions with --transport-order filter"
if [[ -n "$SAMPLE_TRANSPORT_ORDER_ID" && "$SAMPLE_TRANSPORT_ORDER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-location-predictions list \
        --transport-order "$SAMPLE_TRANSPORT_ORDER_ID" \
        --limit 5
    assert_success
else
    skip "No transport order ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction details"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-event-location-predictions show "$SAMPLE_ID"
    assert_success
else
    skip "No prediction ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction without required flags fails"
xbe_run do project-transport-plan-event-location-predictions create
assert_failure

if [[ -n "$EVENT_ID" && -n "$TRANSPORT_ORDER_ID" ]]; then
    test_name "Create project transport plan event location prediction"
    xbe_json do project-transport-plan-event-location-predictions create \
        --project-transport-plan-event "$EVENT_ID" \
        --transport-order "$TRANSPORT_ORDER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-event-location-predictions" "$CREATED_ID"
            pass
        else
            fail "Created prediction but no ID returned"
        fi
    else
        fail "Failed to create project transport plan event location prediction"
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_EVENT_ID and XBE_TEST_TRANSPORT_ORDER_ID to run create tests"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete prediction requires --confirm flag"
    xbe_run do project-transport-plan-event-location-predictions delete "$CREATED_ID"
    assert_failure

    test_name "Delete prediction with --confirm"
    xbe_run do project-transport-plan-event-location-predictions delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No prediction created; skipping delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
