#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Safety Risk Communication Suggestions
#
# Tests create/view/delete operations for job production plan safety risk communication suggestions.
#
# COVERAGE: Create attributes (job-production-plan, options, is-async), list filters + show, delete.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_SUGGESTION_ID=""

describe "Resource: job-production-plan-safety-risk-communication-suggestions"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for suggestion tests"
BROKER_NAME=$(unique_name "JPPSafetyRiskCommBroker")

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
CUSTOMER_NAME=$(unique_name "JPPSafetyRiskCommCustomer")

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

test_name "Create job production plan"
JPP_NAME=$(unique_name "JPPSafetyRiskComm")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$JPP_NAME" \
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
# CREATE Tests
# ============================================================================

test_name "Create suggestion with required fields"
xbe_json do job-production-plan-safety-risk-communication-suggestions create \
    --job-production-plan "$CREATED_JPP_ID" \
    --is-async=true \
    --options '{"temperature":0.2}'

if [[ $status -eq 0 ]]; then
    CREATED_SUGGESTION_ID=$(json_get ".id")
    if [[ -n "$CREATED_SUGGESTION_ID" && "$CREATED_SUGGESTION_ID" != "null" ]]; then
        pass
    else
        fail "Created suggestion but no ID returned"
    fi
else
    fail "Failed to create suggestion"
fi

# Only continue if we successfully created a suggestion
if [[ -z "$CREATED_SUGGESTION_ID" || "$CREATED_SUGGESTION_ID" == "null" ]]; then
    echo "Cannot continue without a valid suggestion ID"
    run_tests
fi

test_name "Create suggestion without --job-production-plan fails"
xbe_json do job-production-plan-safety-risk-communication-suggestions create
assert_failure

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show suggestion"
xbe_json view job-production-plan-safety-risk-communication-suggestions show "$CREATED_SUGGESTION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List suggestions"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --limit 5
assert_success

test_name "List suggestions returns array"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list suggestions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List suggestions with --job-production-plan filter"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --job-production-plan "$CREATED_JPP_ID" --limit 5
assert_success

test_name "List suggestions with --created-at-min filter"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List suggestions with --created-at-max filter"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List suggestions with --updated-at-min filter"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List suggestions with --updated-at-max filter"
xbe_json view job-production-plan-safety-risk-communication-suggestions list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete suggestion requires --confirm flag"
xbe_run do job-production-plan-safety-risk-communication-suggestions delete "$CREATED_SUGGESTION_ID"
assert_failure

test_name "Delete suggestion with --confirm"
xbe_run do job-production-plan-safety-risk-communication-suggestions delete "$CREATED_SUGGESTION_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
