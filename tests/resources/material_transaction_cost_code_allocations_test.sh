#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Cost Code Allocations
#
# Tests create/update/delete operations and list filters for the
# material_transaction_cost_code_allocations resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_COST_CODE_ID=""
CREATED_COST_CODE_ID_2=""
CREATED_MT_ID=""
CREATED_ALLOCATION_ID=""

describe "Resource: material-transaction-cost-code-allocations"

# ==========================================================================
# Prerequisites - Create broker, customer, cost codes, material transaction
# ==========================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MTCostAllocBroker")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "MTCostAllocCustomer")

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

test_name "Create first cost code"
COST_CODE_1="MTA-$(unique_suffix)"

xbe_json do cost-codes create \
    --code "$COST_CODE_1" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID" && "$CREATED_COST_CODE_ID" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID"
        pass
    else
        fail "Created cost code but no ID returned"
        echo "Cannot continue without a cost code"
        run_tests
    fi
else
    fail "Failed to create cost code"
    echo "Cannot continue without a cost code"
    run_tests
fi

test_name "Create second cost code"
COST_CODE_2="MTA2-$(unique_suffix)"

xbe_json do cost-codes create \
    --code "$COST_CODE_2" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_COST_CODE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_COST_CODE_ID_2" && "$CREATED_COST_CODE_ID_2" != "null" ]]; then
        register_cleanup "cost-codes" "$CREATED_COST_CODE_ID_2"
        pass
    else
        fail "Created cost code but no ID returned"
        echo "Cannot continue without a second cost code"
        run_tests
    fi
else
    fail "Failed to create second cost code"
    echo "Cannot continue without a second cost code"
    run_tests
fi

test_name "Create material transaction"
TICKET_NUM="MTA-T$(date +%s)"
TRANS_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

xbe_json do material-transactions create --ticket-number "$TICKET_NUM" --transaction-at "$TRANS_AT"

if [[ $status -eq 0 ]]; then
    CREATED_MT_ID=$(json_get ".id")
    if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
        register_cleanup "material-transactions" "$CREATED_MT_ID"
        pass
    else
        fail "Created material transaction but no ID returned"
        echo "Cannot continue without a material transaction"
        run_tests
    fi
else
    fail "Failed to create material transaction"
    echo "Cannot continue without a material transaction"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material transaction cost code allocation"
DETAILS_JSON="[{\"cost_code_id\":\"$CREATED_COST_CODE_ID\",\"percentage\":1}]"

xbe_json do material-transaction-cost-code-allocations create \
    --material-transaction "$CREATED_MT_ID" \
    --details "$DETAILS_JSON"

if [[ $status -eq 0 ]]; then
    CREATED_ALLOCATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
        register_cleanup "material-transaction-cost-code-allocations" "$CREATED_ALLOCATION_ID"
        pass
    else
        fail "Created allocation but no ID returned"
        echo "Cannot continue without a material transaction cost code allocation"
        run_tests
    fi
else
    fail "Failed to create material transaction cost code allocation"
    echo "Cannot continue without a material transaction cost code allocation"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show material transaction cost code allocation"
xbe_json view material-transaction-cost-code-allocations show "$CREATED_ALLOCATION_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update material transaction cost code allocation details"
UPDATED_DETAILS_JSON="[{\"cost_code_id\":\"$CREATED_COST_CODE_ID\",\"percentage\":0.5},{\"cost_code_id\":\"$CREATED_COST_CODE_ID_2\",\"percentage\":0.5}]"

xbe_json do material-transaction-cost-code-allocations update "$CREATED_ALLOCATION_ID" \
    --details "$UPDATED_DETAILS_JSON"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material transaction cost code allocations"
xbe_json view material-transaction-cost-code-allocations list --limit 10
assert_success

test_name "List material transaction cost code allocations returns array"
xbe_json view material-transaction-cost-code-allocations list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction cost code allocations"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material transaction cost code allocations with --material-transaction filter"
xbe_json view material-transaction-cost-code-allocations list --material-transaction "$CREATED_MT_ID" --limit 10
assert_success

test_name "List material transaction cost code allocations with --created-at-min filter"
xbe_json view material-transaction-cost-code-allocations list --created-at-min 2000-01-01T00:00:00Z --limit 10
assert_success

test_name "List material transaction cost code allocations with --created-at-max filter"
xbe_json view material-transaction-cost-code-allocations list --created-at-max 2100-01-01T00:00:00Z --limit 10
assert_success

test_name "List material transaction cost code allocations with --updated-at-min filter"
xbe_json view material-transaction-cost-code-allocations list --updated-at-min 2000-01-01T00:00:00Z --limit 10
assert_success

test_name "List material transaction cost code allocations with --updated-at-max filter"
xbe_json view material-transaction-cost-code-allocations list --updated-at-max 2100-01-01T00:00:00Z --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete material transaction cost code allocation requires --confirm flag"
xbe_json do material-transaction-cost-code-allocations delete "$CREATED_ALLOCATION_ID"
assert_failure

test_name "Delete material transaction cost code allocation with --confirm"
xbe_json do material-transaction-cost-code-allocations delete "$CREATED_ALLOCATION_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
