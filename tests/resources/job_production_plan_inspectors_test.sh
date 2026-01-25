#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Inspectors
#
# Tests CRUD operations for the job_production_plan_inspectors resource.
#
# COVERAGE: All filters + all create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
INSPECTOR_USER_ID=""
CREATED_INSPECTOR_ID=""

describe "Resource: job-production-plan-inspectors"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for inspector tests"
BROKER_NAME=$(unique_name "JPPInspectorBroker")

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
CUSTOMER_NAME=$(unique_name "JPPInspectorCustomer")

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
JPP_NAME=$(unique_name "JPPInspectorPlan")

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID" \
    --explicit-requires-inspector false

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

test_name "Resolve current user for inspector tests"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    INSPECTOR_USER_ID=$(json_get ".id")
    if [[ -n "$INSPECTOR_USER_ID" && "$INSPECTOR_USER_ID" != "null" ]]; then
        pass
    else
        fail "Resolved user but no ID returned"
        echo "Cannot continue without a user ID"
        run_tests
    fi
else
    fail "Failed to resolve current user"
    echo "Cannot continue without a user ID"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan inspector"

xbe_json do job-production-plan-inspectors create \
    --job-production-plan-id "$CREATED_JPP_ID" \
    --user "$INSPECTOR_USER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_INSPECTOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_INSPECTOR_ID" && "$CREATED_INSPECTOR_ID" != "null" ]]; then
        register_cleanup "job-production-plan-inspectors" "$CREATED_INSPECTOR_ID"
        pass
    else
        fail "Created inspector but no ID returned"
    fi
else
    fail "Failed to create job production plan inspector"
fi

if [[ -z "$CREATED_INSPECTOR_ID" || "$CREATED_INSPECTOR_ID" == "null" ]]; then
    echo "Cannot continue without a valid inspector ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan inspector"
xbe_json view job-production-plan-inspectors show "$CREATED_INSPECTOR_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show job production plan inspector"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan inspectors"
xbe_json view job-production-plan-inspectors list --limit 5
assert_success

test_name "List job production plan inspectors returns array"
xbe_json view job-production-plan-inspectors list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan inspectors"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter inspectors by job production plan"
xbe_json view job-production-plan-inspectors list --job-production-plan-id "$CREATED_JPP_ID"
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$CREATED_INSPECTOR_ID" '.[] | select(.id == $id)' > /dev/null; then
        pass
    else
        fail "Filtered list missing created inspector"
    fi
else
    fail "Failed to filter by job production plan"
fi

test_name "Filter inspectors by user"
xbe_json view job-production-plan-inspectors list --user "$INSPECTOR_USER_ID"
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$CREATED_INSPECTOR_ID" '.[] | select(.id == $id)' > /dev/null; then
        pass
    else
        fail "Filtered list missing created inspector"
    fi
else
    fail "Failed to filter by user"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create inspector without user fails"
xbe_json do job-production-plan-inspectors create --job-production-plan-id "$CREATED_JPP_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete job production plan inspector"
xbe_run do job-production-plan-inspectors delete "$CREATED_INSPECTOR_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
