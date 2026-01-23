#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Alarms
#
# Tests CRUD operations for the job_production_plan_alarms resource.
#
# COVERAGE: All filters + all create/update attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_ALARM_ID=""

describe "Resource: job-production-plan-alarms"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for alarm tests"
BROKER_NAME=$(unique_name "JPPAlarmBroker")

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
CUSTOMER_NAME=$(unique_name "JPPAlarmCustomer")

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
JPP_NAME=$(unique_name "JPPAlarmPlan")

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

test_name "Create alarm with required fields"

xbe_json do job-production-plan-alarms create \
    --job-production-plan "$CREATED_JPP_ID" \
    --tons 150

if [[ $status -eq 0 ]]; then
    CREATED_ALARM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ALARM_ID" && "$CREATED_ALARM_ID" != "null" ]]; then
        register_cleanup "job-production-plan-alarms" "$CREATED_ALARM_ID"
        pass
    else
        fail "Created alarm but no ID returned"
    fi
else
    fail "Failed to create alarm"
fi

if [[ -z "$CREATED_ALARM_ID" || "$CREATED_ALARM_ID" == "null" ]]; then
    echo "Cannot continue without a valid alarm ID"
    run_tests
fi

test_name "Create alarm with optional fields"

xbe_json do job-production-plan-alarms create \
    --job-production-plan "$CREATED_JPP_ID" \
    --tons 200 \
    --base-material-type-fully-qualified-name "Asphalt Mixture" \
    --max-latency-minutes 45 \
    --note "Alarm for 200 tons"

if [[ $status -eq 0 ]]; then
    OPTIONAL_ALARM_ID=$(json_get ".id")
    if [[ -n "$OPTIONAL_ALARM_ID" && "$OPTIONAL_ALARM_ID" != "null" ]]; then
        register_cleanup "job-production-plan-alarms" "$OPTIONAL_ALARM_ID"
        pass
    else
        fail "Created alarm but no ID returned"
    fi
else
    fail "Failed to create alarm with optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update alarm tons"
xbe_json do job-production-plan-alarms update "$CREATED_ALARM_ID" --tons 175
assert_success

test_name "Update alarm base material type"
xbe_json do job-production-plan-alarms update "$CREATED_ALARM_ID" \
    --base-material-type-fully-qualified-name "Concrete"
assert_success

test_name "Update alarm max latency minutes"
xbe_json do job-production-plan-alarms update "$CREATED_ALARM_ID" --max-latency-minutes 60
assert_success

test_name "Update alarm note"
xbe_json do job-production-plan-alarms update "$CREATED_ALARM_ID" --note "Updated alarm note"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show alarm"

xbe_json view job-production-plan-alarms show "$CREATED_ALARM_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show alarm"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List alarms"

xbe_json view job-production-plan-alarms list --limit 5
assert_success

test_name "List alarms returns array"

xbe_json view job-production-plan-alarms list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list alarms"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List alarms with --job-production-plan filter"

xbe_json view job-production-plan-alarms list --job-production-plan "$CREATED_JPP_ID" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List alarms with --limit"

xbe_json view job-production-plan-alarms list --limit 3
assert_success

test_name "List alarms with --offset"

xbe_json view job-production-plan-alarms list --limit 3 --offset 1
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create alarm without job production plan fails"

xbe_json do job-production-plan-alarms create --tons 150
assert_failure

test_name "Create alarm without tons fails"

xbe_json do job-production-plan-alarms create --job-production-plan "$CREATED_JPP_ID"
assert_failure

test_name "Update alarm without fields fails"

xbe_json do job-production-plan-alarms update "$CREATED_ALARM_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete alarm requires --confirm flag"

xbe_run do job-production-plan-alarms delete "$CREATED_ALARM_ID"
assert_failure

test_name "Delete alarm with --confirm"

xbe_run do job-production-plan-alarms delete "$CREATED_ALARM_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
