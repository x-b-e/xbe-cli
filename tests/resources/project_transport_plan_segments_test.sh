#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Segments
#
# Tests view/create/update/delete behavior for project-transport-plan-segments.
#
# COVERAGE: List + list filters + show + create/update/delete + required flag failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-segments"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""
SAMPLE_ORIGIN_ID=""
SAMPLE_DESTINATION_ID=""
SAMPLE_SEGMENT_SET_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_TMS_ORDER=""
SAMPLE_TMS_MOVEMENT=""

CREATED_ID=""

PLAN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
ORIGIN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ORIGIN_ID:-}"
DESTINATION_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_DESTINATION_ID:-}"
SEGMENT_SET_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_SET_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
EXTERNAL_ID_VALUE="${XBE_TEST_EXTERNAL_IDENTIFICATION_VALUE:-}"
TMS_ORDER="${XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_TMS_ORDER_NUMBER:-}"
TMS_MOVEMENT="${XBE_TEST_PROJECT_TRANSPORT_PLAN_SEGMENT_TMS_MOVEMENT_NUMBER:-}"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan segments"
xbe_json view project-transport-plan-segments list --limit 5
assert_success

test_name "List project transport plan segments returns array"
xbe_json view project-transport-plan-segments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan segments"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample segment"
xbe_json view project-transport-plan-segments list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PLAN_ID=$(json_get ".[0].project_transport_plan_id")
    SAMPLE_ORIGIN_ID=$(json_get ".[0].origin_id")
    SAMPLE_DESTINATION_ID=$(json_get ".[0].destination_id")
    SAMPLE_SEGMENT_SET_ID=$(json_get ".[0].project_transport_plan_segment_set_id")
    SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
    SAMPLE_TMS_ORDER=$(json_get ".[0].external_tms_order_number")
    SAMPLE_TMS_MOVEMENT=$(json_get ".[0].external_tms_movement_number")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No segments available for follow-on tests"
    fi
else
    skip "Could not list segments to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List segments with --project-transport-plan filter"
FILTER_PLAN_ID="$SAMPLE_PLAN_ID"
if [[ -z "$FILTER_PLAN_ID" || "$FILTER_PLAN_ID" == "null" ]]; then
    FILTER_PLAN_ID="$PLAN_ID"
