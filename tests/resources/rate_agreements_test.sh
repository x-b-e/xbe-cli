#!/bin/bash
#
# XBE CLI Integration Tests: Rate Agreements
#
# Tests CRUD operations for the rate_agreements resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_BROKER2_ID=""
CREATED_CUSTOMER_ID=""
CREATED_CUSTOMER2_ID=""
CREATED_RATE_AGREEMENT_ID=""

UPDATED_NAME=""


describe "Resource: rate_agreements"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite brokers for rate agreement tests"
BROKER_NAME=$(unique_name "RateAgreementBroker")
BROKER2_NAME=$(unique_name "RateAgreementBroker2")

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

test_name "Create customers for rate agreement tests"
CUSTOMER_NAME=$(unique_name "RateAgreementCustomer")
CUSTOMER2_NAME=$(unique_name "RateAgreementCustomer2")

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

test_name "Create rate agreement with required fields"
RATE_AGREEMENT_NAME=$(unique_name "RateAgreement")

xbe_json do rate-agreements create \
    --name "$RATE_AGREEMENT_NAME" \
    --status active \
    --seller "Broker|$CREATED_BROKER_ID" \
    --buyer "Customer|$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_RATE_AGREEMENT_ID=$(json_get ".id")
    if [[ -n "$CREATED_RATE_AGREEMENT_ID" && "$CREATED_RATE_AGREEMENT_ID" != "null" ]]; then
        register_cleanup "rate-agreements" "$CREATED_RATE_AGREEMENT_ID"
        pass
    else
        fail "Created rate agreement but no ID returned"
    fi
else
    fail "Failed to create rate agreement"
fi

if [[ -z "$CREATED_RATE_AGREEMENT_ID" || "$CREATED_RATE_AGREEMENT_ID" == "null" ]]; then
    echo "Cannot continue without a valid rate agreement ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update rate agreement attributes"
UPDATED_NAME=$(unique_name "UpdatedRateAgreement")

xbe_json do rate-agreements update "$CREATED_RATE_AGREEMENT_ID" \
    --name "$UPDATED_NAME" \
    --status inactive

assert_success

test_name "Update rate agreement relationships"

xbe_json do rate-agreements update "$CREATED_RATE_AGREEMENT_ID" \
    --seller "Broker|$CREATED_BROKER2_ID" \
    --buyer "Customer|$CREATED_CUSTOMER2_ID"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show rate agreement"

xbe_json view rate-agreements show "$CREATED_RATE_AGREEMENT_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show rate agreement"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List rate agreements"

xbe_json view rate-agreements list --limit 5
assert_success

test_name "List rate agreements returns array"

xbe_json view rate-agreements list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list rate agreements"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List rate agreements with --name filter"

xbe_json view rate-agreements list --name "$UPDATED_NAME" --limit 5
assert_success

test_name "List rate agreements with --status filter"

xbe_json view rate-agreements list --status inactive --limit 5
assert_success

test_name "List rate agreements with --buyer filter"

xbe_json view rate-agreements list --buyer "Customer|$CREATED_CUSTOMER2_ID" --limit 5
assert_success

test_name "List rate agreements with --seller filter"

xbe_json view rate-agreements list --seller "Broker|$CREATED_BROKER2_ID" --limit 5
assert_success

test_name "List rate agreements with --search filter"

xbe_json view rate-agreements list --search "$UPDATED_NAME" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List rate agreements with --limit"

xbe_json view rate-agreements list --limit 3
assert_success

test_name "List rate agreements with --offset"

xbe_json view rate-agreements list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete rate agreement requires --confirm flag"

xbe_json do rate-agreements delete "$CREATED_RATE_AGREEMENT_ID"
assert_failure

test_name "Delete rate agreement with --confirm"

xbe_json do rate-agreements delete "$CREATED_RATE_AGREEMENT_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create rate agreement without status fails"

xbe_json do rate-agreements create \
    --seller "Broker|$CREATED_BROKER_ID" \
    --buyer "Customer|$CREATED_CUSTOMER_ID"

assert_failure

test_name "Update rate agreement without any fields fails"

xbe_json do rate-agreements update "$CREATED_RATE_AGREEMENT_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
