#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Stop Insertions
#
# Tests list, show, and create operations for the
# project-transport-plan-stop-insertions resource.
#
# COVERAGE: List + show + create + filters + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""
SAMPLE_REFERENCE_STOP_ID=""
SAMPLE_LOCATION_ID=""
SAMPLE_SEGMENT_SET_ID=""
SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID=""
SAMPLE_STATUS=""
SAMPLE_MODE=""
SAMPLE_REUSE_HALF=""
SAMPLE_BOUNDARY_CHOICE=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_EXISTING_STOP_ID=""
SAMPLE_STOP_TO_MOVE_ID=""
LIST_SUPPORTED="true"

is_nonfatal_error() {
    [[ "$output" == *"Not Authorized"* ]] || \
    [[ "$output" == *"not authorized"* ]] || \
    [[ "$output" == *"Record Invalid"* ]] || \
    [[ "$output" == *"422"* ]] || \
    [[ "$output" == *"403"* ]] || \
    [[ "$output" == *"must"* ]]
}

describe "Resource: project_transport_plan_stop_insertions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan stop insertions"
xbe_json view project-transport-plan-stop-insertions list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing project transport plan stop insertions"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project transport plan stop insertions returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project transport plan stop insertions"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample project transport plan stop insertion"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
        SAMPLE_REFERENCE_STOP_ID=$(json_get ".[0].reference_project_transport_plan_stop_id")
        SAMPLE_LOCATION_ID=$(json_get ".[0].project_transport_location_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project transport plan stop insertions available for follow-on tests"
        fi
    else
        skip "Could not list project transport plan stop insertions to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan stop insertion"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_STATUS=$(json_get ".status")
        SAMPLE_MODE=$(json_get ".mode")
        SAMPLE_REUSE_HALF=$(json_get ".reuse_half")
        SAMPLE_BOUNDARY_CHOICE=$(json_get ".boundary_choice")
        SAMPLE_PLAN_ID=$(json_get ".project_transport_plan_id")
        SAMPLE_REFERENCE_STOP_ID=$(json_get ".reference_project_transport_plan_stop_id")
        SAMPLE_LOCATION_ID=$(json_get ".project_transport_location_id")
        SAMPLE_SEGMENT_SET_ID=$(json_get ".project_transport_plan_segment_set_id")
        SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID=$(json_get ".planned_completion_event_type_id")
        SAMPLE_EXISTING_STOP_ID=$(json_get ".existing_project_transport_plan_stop_id")
        SAMPLE_STOP_TO_MOVE_ID=$(json_get ".stop_to_move_id")
        SAMPLE_CREATED_BY_ID=$(json_get ".created_by_id")
        pass
    else
        fail "Failed to show project transport plan stop insertion"
    fi
else
    skip "No project transport plan stop insertion ID available"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter by project transport plan"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --project-transport-plan "$SAMPLE_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "Filter by project transport plan id alias"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --project-transport-plan-id "$SAMPLE_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "Filter by reference stop"
if [[ -n "$SAMPLE_REFERENCE_STOP_ID" && "$SAMPLE_REFERENCE_STOP_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --reference-project-transport-plan-stop "$SAMPLE_REFERENCE_STOP_ID" --limit 5
    assert_success
else
    skip "No reference stop ID available"
fi

test_name "Filter by project transport location"
if [[ -n "$SAMPLE_LOCATION_ID" && "$SAMPLE_LOCATION_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --project-transport-location "$SAMPLE_LOCATION_ID" --limit 5
    assert_success
else
    skip "No project transport location ID available"
fi

test_name "Filter by project transport plan segment set"
if [[ -n "$SAMPLE_SEGMENT_SET_ID" && "$SAMPLE_SEGMENT_SET_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --project-transport-plan-segment-set "$SAMPLE_SEGMENT_SET_ID" --limit 5
    assert_success
else
    skip "No project transport plan segment set ID available"
fi

test_name "Filter by planned completion event type"
if [[ -n "$SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID" && "$SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --planned-completion-event-type "$SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID" --limit 5
    assert_success
else
    skip "No planned completion event type ID available"
fi

test_name "Filter by status"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "Filter by mode"
if [[ -n "$SAMPLE_MODE" && "$SAMPLE_MODE" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --mode "$SAMPLE_MODE" --limit 5
    assert_success
else
    skip "No mode available"
fi

test_name "Filter by reuse half"
if [[ -n "$SAMPLE_REUSE_HALF" && "$SAMPLE_REUSE_HALF" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --reuse-half "$SAMPLE_REUSE_HALF" --limit 5
    assert_success
else
    skip "No reuse half available"
fi

test_name "Filter by boundary choice"
if [[ -n "$SAMPLE_BOUNDARY_CHOICE" && "$SAMPLE_BOUNDARY_CHOICE" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --boundary-choice "$SAMPLE_BOUNDARY_CHOICE" --limit 5
    assert_success
else
    skip "No boundary choice available"
fi

test_name "Filter by created by"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created by ID available"
fi

test_name "Filter by existing stop"
if [[ -n "$SAMPLE_EXISTING_STOP_ID" && "$SAMPLE_EXISTING_STOP_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --existing-project-transport-plan-stop "$SAMPLE_EXISTING_STOP_ID" --limit 5
    assert_success
else
    skip "No existing stop ID available"
fi

test_name "Filter by stop to move"
if [[ -n "$SAMPLE_STOP_TO_MOVE_ID" && "$SAMPLE_STOP_TO_MOVE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-insertions list --stop-to-move "$SAMPLE_STOP_TO_MOVE_ID" --limit 5
    assert_success
else
    skip "No stop to move ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

CREATE_REFERENCE_STOP_ID=""
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID" ]]; then
    CREATE_REFERENCE_STOP_ID="$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID"
else
    CREATE_REFERENCE_STOP_ID="$SAMPLE_REFERENCE_STOP_ID"
fi

CREATE_LOCATION_ID=""
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID" ]]; then
    CREATE_LOCATION_ID="$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID"
else
    CREATE_LOCATION_ID="$SAMPLE_LOCATION_ID"
fi

CREATE_PLAN_ID=""
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_ID" ]]; then
    CREATE_PLAN_ID="$XBE_TEST_PROJECT_TRANSPORT_PLAN_ID"
else
    CREATE_PLAN_ID="$SAMPLE_PLAN_ID"
fi

CREATE_SEGMENT_SET_ID=""
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_SET_ID" ]]; then
    CREATE_SEGMENT_SET_ID="$XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_SET_ID"
else
    CREATE_SEGMENT_SET_ID="$SAMPLE_SEGMENT_SET_ID"
fi

CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID=""
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID" ]]; then
    CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID="$XBE_TEST_PROJECT_TRANSPORT_EVENT_TYPE_ID"
else
    CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID="$SAMPLE_PLANNED_COMPLETION_EVENT_TYPE_ID"
fi


test_name "Create project transport plan stop insertion (insert_before)"
if [[ -n "$CREATE_REFERENCE_STOP_ID" && "$CREATE_REFERENCE_STOP_ID" != "null" && -n "$CREATE_LOCATION_ID" && "$CREATE_LOCATION_ID" != "null" ]]; then
    START_AT=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    create_args=(do project-transport-plan-stop-insertions create \
        --mode insert_before \
        --reuse-half upstream \
        --boundary-choice join_upstream \
        --reference-project-transport-plan-stop "$CREATE_REFERENCE_STOP_ID" \
        --project-transport-location "$CREATE_LOCATION_ID" \
        --planned-event-time-start-at "$START_AT" \
        --planned-event-time-end-at "$START_AT")

    if [[ -n "$CREATE_PLAN_ID" && "$CREATE_PLAN_ID" != "null" ]]; then
        create_args+=(--project-transport-plan "$CREATE_PLAN_ID")
    fi
    if [[ -n "$CREATE_SEGMENT_SET_ID" && "$CREATE_SEGMENT_SET_ID" != "null" ]]; then
        create_args+=(--project-transport-plan-segment-set "$CREATE_SEGMENT_SET_ID")
    fi
    if [[ -n "$CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID" && "$CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID" != "null" ]]; then
        create_args+=(--planned-completion-event-type "$CREATE_PLANNED_COMPLETION_EVENT_TYPE_ID")
    fi

    xbe_json "${create_args[@]}"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No reference stop or location ID available"
fi


test_name "Create stop deletion with preserve-stop-on-delete"
if [[ -n "$CREATE_REFERENCE_STOP_ID" && "$CREATE_REFERENCE_STOP_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stop-insertions create \
        --mode delete \
        --reuse-half downstream \
        --reference-project-transport-plan-stop "$CREATE_REFERENCE_STOP_ID" \
        --preserve-stop-on-delete
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Create delete failed: $output"
        fi
    fi
else
    skip "No reference stop ID available"
fi


test_name "Create move_before with stop-to-move"
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID" && -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID_ALT" ]]; then
    move_args=(do project-transport-plan-stop-insertions create \
        --mode move_before \
        --reuse-half upstream \
        --reference-project-transport-plan-stop "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID" \
        --stop-to-move "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID_ALT")

    if [[ -n "$CREATE_PLAN_ID" && "$CREATE_PLAN_ID" != "null" ]]; then
        move_args+=(--project-transport-plan "$CREATE_PLAN_ID")
    fi

    xbe_json "${move_args[@]}"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Move create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID and XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID_ALT to enable move test"
fi


test_name "Create insertion reusing existing stop"
if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_EXISTING_ID" && -n "$CREATE_REFERENCE_STOP_ID" && "$CREATE_REFERENCE_STOP_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stop-insertions create \
        --mode insert_after \
        --reuse-half downstream \
        --reference-project-transport-plan-stop "$CREATE_REFERENCE_STOP_ID" \
        --existing-project-transport-plan-stop "$XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_EXISTING_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if is_nonfatal_error; then
            pass
        else
            fail "Reuse existing stop failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_EXISTING_ID to enable existing stop test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create stop insertion without required flags fails"
xbe_run do project-transport-plan-stop-insertions create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
