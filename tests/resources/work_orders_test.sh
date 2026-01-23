#!/bin/bash
#
# XBE CLI Integration Tests: Work Orders
#
# Tests CRUD operations for the work-orders resource.
# Work orders require a broker and responsible party (business unit).
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_WORK_ORDER_ID=""
CREATED_BROKER_ID=""
CREATED_BUSINESS_UNIT_ID=""

describe "Resource: work-orders"

# ============================================================================
# Prerequisites - Create broker and business unit
# ============================================================================

test_name "Create prerequisite broker for work order tests"
BROKER_NAME=$(unique_name "WorkOrderTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite business unit for work order tests"
BUSINESS_UNIT_NAME=$(unique_name "WorkOrderTestBU")

xbe_json do business-units create \
    --name "$BUSINESS_UNIT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_BUSINESS_UNIT_ID=$(json_get ".id")
    if [[ -n "$CREATED_BUSINESS_UNIT_ID" && "$CREATED_BUSINESS_UNIT_ID" != "null" ]]; then
        register_cleanup "business-units" "$CREATED_BUSINESS_UNIT_ID"
        pass
    else
        fail "Created business unit but no ID returned"
        echo "Cannot continue without a business unit"
        run_tests
    fi
else
    fail "Failed to create business unit"
    echo "Cannot continue without a business unit"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create work order with required fields"

xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_WORK_ORDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_WORK_ORDER_ID" && "$CREATED_WORK_ORDER_ID" != "null" ]]; then
        register_cleanup "work-orders" "$CREATED_WORK_ORDER_ID"
        pass
    else
        fail "Created work order but no ID returned"
    fi
else
    fail "Failed to create work order"
fi

# Only continue if we successfully created a work order
if [[ -z "$CREATED_WORK_ORDER_ID" || "$CREATED_WORK_ORDER_ID" == "null" ]]; then
    echo "Cannot continue without a valid work order ID"
    run_tests
fi

test_name "Create work order with --priority"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID" \
    --priority "high"

if [[ $status -eq 0 ]]; then
    PRIORITY_WO_ID=$(json_get ".id")
    if [[ -n "$PRIORITY_WO_ID" && "$PRIORITY_WO_ID" != "null" ]]; then
        register_cleanup "work-orders" "$PRIORITY_WO_ID"
        pass
    else
        fail "Created work order but no ID returned"
    fi
else
    fail "Failed to create work order with --priority"
fi

test_name "Create work order with --note"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID" \
    --note "Test work order note"

if [[ $status -eq 0 ]]; then
    NOTE_WO_ID=$(json_get ".id")
    if [[ -n "$NOTE_WO_ID" && "$NOTE_WO_ID" != "null" ]]; then
        register_cleanup "work-orders" "$NOTE_WO_ID"
        pass
    else
        fail "Created work order but no ID returned"
    fi
else
    fail "Failed to create work order with --note"
fi

test_name "Create work order with --estimated-labor-hours"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID" \
    --estimated-labor-hours 8.5

if [[ $status -eq 0 ]]; then
    HOURS_WO_ID=$(json_get ".id")
    if [[ -n "$HOURS_WO_ID" && "$HOURS_WO_ID" != "null" ]]; then
        register_cleanup "work-orders" "$HOURS_WO_ID"
        pass
    else
        fail "Created work order but no ID returned"
    fi
else
    fail "Failed to create work order with --estimated-labor-hours"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update work order --priority"
xbe_json do work-orders update "$CREATED_WORK_ORDER_ID" --priority "low"
assert_success

test_name "Update work order --note"
xbe_json do work-orders update "$CREATED_WORK_ORDER_ID" --note "Updated note"
assert_success

test_name "Update work order --estimated-labor-hours"
xbe_json do work-orders update "$CREATED_WORK_ORDER_ID" --estimated-labor-hours 4.0
assert_success

test_name "Update work order --estimated-part-cost"
xbe_json do work-orders update "$CREATED_WORK_ORDER_ID" --estimated-part-cost 150.00
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List work orders"
xbe_json view work-orders list --limit 5
assert_success

test_name "List work orders returns array"
xbe_json view work-orders list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list work orders"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List work orders with --broker filter"
xbe_json view work-orders list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List work orders with --responsible-party filter"
xbe_json view work-orders list --responsible-party "$CREATED_BUSINESS_UNIT_ID" --limit 10
assert_success

test_name "List work orders with --priority filter"
xbe_json view work-orders list --priority "high" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List work orders with --limit"
xbe_json view work-orders list --limit 3
assert_success

test_name "List work orders with --offset"
xbe_json view work-orders list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete work order requires --confirm flag"
xbe_run do work-orders delete "$CREATED_WORK_ORDER_ID"
assert_failure

test_name "Delete work order with --confirm"
# Create a work order specifically for deletion
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID" \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"

if [[ $status -eq 0 ]]; then
    DEL_WO_ID=$(json_get ".id")
    if [[ -n "$DEL_WO_ID" && "$DEL_WO_ID" != "null" ]]; then
        xbe_run do work-orders delete "$DEL_WO_ID" --confirm
        assert_success
    else
        skip "Could not create work order for deletion test"
    fi
else
    skip "Could not create work order for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create work order without --broker fails"
xbe_json do work-orders create \
    --responsible-party "$CREATED_BUSINESS_UNIT_ID"
assert_failure

test_name "Create work order without --responsible-party fails"
xbe_json do work-orders create \
    --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update work order without any fields fails"
xbe_run do work-orders update "$CREATED_WORK_ORDER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
