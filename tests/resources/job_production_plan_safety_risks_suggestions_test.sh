#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Safety Risks Suggestions
#
# Tests CRUD operations for the job_production_plan_safety_risks_suggestions resource.
#
# COVERAGE: All filters + all create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_SUGGESTION_ID=""

describe "Resource: job-production-plan-safety-risks-suggestions"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for safety risks suggestion tests"
BROKER_NAME=$(unique_name "JPPSafetyRisksBroker")

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
CUSTOMER_NAME=$(unique_name "JPPSafetyRisksCustomer")

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
JPP_NAME=$(unique_name "JPPSafetyRisksPlan")

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

test_name "Create safety risks suggestion"
OPTIONS_JSON='{"include_other_incidents":true}'

xbe_json do job-production-plan-safety-risks-suggestions create \
    --job-production-plan "$CREATED_JPP_ID" \
    --options "$OPTIONS_JSON" \
    --is-async=true

if [[ $status -eq 0 ]]; then
    CREATED_SUGGESTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_SUGGESTION_ID" && "$CREATED_SUGGESTION_ID" != "null" ]]; then
        pass
    else
        fail "Created suggestion but no ID returned"
    fi
else
    fail "Failed to create safety risks suggestion"
fi

if [[ -z "$CREATED_SUGGESTION_ID" || "$CREATED_SUGGESTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid suggestion ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show safety risks suggestion"

xbe_json view job-production-plan-safety-risks-suggestions show "$CREATED_SUGGESTION_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show safety risks suggestion"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List safety risks suggestions"

xbe_json view job-production-plan-safety-risks-suggestions list --limit 5
assert_success

test_name "List safety risks suggestions returns array"

xbe_json view job-production-plan-safety-risks-suggestions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list safety risks suggestions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter safety risks suggestions by job production plan"

xbe_json view job-production-plan-safety-risks-suggestions list --job-production-plan "$CREATED_JPP_ID"
if [[ $status -eq 0 ]]; then
    if echo "$output" | jq -e --arg id "$CREATED_SUGGESTION_ID" '.[] | select(.id == $id)' > /dev/null; then
        pass
    else
        fail "Filtered list missing created suggestion"
    fi
else
    fail "Failed to filter by job production plan"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create suggestion without job production plan fails"

xbe_json do job-production-plan-safety-risks-suggestions create --options "$OPTIONS_JSON"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
