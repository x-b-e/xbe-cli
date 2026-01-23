#!/bin/bash
#
# XBE CLI Integration Tests: Transport Orders
#
# Tests CRUD operations for the transport_orders resource.
# Transport orders are requests for material transport.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TO_ID=""

describe "Resource: transport-orders"

# ============================================================================
# Prerequisites - Create broker and customer for transport orders
# ============================================================================

test_name "Create prerequisite broker for transport orders tests"
BROKER_NAME=$(unique_name "TOTestBroker")

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

test_name "Create prerequisite customer for transport orders tests"
CUSTOMER_NAME=$(unique_name "TOTestCustomer")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create transport order with required fields"
xbe_json do transport-orders create --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TO_ID=$(json_get ".id")
    if [[ -n "$CREATED_TO_ID" && "$CREATED_TO_ID" != "null" ]]; then
        # Note: No delete available for transport-orders
        pass
    else
        fail "Created transport order but no ID returned"
    fi
else
    fail "Failed to create transport order: $output"
fi

test_name "Create transport order with ordered-at"
xbe_json do transport-orders create \
    --customer "$CREATED_CUSTOMER_ID" \
    --ordered-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create transport order with ordered-at"
fi

test_name "Create managed transport order"
xbe_json do transport-orders create \
    --customer "$CREATED_CUSTOMER_ID" \
    --is-managed

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create managed transport order"
fi

test_name "Create transport order with billable miles"
xbe_json do transport-orders create \
    --customer "$CREATED_CUSTOMER_ID" \
    --billable-miles "50"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create transport order with billable miles"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

# Only continue if we have a valid transport order ID
if [[ -n "$CREATED_TO_ID" && "$CREATED_TO_ID" != "null" ]]; then

test_name "Update transport order billable-miles"
xbe_json do transport-orders update "$CREATED_TO_ID" --billable-miles "75"
assert_success

test_name "Update transport order ordered-at"
xbe_json do transport-orders update "$CREATED_TO_ID" --ordered-at "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
assert_success

test_name "Update transport order is-managed to true"
xbe_json do transport-orders update "$CREATED_TO_ID" --is-managed true
assert_success

test_name "Update transport order is-managed to false"
xbe_json do transport-orders update "$CREATED_TO_ID" --is-managed false
assert_success

fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport orders"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID"
assert_success

test_name "List transport orders returns array"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list transport orders"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List transport orders with --active filter"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID" --active true
assert_success

test_name "List transport orders with --is-managed filter"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID" --is-managed true
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List transport orders with --limit"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List transport orders with --offset"
xbe_json view transport-orders list --broker "$CREATED_BROKER_ID" --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create transport order without customer fails"
xbe_json do transport-orders create
assert_failure

test_name "Update without any fields fails"
if [[ -n "$CREATED_TO_ID" && "$CREATED_TO_ID" != "null" ]]; then
    xbe_json do transport-orders update "$CREATED_TO_ID"
    assert_failure
else
    skip "No transport order ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete transport order requires --confirm flag"
# Create a transport order for deletion test
xbe_json do transport-orders create --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do transport-orders delete "$DEL_ID"
    assert_failure
else
    skip "Could not create transport order for deletion test"
fi

test_name "Delete transport order with --confirm"
xbe_json do transport-orders create --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID2=$(json_get ".id")
    xbe_json do transport-orders delete "$DEL_ID2" --confirm
    assert_success
else
    skip "Could not create transport order for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
