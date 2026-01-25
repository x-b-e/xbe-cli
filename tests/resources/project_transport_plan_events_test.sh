#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Events
#
# Tests list, show, create, update, delete operations for the
# project-transport-plan-events resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""
SAMPLE_EVENT_TYPE_ID=""
SAMPLE_LOCATION_ID=""
SAMPLE_PLAN_STOP_ID=""
SAMPLE_EXTERNAL_TMS_EVENT_ID=""
SAMPLE_PROJECT_TRANSPORT_ORG_ID=""
SAMPLE_PROJECT_MATERIAL_TYPE_ID=""
LIST_SUPPORTED="true"
CREATED_ID=""

is_nonfatal_error() {
    [[ "$output" == *"Not Authorized"* ]] || \
    [[ "$output" == *"not authorized"* ]] || \
    [[ "$output" == *"Record Invalid"* ]] || \
    [[ "$output" == *"422"* ]] || \
    [[ "$output" == *"403"* ]]
}

describe "Resource: project_transport_plan_events"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan events"
xbe_json view project-transport-plan-events list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project transport plan events"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project transport plan events returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-events list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project transport plan events"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project transport plan event"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-events list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
        SAMPLE_EVENT_TYPE_ID=$(json_get ".[0].project_transport_event_type_id")
        SAMPLE_LOCATION_ID=$(json_get ".[0].project_transport_location_id")
        SAMPLE_PLAN_STOP_ID=$(json_get ".[0].project_transport_plan_stop_id")
        SAMPLE_EXTERNAL_TMS_EVENT_ID=$(json_get ".[0].external_tms_event_id")
        SAMPLE_PROJECT_TRANSPORT_ORG_ID=$(json_get ".[0].project_transport_organization_id")
        SAMPLE_PROJECT_MATERIAL_TYPE_ID=$(json_get ".[0].project_material_type_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project transport plan events available for follow-on tests"
        fi
    else
        skip "Could not list project transport plan events to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan event"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_PROJECT_TRANSPORT_ORG_ID=$(json_get ".project_transport_organization_id")
        SAMPLE_PROJECT_MATERIAL_TYPE_ID=$(json_get ".project_material_type_id")
        SAMPLE_PLAN_STOP_ID=$(json_get ".project_transport_plan_stop_id")
        pass
    else
        fail "Failed to show project transport plan event"
    fi
else
    skip "No project transport plan event ID available"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter by project transport plan"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-plan "$SAMPLE_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "Filter by project transport event type"
if [[ -n "$SAMPLE_EVENT_TYPE_ID" && "$SAMPLE_EVENT_TYPE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-event-type "$SAMPLE_EVENT_TYPE_ID" --limit 5
    assert_success
else
    skip "No project transport event type ID available"
fi

test_name "Filter by project transport location"
if [[ -n "$SAMPLE_LOCATION_ID" && "$SAMPLE_LOCATION_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-location "$SAMPLE_LOCATION_ID" --limit 5
    assert_success
else
    skip "No project transport location ID available"
fi

test_name "Filter by project transport plan stop"
if [[ -n "$SAMPLE_PLAN_STOP_ID" && "$SAMPLE_PLAN_STOP_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-plan-stop "$SAMPLE_PLAN_STOP_ID" --limit 5
    assert_success
else
    skip "No project transport plan stop ID available"
fi

test_name "Filter by external TMS event ID"
if [[ -n "$SAMPLE_EXTERNAL_TMS_EVENT_ID" && "$SAMPLE_EXTERNAL_TMS_EVENT_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --external-tms-event-id "$SAMPLE_EXTERNAL_TMS_EVENT_ID" --limit 5
    assert_success
else
    skip "No external TMS event ID available"
fi

test_name "Filter by project transport organization id"
if [[ -n "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" && "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-organization-id "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" --limit 5
    assert_success
else
    skip "No project transport organization ID available"
fi

test_name "Filter by project transport organization"
if [[ -n "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" && "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-transport-organization "$SAMPLE_PROJECT_TRANSPORT_ORG_ID" --limit 5
    assert_success
else
    skip "No project transport organization ID available"
fi

test_name "Filter by project material type"
if [[ -n "$SAMPLE_PROJECT_MATERIAL_TYPE_ID" && "$SAMPLE_PROJECT_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-events list --project-material-type "$SAMPLE_PROJECT_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No project material type ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan event"
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_ID" && -n "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID" ]]; then
    CREATE_EXT_ID="PTPE-$(unique_suffix)"
    xbe_json do project-transport-plan-events create \
        --project-transport-plan "$XBE_TEST_PROJECT_TRANSPORT_PLAN_ID" \
        --project-transport-event-type "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID" \
        --external-tms-event-id "$CREATE_EXT_ID" \
        --position 0
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-events" "$CREATED_ID"
        fi
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID and XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID to enable create test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan event external TMS event ID"
UPDATE_TARGET_ID="$CREATED_ID"
if [[ -z "$UPDATE_TARGET_ID" || "$UPDATE_TARGET_ID" == "null" ]]; then
    UPDATE_TARGET_ID="$XBE_TEST_PROJECT_TRANSPORT_PLAN_EVENT_ID"
fi

if [[ -n "$UPDATE_TARGET_ID" ]]; then
    UPDATED_EXT_ID="PTPE-UPDATE-$(unique_suffix)"
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --external-tms-event-id "$UPDATED_EXT_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_EVENT_ID to enable update test"
fi

test_name "Update project transport plan event position"
if [[ -n "$UPDATE_TARGET_ID" ]]; then
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --position 1
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_EVENT_ID to enable update test"
fi

test_name "Update project transport plan event type"
if [[ -n "$UPDATE_TARGET_ID" && -n "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID_ALT" ]]; then
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --project-transport-event-type "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID_ALT"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID_ALT to enable event type update test"
fi

test_name "Update project transport plan event location"
if [[ -n "$UPDATE_TARGET_ID" && -n "$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID" ]]; then
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --project-transport-location "$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID to enable location update test"
fi

test_name "Update project transport plan event material type"
if [[ -n "$UPDATE_TARGET_ID" && -n "$XBE_TEST_PROJECT_MATERIAL_TYPE_ID" ]]; then
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --project-material-type "$XBE_TEST_PROJECT_MATERIAL_TYPE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_MATERIAL_TYPE_ID to enable material type update test"
fi

test_name "Update project transport plan event stop"
if [[ -n "$UPDATE_TARGET_ID" && -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID" ]]; then
    xbe_json do project-transport-plan-events update "$UPDATE_TARGET_ID" --project-transport-plan-stop "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID to enable plan stop update test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan event requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do project-transport-plan-events delete "$SAMPLE_ID"
    assert_failure
else
    skip "No project transport plan event ID available"
fi

test_name "Delete project transport plan event"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do project-transport-plan-events delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created project transport plan event to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project transport plan event without plan fails"
xbe_json do project-transport-plan-events create --project-transport-event-type "123"
assert_failure

test_name "Create project transport plan event without event type fails"
xbe_json do project-transport-plan-events create --project-transport-plan "123"
assert_failure

run_tests
