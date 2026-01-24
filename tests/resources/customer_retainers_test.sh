#!/bin/bash
#
# XBE CLI Integration Tests: Customer Retainers
#
# Tests CRUD operations for the customer_retainers resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BROKER2_ID=""
CREATED_CUSTOMER_ID=""
CREATED_CUSTOMER2_ID=""
CREATED_RETAINER_ID=""

UPDATED_MAX_DAILY_HOURS=""


describe "Resource: customer_retainers"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite brokers for customer retainer tests"
BROKER_NAME=$(unique_name "CustomerRetainerBroker")
BROKER2_NAME=$(unique_name "CustomerRetainerBroker2")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
    else
        fail "Failed to create broker"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

xbe_json do brokers create --name "$BROKER2_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER2_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER2_ID" && "$CREATED_BROKER2_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER2_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a second broker"
        run_tests
    fi
else
    fail "Failed to create second broker"
    echo "Cannot continue without a second broker"
    run_tests
fi

# ============================================================================
# Create customers
# ============================================================================

test_name "Create customers for customer retainer tests"
CUSTOMER_NAME=$(unique_name "CustomerRetainerCustomer")
CUSTOMER2_NAME=$(unique_name "CustomerRetainerCustomer2")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
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

xbe_json do customers create --name "$CUSTOMER2_NAME" --broker "$CREATED_BROKER2_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER2_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER2_ID" && "$CREATED_CUSTOMER2_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER2_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a second customer"
        run_tests
    fi
else
    fail "Failed to create second customer"
    echo "Cannot continue without a second customer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create customer retainer with required fields"

xbe_json do customer-retainers create \
    --customer "$CREATED_CUSTOMER_ID" \
    --broker "$CREATED_BROKER_ID" \
    --status editing \
    --maximum-expected-daily-hours 8 \
    --maximum-travel-minutes 90 \
    --billable-travel-minutes-per-travel-mile 2.5

if [[ $status -eq 0 ]]; then
    CREATED_RETAINER_ID=$(json_get ".id")
    if [[ -n "$CREATED_RETAINER_ID" && "$CREATED_RETAINER_ID" != "null" ]]; then
        register_cleanup "customer-retainers" "$CREATED_RETAINER_ID"
        pass
    else
        fail "Created customer retainer but no ID returned"
    fi
else
    fail "Failed to create customer retainer"
fi

if [[ -z "$CREATED_RETAINER_ID" || "$CREATED_RETAINER_ID" == "null" ]]; then
    echo "Cannot continue without a valid customer retainer ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update customer retainer attributes"
UPDATED_MAX_DAILY_HOURS="10"

xbe_json do customer-retainers update "$CREATED_RETAINER_ID" \
    --maximum-expected-daily-hours "$UPDATED_MAX_DAILY_HOURS" \
    --maximum-travel-minutes 120 \
    --billable-travel-minutes-per-travel-mile 1.5

assert_success

test_name "Update customer retainer relationships"

xbe_json do customer-retainers update "$CREATED_RETAINER_ID" \
    --customer "$CREATED_CUSTOMER2_ID" \
    --broker "$CREATED_BROKER2_ID"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show customer retainer"

xbe_json view customer-retainers show "$CREATED_RETAINER_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show customer retainer"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List customer retainers"

xbe_json view customer-retainers list --limit 5
assert_success

test_name "List customer retainers returns array"

xbe_json view customer-retainers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list customer retainers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List customer retainers with --status filter"

xbe_json view customer-retainers list --status editing --limit 5
assert_success

test_name "List customer retainers with --customer filter"

xbe_json view customer-retainers list --customer "$CREATED_CUSTOMER2_ID" --limit 5
assert_success

test_name "List customer retainers with --broker filter"

xbe_json view customer-retainers list --broker "$CREATED_BROKER2_ID" --limit 5
assert_success

test_name "List customer retainers with --created-at-min filter"

xbe_json view customer-retainers list --created-at-min "2000-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer retainers with --created-at-max filter"

xbe_json view customer-retainers list --created-at-max "2099-12-31T23:59:59Z" --limit 5
assert_success

test_name "List customer retainers with --updated-at-min filter"

xbe_json view customer-retainers list --updated-at-min "2000-01-01T00:00:00Z" --limit 5
assert_success

test_name "List customer retainers with --updated-at-max filter"

xbe_json view customer-retainers list --updated-at-max "2099-12-31T23:59:59Z" --limit 5
assert_success

test_name "List customer retainers with --not-id filter"

xbe_json view customer-retainers list --not-id "$CREATED_RETAINER_ID" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List customer retainers with --limit"

xbe_json view customer-retainers list --limit 3
assert_success

test_name "List customer retainers with --offset"

xbe_json view customer-retainers list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete customer retainer requires --confirm flag"

xbe_json do customer-retainers delete "$CREATED_RETAINER_ID"
assert_failure

test_name "Delete customer retainer with --confirm"

xbe_json do customer-retainers delete "$CREATED_RETAINER_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create customer retainer without customer fails"

xbe_json do customer-retainers create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Update customer retainer without any fields fails"

xbe_json do customer-retainers update "$CREATED_RETAINER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
