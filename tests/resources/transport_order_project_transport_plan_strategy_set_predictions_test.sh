#!/bin/bash
#
# XBE CLI Integration Tests: Transport Order Strategy Set Predictions
#
# Tests create, list, show, and delete operations for the
# transport_order_project_transport_plan_strategy_set_predictions resource.
#
# COVERAGE: Create + list filters + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRANSPORT_ORDER_ID=""
CREATED_PREDICTION_ID=""

describe "Resource: transport-order-project-transport-plan-strategy-set-predictions"

# ============================================================================
# Prerequisites - Create broker, customer, and transport order
# ============================================================================

test_name "Create prerequisite broker for strategy set prediction tests"
BROKER_NAME=$(unique_name "TOStrategyBroker")

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

test_name "Create prerequisite customer for strategy set prediction tests"
CUSTOMER_NAME=$(unique_name "TOStrategyCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Select transport order for strategy set predictions"
if [[ -n "$XBE_TEST_TRANSPORT_ORDER_ID" ]]; then
    CREATED_TRANSPORT_ORDER_ID="$XBE_TEST_TRANSPORT_ORDER_ID"
    echo "    Using XBE_TEST_TRANSPORT_ORDER_ID: $CREATED_TRANSPORT_ORDER_ID"
    pass
elif [[ -n "$XBE_TEST_BROKER_ID" ]]; then
    xbe_json view transport-orders list --broker "$XBE_TEST_BROKER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        existing_id=$(json_get ".[0].id")
        if [[ -n "$existing_id" && "$existing_id" != "null" ]]; then
            CREATED_TRANSPORT_ORDER_ID="$existing_id"
            echo "    Using existing transport order: $CREATED_TRANSPORT_ORDER_ID"
            pass
        fi
    fi
fi

if [[ -z "$CREATED_TRANSPORT_ORDER_ID" || "$CREATED_TRANSPORT_ORDER_ID" == "null" ]]; then
    xbe_json do transport-orders create --customer "$CREATED_CUSTOMER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_TRANSPORT_ORDER_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRANSPORT_ORDER_ID" && "$CREATED_TRANSPORT_ORDER_ID" != "null" ]]; then
            pass
        else
            fail "Created transport order but no ID returned"
            echo "Cannot continue without a transport order"
            run_tests
        fi
    else
        fail "Failed to create transport order"
        echo "Cannot continue without a transport order"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create strategy set prediction for transport order"
xbe_json do transport-order-project-transport-plan-strategy-set-predictions create \
    --transport-order "$CREATED_TRANSPORT_ORDER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PREDICTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PREDICTION_ID" && "$CREATED_PREDICTION_ID" != "null" ]]; then
        register_cleanup "transport-order-project-transport-plan-strategy-set-predictions" "$CREATED_PREDICTION_ID"
        pass
    else
        fail "Created prediction but no ID returned"
    fi
else
    if [[ "$output" == *"predictions - can't be blank"* ]]; then
        skip "Prediction requires transport orders with stops/locations; set XBE_TEST_TRANSPORT_ORDER_ID to a valid order"
    else
        fail "Failed to create strategy set prediction: $output"
    fi
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================


test_name "List strategy set predictions"
xbe_json view transport-order-project-transport-plan-strategy-set-predictions list
assert_success

test_name "List strategy set predictions returns array"
xbe_json view transport-order-project-transport-plan-strategy-set-predictions list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list strategy set predictions"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================


test_name "List strategy set predictions with --transport-order filter"
xbe_json view transport-order-project-transport-plan-strategy-set-predictions list \
    --transport-order "$CREATED_TRANSPORT_ORDER_ID"
assert_success

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================


test_name "List strategy set predictions with --limit"
xbe_json view transport-order-project-transport-plan-strategy-set-predictions list --limit 5
assert_success

test_name "List strategy set predictions with --offset"
xbe_json view transport-order-project-transport-plan-strategy-set-predictions list --limit 5 --offset 5
assert_success

# ==========================================================================
# SHOW Tests
# ==========================================================================

if [[ -n "$CREATED_PREDICTION_ID" && "$CREATED_PREDICTION_ID" != "null" ]]; then
    test_name "Show strategy set prediction details"
    xbe_json view transport-order-project-transport-plan-strategy-set-predictions show "$CREATED_PREDICTION_ID"
    assert_success
else
    skip "No prediction ID available for show test"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

if [[ -n "$CREATED_PREDICTION_ID" && "$CREATED_PREDICTION_ID" != "null" ]]; then
    test_name "Delete prediction requires --confirm"
    xbe_json do transport-order-project-transport-plan-strategy-set-predictions delete "$CREATED_PREDICTION_ID"
    assert_failure

    test_name "Delete prediction with --confirm"
    xbe_json do transport-order-project-transport-plan-strategy-set-predictions delete "$CREATED_PREDICTION_ID" --confirm
    assert_success
else
    skip "No prediction ID available for delete test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================


test_name "Create prediction without transport order fails"
xbe_json do transport-order-project-transport-plan-strategy-set-predictions create
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
