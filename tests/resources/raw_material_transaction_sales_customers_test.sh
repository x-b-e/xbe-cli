#!/bin/bash
#
# XBE CLI Integration Tests: Raw Material Transaction Sales Customers
#
# Tests list, show, create, update, and delete operations for raw-material-transaction-sales-customers.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_CUSTOMER_ALT_ID=""
CREATED_RAW_SALES_CUSTOMER_ID=""

describe "Resource: raw-material-transaction-sales-customers"

# ============================================================================
# Prerequisites - Create broker and customers
# ============================================================================

test_name "Create broker for raw material transaction sales customer tests"
BROKER_NAME=$(unique_name "RMTBroker")

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

test_name "Create customer for raw material transaction sales customer tests"
CUSTOMER_NAME=$(unique_name "RMTCustomer")

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

test_name "Create alternate customer for update tests"
CUSTOMER_ALT_NAME=$(unique_name "RMTCustomerAlt")

xbe_json do customers create \
    --name "$CUSTOMER_ALT_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ALT_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ALT_ID" && "$CREATED_CUSTOMER_ALT_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ALT_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without an alternate customer"
        run_tests
    fi
else
    fail "Failed to create alternate customer"
    echo "Cannot continue without an alternate customer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create raw material transaction sales customer"
RAW_SALES_CUSTOMER_ID=$(unique_name "RawSalesCustomer")

if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
    xbe_json do raw-material-transaction-sales-customers create \
        --raw-sales-customer-id "$RAW_SALES_CUSTOMER_ID" \
        --customer "$CREATED_CUSTOMER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_RAW_SALES_CUSTOMER_ID=$(json_get ".id")
        if [[ -n "$CREATED_RAW_SALES_CUSTOMER_ID" && "$CREATED_RAW_SALES_CUSTOMER_ID" != "null" ]]; then
            register_cleanup "raw-material-transaction-sales-customers" "$CREATED_RAW_SALES_CUSTOMER_ID"
            pass
        else
            fail "Created raw material transaction sales customer but no ID returned"
        fi
    else
        fail "Failed to create raw material transaction sales customer"
    fi
else
    skip "Missing customer ID for creation"
fi

if [[ -z "$CREATED_RAW_SALES_CUSTOMER_ID" || "$CREATED_RAW_SALES_CUSTOMER_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw material transaction sales customer ID"
    run_tests
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List raw material transaction sales customers"
xbe_json view raw-material-transaction-sales-customers list --limit 50
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show raw material transaction sales customer"
xbe_json view raw-material-transaction-sales-customers show "$CREATED_RAW_SALES_CUSTOMER_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by customer"
xbe_json view raw-material-transaction-sales-customers list --customer "$CREATED_CUSTOMER_ID" --limit 5
assert_success

test_name "Filter by broker"
xbe_json view raw-material-transaction-sales-customers list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update raw sales customer identifier"
UPDATED_RAW_SALES_CUSTOMER_ID=$(unique_name "RawSalesCustomerUpdated")
xbe_json do raw-material-transaction-sales-customers update "$CREATED_RAW_SALES_CUSTOMER_ID" \
    --raw-sales-customer-id "$UPDATED_RAW_SALES_CUSTOMER_ID"
assert_success

test_name "Update customer"
xbe_json do raw-material-transaction-sales-customers update "$CREATED_RAW_SALES_CUSTOMER_ID" \
    --customer "$CREATED_CUSTOMER_ALT_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete raw material transaction sales customer"
xbe_run do raw-material-transaction-sales-customers delete "$CREATED_RAW_SALES_CUSTOMER_ID" --confirm
assert_success

run_tests
