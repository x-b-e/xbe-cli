#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Cost Codes
#
# Tests CRUD operations for the job_production_plan_cost_codes resource.
#
# COVERAGE: All filters + all create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_COST_CODE_ID=""
CREATED_CLASSIFICATION_ID=""
CREATED_CLASSIFICATION_ID2=""
CREATED_JPP_COST_CODE_ID=""

COST_CODE_VALUE=""
COST_CODE_DESC=""


describe "Resource: job-production-plan-cost-codes"

# ============================================================================
# Prerequisites - Create broker, customer, job production plan, cost code
# ============================================================================

test_name "Create prerequisite broker for job production plan cost code tests"
BROKER_NAME=$(unique_name "JPPCCTestBroker")

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
        fail "Failed to create broker"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPCCTestCustomer")

xbe_json do customers create \
    --name "$CUSTOMER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --can-manage-crew-requirements true

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

test_name "Create prerequisite job production plan"
TODAY=$(date +%Y-%m-%d)
JPP_NAME=$(unique_name "JPPCCPlan")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        register_cleanup "job-production-plans" "$CREATED_JPP_ID"
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

test_name "Create prerequisite cost code"
COST_CODE_VALUE=$(unique_name "JPPCC")
COST_CODE_DESC="Job Production Plan Cost Code $(unique_name "Desc")"

xbe_json do cost-codes create \
    --code "$COST_CODE_VALUE" \
    --description "$COST_CODE_DESC" \
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

test_name "Create prerequisite project resource classifications"
PRC_NAME=$(unique_name "JPPCCResource")
PRC_NAME2=$(unique_name "JPPCCResourceAlt")

xbe_json do project-resource-classifications create \
    --name "$PRC_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-resource-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created project resource classification but no ID returned"
        echo "Cannot continue without a project resource classification"
        run_tests
    fi
else
    fail "Failed to create project resource classification"
    echo "Cannot continue without a project resource classification"
    run_tests
fi

xbe_json do project-resource-classifications create \
    --name "$PRC_NAME2" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID2=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID2" && "$CREATED_CLASSIFICATION_ID2" != "null" ]]; then
        register_cleanup "project-resource-classifications" "$CREATED_CLASSIFICATION_ID2"
        pass
    else
        fail "Created project resource classification but no ID returned"
        echo "Cannot continue without a second project resource classification"
        run_tests
    fi
else
    fail "Failed to create second project resource classification"
    echo "Cannot continue without a second project resource classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan cost code with optional classification"

xbe_json do job-production-plan-cost-codes create \
    --job-production-plan "$CREATED_JPP_ID" \
    --cost-code "$CREATED_COST_CODE_ID" \
    --project-resource-classification "$CREATED_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_COST_CODE_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_COST_CODE_ID" && "$CREATED_JPP_COST_CODE_ID" != "null" ]]; then
        register_cleanup "job-production-plan-cost-codes" "$CREATED_JPP_COST_CODE_ID"
        pass
    else
        fail "Created job production plan cost code but no ID returned"
    fi
else
    fail "Failed to create job production plan cost code"
fi

if [[ -z "$CREATED_JPP_COST_CODE_ID" || "$CREATED_JPP_COST_CODE_ID" == "null" ]]; then
    echo "Cannot continue without a valid job production plan cost code ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job production plan cost code project resource classification"

xbe_json do job-production-plan-cost-codes update "$CREATED_JPP_COST_CODE_ID" \
    --project-resource-classification "$CREATED_CLASSIFICATION_ID2"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan cost code"

xbe_json view job-production-plan-cost-codes show "$CREATED_JPP_COST_CODE_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show job production plan cost code"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan cost codes"

xbe_json view job-production-plan-cost-codes list --limit 5
assert_success

test_name "List job production plan cost codes returns array"

xbe_json view job-production-plan-cost-codes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan cost codes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job production plan cost codes with --job-production-plan filter"

xbe_json view job-production-plan-cost-codes list --job-production-plan "$CREATED_JPP_ID" --limit 10
assert_success

test_name "List job production plan cost codes with --cost-code filter"

xbe_json view job-production-plan-cost-codes list --cost-code "$CREATED_COST_CODE_ID" --limit 10
assert_success

test_name "List job production plan cost codes with --project-resource-classification filter"

xbe_json view job-production-plan-cost-codes list --project-resource-classification "$CREATED_CLASSIFICATION_ID2" --limit 10
assert_success

test_name "List job production plan cost codes with --code filter"

xbe_json view job-production-plan-cost-codes list --code "$COST_CODE_VALUE" --limit 10
assert_success

test_name "List job production plan cost codes with --description filter"

xbe_json view job-production-plan-cost-codes list --description "$COST_CODE_DESC" --limit 10
assert_success

test_name "List job production plan cost codes with --query filter"

xbe_json view job-production-plan-cost-codes list --query "$COST_CODE_VALUE" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List job production plan cost codes with --limit"

xbe_json view job-production-plan-cost-codes list --limit 3
assert_success

test_name "List job production plan cost codes with --offset"

xbe_json view job-production-plan-cost-codes list --limit 3 --offset 1
assert_success

test_name "List job production plan cost codes with pagination (limit + offset)"

xbe_json view job-production-plan-cost-codes list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create job production plan cost code without job production plan fails"

xbe_json do job-production-plan-cost-codes create --cost-code "$CREATED_COST_CODE_ID"
assert_failure

test_name "Create job production plan cost code without cost code fails"

xbe_json do job-production-plan-cost-codes create --job-production-plan "$CREATED_JPP_ID"
assert_failure

test_name "Update without any relationships fails"

xbe_json do job-production-plan-cost-codes update "$CREATED_JPP_COST_CODE_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job production plan cost code requires --confirm flag"

xbe_run do job-production-plan-cost-codes delete "$CREATED_JPP_COST_CODE_ID"
assert_failure

test_name "Delete job production plan cost code with --confirm"

xbe_run do job-production-plan-cost-codes delete "$CREATED_JPP_COST_CODE_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
