#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Safety Risks
#
# Tests list, show, create, update, and delete operations for
# job_production_plan_safety_risks.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_RISK_ID=""
SAMPLE_ID=""
SAMPLE_JPP_ID=""

describe "Resource: job-production-plan-safety-risks"

# ============================================================================
# Prerequisites - Create broker, customer, job plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPSafetyRiskBroker")

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

test_name "Create prerequisite customer"
CUSTOMER_NAME=$(unique_name "JPPSafetyRiskCustomer")

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
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Create prerequisite job production plan"
PLAN_NAME=$(unique_name "JPPSafetyRiskPlan")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$PLAN_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JPP_ID=$(json_get ".id")
    if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
        pass
    else
        fail "Created job production plan but no ID returned"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_JOB_PRODUCTION_PLAN_ID" ]]; then
        CREATED_JPP_ID="$XBE_TEST_JOB_PRODUCTION_PLAN_ID"
        echo "    Using XBE_TEST_JOB_PRODUCTION_PLAN_ID: $CREATED_JPP_ID"
        pass
    else
        fail "Failed to create job production plan and XBE_TEST_JOB_PRODUCTION_PLAN_ID not set"
        echo "Cannot continue without a job production plan"
        run_tests
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan safety risks"
xbe_json view job-production-plan-safety-risks list --limit 5
assert_success

test_name "List job production plan safety risks returns array"
xbe_json view job-production-plan-safety-risks list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_JPP_ID=$(echo "$output" | jq -r '.[0].job_production_plan_id // empty')
else
    fail "Failed to list job production plan safety risks"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan safety risk"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-safety-risks show "$SAMPLE_ID"
    assert_success
else
    skip "No safety risk ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan safety risk"
DESCRIPTION=$(unique_name "JPPSafetyRisk")

xbe_json do job-production-plan-safety-risks create \
    --job-production-plan "$CREATED_JPP_ID" \
    --description "$DESCRIPTION"

if [[ $status -eq 0 ]]; then
    CREATED_RISK_ID=$(json_get ".id")
    if [[ -n "$CREATED_RISK_ID" && "$CREATED_RISK_ID" != "null" ]]; then
        register_cleanup "job-production-plan-safety-risks" "$CREATED_RISK_ID"
        pass
    else
        fail "Created safety risk but no ID returned"
    fi
else
    fail "Failed to create job production plan safety risk"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update job production plan safety risk description"
if [[ -n "$CREATED_RISK_ID" && "$CREATED_RISK_ID" != "null" ]]; then
    UPDATED_DESCRIPTION=$(unique_name "JPPSafetyRiskUpdated")
    xbe_json do job-production-plan-safety-risks update "$CREATED_RISK_ID" --description "$UPDATED_DESCRIPTION"
    assert_success
else
    skip "No safety risk created for update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List safety risks with --job-production-plan filter"
FILTER_JPP_ID="${CREATED_JPP_ID:-$SAMPLE_JPP_ID}"
if [[ -n "$FILTER_JPP_ID" && "$FILTER_JPP_ID" != "null" ]]; then
    xbe_json view job-production-plan-safety-risks list --job-production-plan "$FILTER_JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete safety risk requires --confirm flag"
if [[ -n "$CREATED_RISK_ID" && "$CREATED_RISK_ID" != "null" ]]; then
    xbe_run do job-production-plan-safety-risks delete "$CREATED_RISK_ID"
    assert_failure
else
    skip "No created safety risk for delete confirmation test"
fi

test_name "Delete safety risk with --confirm"
if [[ -n "$CREATED_RISK_ID" && "$CREATED_RISK_ID" != "null" ]]; then
    xbe_run do job-production-plan-safety-risks delete "$CREATED_RISK_ID" --confirm
    assert_success
else
    skip "No created safety risk to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create safety risk without job production plan fails"
DESCRIPTION=$(unique_name "JPPSafetyRiskMissingPlan")
xbe_run do job-production-plan-safety-risks create --description "$DESCRIPTION"
assert_failure

test_name "Create safety risk without description fails"
if [[ -n "$CREATED_JPP_ID" && "$CREATED_JPP_ID" != "null" ]]; then
    xbe_run do job-production-plan-safety-risks create --job-production-plan "$CREATED_JPP_ID"
    assert_failure
else
    skip "No job production plan ID available for missing description test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
