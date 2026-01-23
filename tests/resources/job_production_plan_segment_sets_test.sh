#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Segment Sets
#
# Tests CRUD operations for the job_production_plan_segment_sets resource.
#
# COVERAGE: All filters + all create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_SEGMENT_SET_ID=""

describe "Resource: job-production-plan-segment-sets"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for segment set tests"
BROKER_NAME=$(unique_name "JPPSSBroker")

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
CUSTOMER_NAME=$(unique_name "JPPSSCustomer")

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
JPP_NAME=$(unique_name "JPPSegmentSetPlan")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create segment set with optional fields"

xbe_json do job-production-plan-segment-sets create \
    --job-production-plan "$CREATED_JPP_ID" \
    --name "AM Shift" \
    --is-default=false \
    --start-offset-minutes 15

if [[ $status -eq 0 ]]; then
    CREATED_SEGMENT_SET_ID=$(json_get ".id")
    if [[ -n "$CREATED_SEGMENT_SET_ID" && "$CREATED_SEGMENT_SET_ID" != "null" ]]; then
        register_cleanup "job-production-plan-segment-sets" "$CREATED_SEGMENT_SET_ID"
        pass
    else
        fail "Created segment set but no ID returned"
    fi
else
    fail "Failed to create segment set"
fi

if [[ -z "$CREATED_SEGMENT_SET_ID" || "$CREATED_SEGMENT_SET_ID" == "null" ]]; then
    echo "Cannot continue without a valid segment set ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show segment set"
xbe_json view job-production-plan-segment-sets show "$CREATED_SEGMENT_SET_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update segment set name"
xbe_json do job-production-plan-segment-sets update "$CREATED_SEGMENT_SET_ID" --name "PM Shift"
assert_success

test_name "Update segment set start offset"
xbe_json do job-production-plan-segment-sets update "$CREATED_SEGMENT_SET_ID" --start-offset-minutes 30
assert_success

test_name "Update segment set default flag"
xbe_json do job-production-plan-segment-sets update "$CREATED_SEGMENT_SET_ID" --is-default=false
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List segment sets with --job-production-plan filter"
xbe_json view job-production-plan-segment-sets list --job-production-plan "$CREATED_JPP_ID" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List segment sets with --limit"
xbe_json view job-production-plan-segment-sets list --limit 3
assert_success

test_name "List segment sets with --offset"
xbe_json view job-production-plan-segment-sets list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create segment set without job production plan fails"
xbe_json do job-production-plan-segment-sets create --name "No Plan"
assert_failure

test_name "Update segment set without fields fails"
xbe_json do job-production-plan-segment-sets update "$CREATED_SEGMENT_SET_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete segment set requires --confirm flag"
xbe_run do job-production-plan-segment-sets delete "$CREATED_SEGMENT_SET_ID"
assert_failure

test_name "Delete segment set with --confirm"
xbe_run do job-production-plan-segment-sets delete "$CREATED_SEGMENT_SET_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
