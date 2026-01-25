#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Stops
#
# Tests CRUD operations and list filters for the project_transport_plan_stops resource.
#
# COVERAGE: All create/update attributes + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
PLAN_ID=""
LOCATION_ID=""
EVENT_TYPE_ID=""
STATUS=""
POSITION=""
EXT_TMS=""
SKIP_FILTERS=0
CREATED_STOP_ID=""

describe "Resource: project-transport-plan-stops"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan stops"
xbe_json view project-transport-plan-stops list --limit 5
assert_success

test_name "List project transport plan stops returns array"
xbe_json view project-transport-plan-stops list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan stops"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample project transport plan stop"
xbe_json view project-transport-plan-stops list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_FILTERS=1
        skip "No project transport plan stops available"
    fi
else
    SKIP_FILTERS=1
    fail "Failed to list project transport plan stops"
fi

if [[ $SKIP_FILTERS -eq 0 ]]; then
    xbe_json view project-transport-plan-stops show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        PLAN_ID=$(json_get ".project_transport_plan_id")
        LOCATION_ID=$(json_get ".project_transport_location_id")
        EVENT_TYPE_ID=$(json_get ".planned_completion_event_type_id")
        STATUS=$(json_get ".status")
        POSITION=$(json_get ".position")
        EXT_TMS=$(json_get ".external_tms_stop_number")
    else
        SKIP_FILTERS=1
        fail "Failed to fetch project transport plan stop details"
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport plan stops with --project-transport-plan filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stops list --project-transport-plan "$PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List project transport plan stops with --project-transport-location filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$LOCATION_ID" && "$LOCATION_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stops list --project-transport-location "$LOCATION_ID" --limit 5
    assert_success
else
    skip "No project transport location ID available"
fi

test_name "List project transport plan stops with --planned-completion-event-type filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$EVENT_TYPE_ID" && "$EVENT_TYPE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stops list --planned-completion-event-type "$EVENT_TYPE_ID" --limit 5
    assert_success
else
    skip "No planned completion event type ID available"
fi

test_name "List project transport plan stops with --external-tms-stop-number filter"
if [[ $SKIP_FILTERS -eq 0 && -n "$EXT_TMS" && "$EXT_TMS" != "null" ]]; then
    xbe_json view project-transport-plan-stops list --external-tms-stop-number "$EXT_TMS" --limit 5
    assert_success
else
    skip "No external TMS stop number available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan stop"
if [[ $SKIP_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stops show "$SAMPLE_ID"
    assert_success
else
    skip "No project transport plan stop ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan stop requires --project-transport-plan"
xbe_json do project-transport-plan-stops create --project-transport-location 123
assert_failure

test_name "Create project transport plan stop requires --project-transport-location"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stops create --project-transport-plan "$PLAN_ID"
    assert_failure
else
    skip "No project transport plan ID available"
fi

test_name "Create project transport plan stop"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" && -n "$LOCATION_ID" && "$LOCATION_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stops create \
        --project-transport-plan "$PLAN_ID" \
        --project-transport-location "$LOCATION_ID" \
        --status planned

    if [[ $status -eq 0 ]]; then
        CREATED_STOP_ID=$(json_get ".id")
        if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-stops" "$CREATED_STOP_ID"
            pass
        else
            fail "Created stop but no ID returned"
        fi
    else
        fail "Failed to create project transport plan stop"
    fi
else
    skip "Missing project transport plan or location ID"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan stop status"
if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stops update "$CREATED_STOP_ID" --status started
    assert_success
else
    skip "No created stop available for update"
fi

test_name "Update project transport plan stop position"
if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stops update "$CREATED_STOP_ID" --position 2
    assert_success
else
    skip "No created stop available for update"
fi

test_name "Update project transport plan stop planned completion event type"
if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" && -n "$EVENT_TYPE_ID" && "$EVENT_TYPE_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stops update "$CREATED_STOP_ID" --planned-completion-event-type "$EVENT_TYPE_ID"
    assert_success
else
    skip "No event type available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan stop requires --confirm flag"
if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stops delete "$CREATED_STOP_ID"
    assert_failure
else
    skip "No created stop available for delete"
fi

test_name "Delete project transport plan stop"
if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stops delete "$CREATED_STOP_ID" --confirm
    assert_success
else
    skip "No created stop available for delete"
fi

run_tests