fi
if [[ -n "$FILTER_PLAN_ID" && "$FILTER_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --project-transport-plan "$FILTER_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available"
fi

test_name "List segments with --origin filter"
if [[ -n "$SAMPLE_ORIGIN_ID" && "$SAMPLE_ORIGIN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --origin "$SAMPLE_ORIGIN_ID" --limit 5
    assert_success
else
    skip "No origin stop ID available"
fi

test_name "List segments with --destination filter"
if [[ -n "$SAMPLE_DESTINATION_ID" && "$SAMPLE_DESTINATION_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --destination "$SAMPLE_DESTINATION_ID" --limit 5
    assert_success
else
    skip "No destination stop ID available"
fi

test_name "List segments with --project-transport-plan-segment-set filter"
FILTER_SEGMENT_SET_ID="$SAMPLE_SEGMENT_SET_ID"
if [[ -z "$FILTER_SEGMENT_SET_ID" || "$FILTER_SEGMENT_SET_ID" == "null" ]]; then
    FILTER_SEGMENT_SET_ID="$SEGMENT_SET_ID"
fi
if [[ -n "$FILTER_SEGMENT_SET_ID" && "$FILTER_SEGMENT_SET_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --project-transport-plan-segment-set "$FILTER_SEGMENT_SET_ID" --limit 5
    assert_success
else
    skip "No segment set ID available"
fi

test_name "List segments with --trucker filter"
FILTER_TRUCKER_ID="$SAMPLE_TRUCKER_ID"
if [[ -z "$FILTER_TRUCKER_ID" || "$FILTER_TRUCKER_ID" == "null" ]]; then
    FILTER_TRUCKER_ID="$TRUCKER_ID"
fi
if [[ -n "$FILTER_TRUCKER_ID" && "$FILTER_TRUCKER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --trucker "$FILTER_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List segments with --external-tms-order-number filter"
FILTER_TMS_ORDER="$SAMPLE_TMS_ORDER"
if [[ -z "$FILTER_TMS_ORDER" || "$FILTER_TMS_ORDER" == "null" ]]; then
    FILTER_TMS_ORDER="$TMS_ORDER"
fi
if [[ -n "$FILTER_TMS_ORDER" && "$FILTER_TMS_ORDER" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --external-tms-order-number "$FILTER_TMS_ORDER" --limit 5
    assert_success
else
    skip "No external TMS order number available"
fi

test_name "List segments with --external-tms-movement-number filter"
FILTER_TMS_MOVEMENT="$SAMPLE_TMS_MOVEMENT"
if [[ -z "$FILTER_TMS_MOVEMENT" || "$FILTER_TMS_MOVEMENT" == "null" ]]; then
    FILTER_TMS_MOVEMENT="$TMS_MOVEMENT"
fi
if [[ -n "$FILTER_TMS_MOVEMENT" && "$FILTER_TMS_MOVEMENT" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --external-tms-movement-number "$FILTER_TMS_MOVEMENT" --limit 5
    assert_success
else
    skip "No external TMS movement number available"
fi

test_name "List segments with --external-identification-value filter"
if [[ -n "$EXTERNAL_ID_VALUE" && "$EXTERNAL_ID_VALUE" != "null" ]]; then
    xbe_json view project-transport-plan-segments list --external-identification-value "$EXTERNAL_ID_VALUE" --limit 5
    assert_success
else
    skip "Set XBE_TEST_EXTERNAL_IDENTIFICATION_VALUE to test external identification filter"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show segment details"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segments show "$SAMPLE_ID"
    assert_success
else
    skip "No segment ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create segment without required flags fails"
xbe_run do project-transport-plan-segments create
assert_failure

if [[ -n "$PLAN_ID" && -n "$ORIGIN_ID" && -n "$DESTINATION_ID" ]]; then
    test_name "Create project transport plan segment"

    create_args=(do project-transport-plan-segments create \
        --project-transport-plan "$PLAN_ID" \
        --origin "$ORIGIN_ID" \
        --destination "$DESTINATION_ID" \
        --position 1 \
        --miles 12.5 \
        --miles-source unknown)

    if [[ -n "$SEGMENT_SET_ID" ]]; then
        create_args+=(--project-transport-plan-segment-set "$SEGMENT_SET_ID")
    fi
    if [[ -n "$TRUCKER_ID" ]]; then
        create_args+=(--trucker "$TRUCKER_ID")
    fi
    if [[ -n "$TMS_ORDER" ]]; then
        create_args+=(--external-tms-order-number "$TMS_ORDER")
    fi
    if [[ -n "$TMS_MOVEMENT" ]]; then
        create_args+=(--external-tms-movement-number "$TMS_MOVEMENT")
    fi

    xbe_json "${create_args[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-segments" "$CREATED_ID"
            pass
        else
            fail "Created segment but no ID returned"
        fi
    else
        fail "Failed to create project transport plan segment"
    fi
else
    skip "Set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID, XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ORIGIN_ID, and XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_DESTINATION_ID to run create tests"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Update segment position"
    xbe_json do project-transport-plan-segments update "$CREATED_ID" --position 2
    assert_success

    test_name "Update segment miles"
    xbe_json do project-transport-plan-segments update "$CREATED_ID" --miles 15.4
    assert_success

    test_name "Update segment miles source"
    xbe_json do project-transport-plan-segments update "$CREATED_ID" --miles-source transport_route
    assert_success

    if [[ -n "$SEGMENT_SET_ID" ]]; then
        test_name "Update segment set"
        xbe_json do project-transport-plan-segments update "$CREATED_ID" --project-transport-plan-segment-set "$SEGMENT_SET_ID"
        assert_success
    else
        skip "No segment set ID available for update"
    fi

    if [[ -n "$TRUCKER_ID" ]]; then
        test_name "Update segment trucker"
        xbe_json do project-transport-plan-segments update "$CREATED_ID" --trucker "$TRUCKER_ID"
        assert_success
    else
        skip "No trucker ID available for update"
    fi

    if [[ -n "$TMS_ORDER" ]]; then
        test_name "Update external TMS order number"
        xbe_json do project-transport-plan-segments update "$CREATED_ID" --external-tms-order-number "$TMS_ORDER"
        assert_success
    else
        skip "No external TMS order number available for update"
    fi

    if [[ -n "$TMS_MOVEMENT" ]]; then
        test_name "Update external TMS movement number"
        xbe_json do project-transport-plan-segments update "$CREATED_ID" --external-tms-movement-number "$TMS_MOVEMENT"
        assert_success
    else
        skip "No external TMS movement number available for update"
    fi
else
    skip "No segment created; skipping update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    test_name "Delete segment requires --confirm flag"
    xbe_run do project-transport-plan-segments delete "$CREATED_ID"
    assert_failure

    test_name "Delete segment with --confirm"
    xbe_run do project-transport-plan-segments delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No segment created; skipping delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
