#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Segment Sets
#
# Tests list, show, create, update, and delete operations for project_transport_plan_segment_sets.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_SEGMENT_SET_ID=""
PROJECT_TRANSPORT_PLAN_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
EXTERNAL_TMS_LEG_NUMBER=""
SEGMENT_MILES_SUM=""
CREATED_SEGMENT_SET_ID=""
CREATED_EXTERNAL_TMS_LEG_NUMBER=""

describe "Resource: project-transport-plan-segment-sets"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan segment sets"
xbe_json view project-transport-plan-segment-sets list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_SEGMENT_SET_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$PROJECT_TRANSPORT_PLAN_ID" || "$PROJECT_TRANSPORT_PLAN_ID" == "null" ]]; then
            PROJECT_TRANSPORT_PLAN_ID=$(echo "$output" | jq -r '.[0].project_transport_plan_id')
        fi
        if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
            TRUCKER_ID=$(echo "$output" | jq -r '.[0].trucker_id')
        fi
        EXTERNAL_TMS_LEG_NUMBER=$(echo "$output" | jq -r '.[0].external_tms_leg_number')
        SEGMENT_MILES_SUM=$(echo "$output" | jq -r '.[0].segment_miles_sum')
    fi
else
    fail "Failed to list project transport plan segment sets"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan segment set"
if [[ -n "$SEED_SEGMENT_SET_ID" && "$SEED_SEGMENT_SET_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-sets show "$SEED_SEGMENT_SET_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show project transport plan segment set"
    fi
else
    skip "No project transport plan segment set available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan segment set"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    CREATED_EXTERNAL_TMS_LEG_NUMBER="CLI-SEGSET-$(date -u +\"%Y%m%d%H%M%S\")"
    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        xbe_json do project-transport-plan-segment-sets create \
            --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" \
            --trucker "$TRUCKER_ID" \
            --external-tms-leg-number "$CREATED_EXTERNAL_TMS_LEG_NUMBER" \
            --position 1
    else
        xbe_json do project-transport-plan-segment-sets create \
            --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" \
            --external-tms-leg-number "$CREATED_EXTERNAL_TMS_LEG_NUMBER" \
            --position 1
    fi
    if [[ $status -eq 0 ]]; then
        CREATED_SEGMENT_SET_ID=$(json_get ".id")
        if [[ -n "$CREATED_SEGMENT_SET_ID" && "$CREATED_SEGMENT_SET_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-segment-sets" "$CREATED_SEGMENT_SET_ID"
            SEGMENT_MILES_SUM=$(json_get ".segment_miles_sum")
            pass
        else
            fail "Created project transport plan segment set but no ID returned"
        fi
    else
        fail "Failed to create project transport plan segment set"
    fi
else
    skip "Missing project transport plan ID (set XBE_TEST_PROJECT_TRANSPORT_PLAN_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project transport plan segment set attributes"
if [[ -n "$CREATED_SEGMENT_SET_ID" && "$CREATED_SEGMENT_SET_ID" != "null" ]]; then
    UPDATED_EXTERNAL_TMS_LEG_NUMBER="CLI-SEGSET-UPDATED-$(date -u +\"%H%M%S\")"
    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        xbe_json do project-transport-plan-segment-sets update "$CREATED_SEGMENT_SET_ID" \
            --external-tms-leg-number "$UPDATED_EXTERNAL_TMS_LEG_NUMBER" \
            --position 2 \
            --trucker ""
    else
        xbe_json do project-transport-plan-segment-sets update "$CREATED_SEGMENT_SET_ID" \
            --external-tms-leg-number "$UPDATED_EXTERNAL_TMS_LEG_NUMBER" \
            --position 2
    fi
    assert_success
else
    skip "No created project transport plan segment set to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan segment set requires --confirm flag"
if [[ -n "$CREATED_SEGMENT_SET_ID" && "$CREATED_SEGMENT_SET_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-sets delete "$CREATED_SEGMENT_SET_ID"
    assert_failure
else
    skip "No created project transport plan segment set to delete"
fi

test_name "Delete project transport plan segment set"
if [[ -n "$CREATED_SEGMENT_SET_ID" && "$CREATED_SEGMENT_SET_ID" != "null" ]]; then
    xbe_run do project-transport-plan-segment-sets delete "$CREATED_SEGMENT_SET_ID" --confirm
    assert_success
else
    skip "No created project transport plan segment set to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by project transport plan"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-sets list --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available for filter"
fi

test_name "Filter by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-segment-sets list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available for filter"
fi

test_name "Filter by external TMS leg number"
FILTER_EXTERNAL_TMS_LEG_NUMBER="$EXTERNAL_TMS_LEG_NUMBER"
if [[ -n "$CREATED_EXTERNAL_TMS_LEG_NUMBER" && "$CREATED_EXTERNAL_TMS_LEG_NUMBER" != "null" ]]; then
    FILTER_EXTERNAL_TMS_LEG_NUMBER="$CREATED_EXTERNAL_TMS_LEG_NUMBER"
fi
if [[ -n "$FILTER_EXTERNAL_TMS_LEG_NUMBER" && "$FILTER_EXTERNAL_TMS_LEG_NUMBER" != "null" ]]; then
    xbe_json view project-transport-plan-segment-sets list --external-tms-leg-number "$FILTER_EXTERNAL_TMS_LEG_NUMBER" --limit 5
    assert_success
else
    skip "No external TMS leg number available for filter"
fi

test_name "Filter by segment miles sum"
if [[ -n "$SEGMENT_MILES_SUM" && "$SEGMENT_MILES_SUM" != "null" ]]; then
    xbe_json view project-transport-plan-segment-sets list --segment-miles-sum "$SEGMENT_MILES_SUM" --limit 5
    assert_success
else
    skip "No segment miles sum available for filter"
fi

run_tests
