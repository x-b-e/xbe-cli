#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Abandonments
#
# Tests create operations for the job_production_plan_abandonments resource.
#
# COVERAGE: Create + list + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""

describe "Resource: job-production-plan-abandonments"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPPAbandonBroker")

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
CUSTOMER_NAME=$(unique_name "JPPAbandonCustomer")

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

test_name "Create job production plan for abandonment"
TEST_NAME=$(unique_name "JPPAbandon")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$TEST_NAME" \
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
    fail "Failed to create job production plan"
    echo "Cannot continue without a job production plan"
    run_tests
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create abandonment without required job production plan fails"
xbe_run do job-production-plan-abandonments create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan abandonment"
COMMENT="Abandoning plan for test"

xbe_json do job-production-plan-abandonments create \
    --job-production-plan "$CREATED_JPP_ID" \
    --comment "$COMMENT" \
    --suppress-status-change-notifications

if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_equals ".job_production_plan_id" "$CREATED_JPP_ID"
    assert_json_equals ".comment" "$COMMENT"
    assert_json_bool ".suppress_status_change_notifications" "true"
else
    fail "Failed to create job production plan abandonment"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
