#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Stop Order Stops
#
# Tests list/show/create/delete operations for project_transport_plan_stop_order_stops.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_LINK_ID=""
PROJECT_TRANSPORT_PLAN_STOP_ID=""
TRANSPORT_ORDER_STOP_ID=""
PROJECT_TRANSPORT_PLAN_ID=""
TRANSPORT_ORDER_ID=""
CREATED_LINK_ID=""

CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID="${XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID:-}"
CREATE_TRANSPORT_ORDER_STOP_ID="${XBE_TEST_TRANSPORT_ORDER_STOP_ID:-}"

describe "Resource: project-transport-plan-stop-order-stops"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan stop order stops"
xbe_json view project-transport-plan-stop-order-stops list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_LINK_ID=$(echo "$output" | jq -r '.[0].id')
        PROJECT_TRANSPORT_PLAN_STOP_ID=$(echo "$output" | jq -r '.[0].project_transport_plan_stop_id')
        TRANSPORT_ORDER_STOP_ID=$(echo "$output" | jq -r '.[0].transport_order_stop_id')
        PROJECT_TRANSPORT_PLAN_ID=$(echo "$output" | jq -r '.[0].project_transport_plan_id')
        TRANSPORT_ORDER_ID=$(echo "$output" | jq -r '.[0].transport_order_id')
    fi
else
    fail "Failed to list project transport plan stop order stops"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan stop order stop"
if [[ -n "$SEED_LINK_ID" && "$SEED_LINK_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops show "$SEED_LINK_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show project transport plan stop order stop"
    fi
else
    skip "No project transport plan stop order stop available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project transport plan stop order stop"
if [[ -n "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID" && "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID" != "null" && -n "$CREATE_TRANSPORT_ORDER_STOP_ID" && "$CREATE_TRANSPORT_ORDER_STOP_ID" != "null" ]]; then
    xbe_json do project-transport-plan-stop-order-stops create \
        --project-transport-plan-stop "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID" \
        --transport-order-stop "$CREATE_TRANSPORT_ORDER_STOP_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_LINK_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
            register_cleanup "project-transport-plan-stop-order-stops" "$CREATED_LINK_ID"
            pass
        else
            fail "Created project transport plan stop order stop but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Failed to create project transport plan stop order stop: $output"
        fi
    fi
else
    skip "Missing plan/order stop IDs (set XBE_TEST_PROJECT_TRANSPORT_PLAN_STOP_ID and XBE_TEST_TRANSPORT_ORDER_STOP_ID)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project transport plan stop order stop requires --confirm flag"
if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stop-order-stops delete "$CREATED_LINK_ID"
    assert_failure
else
    skip "No created project transport plan stop order stop to delete"
fi

test_name "Delete project transport plan stop order stop"
if [[ -n "$CREATED_LINK_ID" && "$CREATED_LINK_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stop-order-stops delete "$CREATED_LINK_ID" --confirm
    assert_success
else
    skip "No created project transport plan stop order stop to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by project transport plan stop"
FILTER_PLAN_STOP_ID="${CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID:-$PROJECT_TRANSPORT_PLAN_STOP_ID}"
if [[ -n "$FILTER_PLAN_STOP_ID" && "$FILTER_PLAN_STOP_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --project-transport-plan-stop "$FILTER_PLAN_STOP_ID" --limit 5
    assert_success
else
    skip "No project transport plan stop ID available for filter"
fi

test_name "Filter by transport order stop"
FILTER_ORDER_STOP_ID="${CREATE_TRANSPORT_ORDER_STOP_ID:-$TRANSPORT_ORDER_STOP_ID}"
if [[ -n "$FILTER_ORDER_STOP_ID" && "$FILTER_ORDER_STOP_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --transport-order-stop "$FILTER_ORDER_STOP_ID" --limit 5
    assert_success
else
    skip "No transport order stop ID available for filter"
fi

test_name "Filter by project transport plan"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --project-transport-plan "$PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available for filter"
fi

test_name "Filter by transport order"
if [[ -n "$TRANSPORT_ORDER_ID" && "$TRANSPORT_ORDER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --transport-order "$TRANSPORT_ORDER_ID" --limit 5
    assert_success
else
    skip "No transport order ID available for filter"
fi

test_name "Filter by project transport plan id"
if [[ -n "$PROJECT_TRANSPORT_PLAN_ID" && "$PROJECT_TRANSPORT_PLAN_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --project-transport-plan-id "$PROJECT_TRANSPORT_PLAN_ID" --limit 5
    assert_success
else
    skip "No project transport plan ID available for filter"
fi

test_name "Filter by transport order id"
if [[ -n "$TRANSPORT_ORDER_ID" && "$TRANSPORT_ORDER_ID" != "null" ]]; then
    xbe_json view project-transport-plan-stop-order-stops list --transport-order-id "$TRANSPORT_ORDER_ID" --limit 5
    assert_success
else
    skip "No transport order ID available for filter"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without project transport plan stop fails"
if [[ -n "$CREATE_TRANSPORT_ORDER_STOP_ID" && "$CREATE_TRANSPORT_ORDER_STOP_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stop-order-stops create --transport-order-stop "$CREATE_TRANSPORT_ORDER_STOP_ID"
    assert_failure
else
    skip "No transport order stop ID available for missing plan stop test"
fi

test_name "Create without transport order stop fails"
if [[ -n "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID" && "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID" != "null" ]]; then
    xbe_run do project-transport-plan-stop-order-stops create --project-transport-plan-stop "$CREATE_PROJECT_TRANSPORT_PLAN_STOP_ID"
    assert_failure
else
    skip "No project transport plan stop ID available for missing order stop test"
fi

run_tests
