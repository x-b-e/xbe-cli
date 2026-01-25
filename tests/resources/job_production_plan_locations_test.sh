#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Locations
#
# Tests CRUD operations for the job-production-plan-locations resource.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_JOB_PRODUCTION_PLAN_ID=""
CREATED_LOCATION_ID=""
SAMPLE_LOCATION_ID=""
SAMPLE_SEGMENT_ID=""
SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=""
SKIP_CRUD_TESTS=false

describe "Resource: job-production-plan-locations"

# ============================================================================
# Prerequisites - Create broker, customer, and job production plan
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "JPP-Location-Broker")

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
CUSTOMER_NAME=$(unique_name "JPP-Location-Customer")

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
PLAN_NAME=$(unique_name "JPP-Location")
TODAY=$(date +%Y-%m-%d)

xbe_json do job-production-plans create \
    --job-name "$PLAN_NAME" \
    --start-on "$TODAY" \
    --start-time "07:00" \
    --customer "$CREATED_CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_JOB_PRODUCTION_PLAN_ID=$(json_get ".id")
    if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
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

test_name "Create job production plan location"
LOCATION_NAME=$(unique_name "JPP-Location")

xbe_json do job-production-plan-locations create \
    --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" \
    --name "$LOCATION_NAME" \
    --site-kind job_site \
    --is-start-site-candidate \
    --address "100 Test St, Chicago, IL" \
    --address-latitude "41.8781" \
    --address-longitude "-87.6298" \
    --skip-address-geocoding

if [[ $status -eq 0 ]]; then
    CREATED_LOCATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
        register_cleanup "job-production-plan-locations" "$CREATED_LOCATION_ID"
        pass
    else
        fail "Created job production plan location but no ID returned"
    fi
else
    fail "Failed to create job production plan location"
fi

if [[ -z "$CREATED_LOCATION_ID" || "$CREATED_LOCATION_ID" == "null" ]]; then
    SKIP_CRUD_TESTS=true
    echo "Cannot continue CRUD tests without a valid location ID"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ "$SKIP_CRUD_TESTS" != "true" ]]; then

test_name "Update location name"
xbe_json do job-production-plan-locations update "$CREATED_LOCATION_ID" --name "${LOCATION_NAME}-Updated"
assert_success

test_name "Update site kind"
xbe_json do job-production-plan-locations update "$CREATED_LOCATION_ID" --site-kind other
assert_success

test_name "Clear start site candidate flag"
xbe_json do job-production-plan-locations update "$CREATED_LOCATION_ID" --no-is-start-site-candidate
assert_success

test_name "Update address coordinates"
xbe_json do job-production-plan-locations update "$CREATED_LOCATION_ID" \
    --address-latitude "41.8900" \
    --address-longitude "-87.6200" \
    --skip-address-geocoding
assert_success

test_name "Update address place ID and plus code"
xbe_json do job-production-plan-locations update "$CREATED_LOCATION_ID" \
    --address-place-id "ChIJ-CLI-Test" \
    --address-plus-code "86HJ+XX"
assert_success

fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan locations"
xbe_json view job-production-plan-locations list --limit 5
assert_success

test_name "List job production plan locations returns array"
xbe_json view job-production-plan-locations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan locations"
fi

# ==========================================================================
# Sample Record (used for show + filters)
# ==========================================================================

test_name "Capture sample job production plan location"
xbe_json view job-production-plan-locations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_LOCATION_ID=$(json_get ".[0].id")
    SAMPLE_SEGMENT_ID=$(json_get ".[0].segment_id")
    if [[ -n "$SAMPLE_LOCATION_ID" && "$SAMPLE_LOCATION_ID" != "null" ]]; then
        pass
    else
        skip "No job production plan locations available for follow-on tests"
    fi
else
    skip "Could not list job production plan locations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job production plan locations with --job-production-plan filter"
if [[ -n "$CREATED_JOB_PRODUCTION_PLAN_ID" && "$CREATED_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view job-production-plan-locations list --job-production-plan "$CREATED_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List job production plan locations with --segment filter"
if [[ -n "$SAMPLE_SEGMENT_ID" && "$SAMPLE_SEGMENT_ID" != "null" ]]; then
    xbe_json view job-production-plan-locations list --segment "$SAMPLE_SEGMENT_ID" --limit 5
    assert_success
else
    skip "No segment ID available"
fi

test_name "List job production plan locations with --broker-tender-job-schedule-shift filter"
# Try to find a tender job schedule shift ID from shift feedbacks
xbe_json view shift-feedbacks list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID=$(json_get ".[0].tender_job_schedule_shift_id")
fi

if [[ -n "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" && "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view job-production-plan-locations list \
        --broker-tender-job-schedule-shift "$SAMPLE_TENDER_JOB_SCHEDULE_SHIFT_ID" \
        --limit 5
    assert_success
else
    skip "No tender job schedule shift ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan location"
if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
    xbe_json view job-production-plan-locations show "$CREATED_LOCATION_ID"
    assert_success
elif [[ -n "$SAMPLE_LOCATION_ID" && "$SAMPLE_LOCATION_ID" != "null" ]]; then
    xbe_json view job-production-plan-locations show "$SAMPLE_LOCATION_ID"
    assert_success
else
    skip "No job production plan location ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ "$SKIP_CRUD_TESTS" != "true" && -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then

test_name "Delete job production plan location"
xbe_run do job-production-plan-locations delete "$CREATED_LOCATION_ID" --confirm
assert_success

fi

# ============================================================================
# Summary
# ============================================================================

run_tests
