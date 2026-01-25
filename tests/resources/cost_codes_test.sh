#!/bin/bash
#
# XBE CLI Integration Tests: Cost Codes
#
# Tests CRUD operations for the cost_codes resource.
# Cost codes are used to categorize and track costs.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_COST_CODE_ID=""
CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRUCKER_ID=""

describe "Resource: cost_codes"

# ============================================================================
# Prerequisites - Create broker, customer, and trucker for tests
# ============================================================================

test_name "Create prerequisite broker for cost code tests"
BROKER_NAME=$(unique_name "CCTestBroker")

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
    fail "Failed to create broker"
    echo "Cannot continue without a broker"
    run_tests
fi

test_name "Create prerequisite customer for cost code tests"
CUSTOMER_NAME=$(unique_name "CCTestCustomer")

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
    fi
else
    fail "Failed to create customer"
fi

test_name "Create prerequisite trucker for cost code tests"
TRUCKER_NAME=$(unique_name "CCTestTrucker")

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "100 Trucker Way, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
    fi
else
    fail "Failed to create trucker"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cost code with required fields"
TEST_CODE="CC-$(date +%s | tail -c 6)"

# Note: Cost codes must be associated with a customer or trucker (not broker directly)
xbe_json do cost-codes create \
    --code "$TEST_CODE" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID" && "$CREATED_COST_CODE_ID" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID"
        pass
    else
        fail "Created cost code but no ID returned"
    fi
else
    fail "Failed to create cost code"
fi

# Only continue if we successfully created a cost code
if [[ -z "$CREATED_COST_CODE_ID" || "$CREATED_COST_CODE_ID" == "null" ]]; then
    echo "Cannot continue without a valid cost code ID"
    run_tests
fi

test_name "Create cost code with description"
TEST_CODE2="CC2-$(date +%s | tail -c 6)"
xbe_json do cost-codes create \
    --code "$TEST_CODE2" \
    --description "Test cost code with description" \
    --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-codes" "$id"
    pass
else
    fail "Failed to create cost code with description"
fi

test_name "Create cost code with active=false"
TEST_CODE3="CC3-$(date +%s | tail -c 6)"
xbe_json do cost-codes create \
    --code "$TEST_CODE3" \
    --active=false \
    --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-codes" "$id"
    pass
else
    fail "Failed to create cost code with active=false"
fi

test_name "Create cost code with customer"
TEST_CODE4="CC4-$(date +%s | tail -c 6)"
xbe_json do cost-codes create \
    --code "$TEST_CODE4" \
    --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-codes" "$id"
    pass
else
    fail "Failed to create cost code with customer"
fi

test_name "Create cost code with trucker"
TEST_CODE5="CC5-$(date +%s | tail -c 6)"
xbe_json do cost-codes create \
    --code "$TEST_CODE5" \
    --trucker "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "cost-codes" "$id"
    pass
else
    fail "Failed to create cost code with trucker"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cost code description"
xbe_json do cost-codes update "$CREATED_COST_CODE_ID" --description "Updated description"
assert_success

test_name "Update cost code to inactive"
xbe_json do cost-codes update "$CREATED_COST_CODE_ID" --no-active
assert_success

test_name "Update cost code to active"
xbe_json do cost-codes update "$CREATED_COST_CODE_ID" --active
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List cost codes"
xbe_json view cost-codes list --limit 5
assert_success

test_name "List cost codes returns array"
xbe_json view cost-codes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list cost codes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List cost codes with --code filter"
xbe_json view cost-codes list --code "$TEST_CODE" --limit 10
assert_success

test_name "List cost codes with --description filter"
xbe_json view cost-codes list --description "Updated" --limit 10
assert_success

test_name "List cost codes with --q filter"
xbe_json view cost-codes list --q "CC" --limit 10
assert_success

test_name "List cost codes with --broker filter"
xbe_json view cost-codes list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List cost codes with --customer filter"
xbe_json view cost-codes list --customer "$CREATED_CUSTOMER_ID" --limit 10
assert_success

test_name "List cost codes with --trucker filter"
xbe_json view cost-codes list --trucker "$CREATED_TRUCKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List cost codes with --limit"
xbe_json view cost-codes list --limit 3
assert_success

test_name "List cost codes with --offset"
xbe_json view cost-codes list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete cost code requires --confirm flag"
xbe_run do cost-codes delete "$CREATED_COST_CODE_ID"
assert_failure

test_name "Delete cost code with --confirm"
# Create a cost code specifically for deletion
TEST_DEL_CODE="DEL-$(date +%s | tail -c 6)"
xbe_json do cost-codes create \
    --code "$TEST_DEL_CODE" \
    --customer "$CREATED_CUSTOMER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do cost-codes delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create cost code for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cost code without code fails"
xbe_json do cost-codes create --description "No code"
assert_failure

test_name "Update without any fields fails"
xbe_json do cost-codes update "$CREATED_COST_CODE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
