#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Trucking Incident Detectors
#
# Tests CRUD operations for the job_production_plan_trucking_incident_detectors resource.
#
# COVERAGE: All filters + all create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JPP_ID=""
CREATED_DETECTOR_ID=""

describe "Resource: job-production-plan-trucking-incident-detectors"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker for trucking incident detector tests"
BROKER_NAME=$(unique_name "JPPTruckingDetectorBroker")

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
CUSTOMER_NAME=$(unique_name "JPPTruckingDetectorCustomer")

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
JPP_NAME=$(unique_name "JPPTruckingDetectorPlan")

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

test_name "Create trucking incident detector"
AS_OF=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

xbe_json do job-production-plan-trucking-incident-detectors create \
    --job-production-plan "$CREATED_JPP_ID" \
    --as-of "$AS_OF" \
    --persist-changes=false

if [[ $status -eq 0 ]]; then
    CREATED_DETECTOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_DETECTOR_ID" && "$CREATED_DETECTOR_ID" != "null" ]]; then
        pass
    else
        fail "Created detector but no ID returned"
    fi
else
    fail "Failed to create trucking incident detector"
fi

if [[ -z "$CREATED_DETECTOR_ID" || "$CREATED_DETECTOR_ID" == "null" ]]; then
    echo "Cannot continue without a valid detector ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucking incident detector"

xbe_json view job-production-plan-trucking-incident-detectors show "$CREATED_DETECTOR_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show trucking incident detector"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucking incident detectors"

xbe_json view job-production-plan-trucking-incident-detectors list --limit 5
assert_success

test_name "List trucking incident detectors returns array"

xbe_json view job-production-plan-trucking-incident-detectors list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucking incident detectors"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trucking incident detectors with --job-production-plan filter"

xbe_json view job-production-plan-trucking-incident-detectors list --job-production-plan "$CREATED_JPP_ID" --limit 5
assert_success

test_name "List trucking incident detectors with --is-performed filter (true)"

xbe_json view job-production-plan-trucking-incident-detectors list --is-performed true --limit 5
assert_success

test_name "List trucking incident detectors with --is-performed filter (false)"

xbe_json view job-production-plan-trucking-incident-detectors list --is-performed false --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create detector without job production plan fails"

xbe_json do job-production-plan-trucking-incident-detectors create --as-of "$AS_OF"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
