#!/bin/bash
#
# XBE CLI Integration Tests: Customer Truckers
#
# Tests list, show, create, and delete operations for customer_truckers.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRUCKER_ID=""
CREATED_CUSTOMER_TRUCKER_ID=""

describe "Resource: customer-truckers"

# ============================================================================
# Prerequisites - Create broker, customer, and trucker
# ============================================================================

test_name "Create broker for customer trucker tests"
BROKER_NAME=$(unique_name "CTBroker")

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

test_name "Create customer for customer trucker tests"
CUSTOMER_NAME=$(unique_name "CTCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID"

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
    fail "Failed to create customer"
    echo "Cannot continue without a customer"
    run_tests
fi

test_name "Create trucker for customer trucker tests"
TRUCKER_NAME=$(unique_name "CTTrucker")
TRUCKER_ADDRESS="100 Customer Trucker Lane"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer trucker link"
if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" && -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
    xbe_json do customer-truckers create --customer "$CREATED_CUSTOMER_ID" --trucker "$CREATED_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_CUSTOMER_TRUCKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_CUSTOMER_TRUCKER_ID" && "$CREATED_CUSTOMER_TRUCKER_ID" != "null" ]]; then
            register_cleanup "customer-truckers" "$CREATED_CUSTOMER_TRUCKER_ID"
            pass
        else
            fail "Created customer trucker but no ID returned"
        fi
    else
        fail "Failed to create customer trucker"
    fi
else
    skip "Missing customer or trucker ID for creation"
fi

if [[ -z "$CREATED_CUSTOMER_TRUCKER_ID" || "$CREATED_CUSTOMER_TRUCKER_ID" == "null" ]]; then
    echo "Cannot continue without a valid customer trucker ID"
    run_tests
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List customer truckers"
xbe_json view customer-truckers list --limit 50
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer trucker"
xbe_json view customer-truckers show "$CREATED_CUSTOMER_TRUCKER_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by customer"
xbe_json view customer-truckers list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "Filter by trucker"
xbe_json view customer-truckers list --trucker "$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "Filter by broker"
xbe_json view customer-truckers list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer trucker"
xbe_run do customer-truckers delete "$CREATED_CUSTOMER_TRUCKER_ID" --confirm
assert_success

run_tests
