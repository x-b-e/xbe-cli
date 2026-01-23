#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Trips
#
# Tests list, show, create, update, and delete operations for the
# equipment-movement-trips resource.
#
# COVERAGE: List filters + show + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRIP_ID=""
CREATED_BROKER_ID=""
TRAILER_CLASSIFICATION_ID=""
TRAILER_CLASSIFICATION_EQUIV_ID=""
STUOM_ID=""
SAMPLE_JPP_ID=""

describe "Resource: equipment-movement-trips"

# ============================================================================
# Prerequisites - Broker
# ============================================================================

test_name "Create prerequisite broker for equipment movement trip tests"
BROKER_NAME=$(unique_name "EquipMoveBroker")

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

test_name "Capture trailer classification ID"
xbe_json view trailer-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    TRAILER_CLASSIFICATION_ID=$(echo "$output" | jq -r '[.[] | select(.is_heavy_equipment_transport == true)][0].id')
    TRAILER_CLASSIFICATION_EQUIV_ID=$(echo "$output" | jq -r '[.[] | select(.is_heavy_equipment_transport == true)][1].id')
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        pass
    else
        skip "No heavy equipment transport trailer classifications available"
    fi
else
    skip "Failed to list trailer classifications"
fi

test_name "Capture service type unit of measure ID from rates"
xbe_json view rates list --limit 10
if [[ $status -eq 0 ]]; then
    STUOM_ID=$(echo "$output" | jq -r '[.[] | select(.service_type_unit_of_measure_id != null)][0].service_type_unit_of_measure_id')
    if [[ -n "$STUOM_ID" && "$STUOM_ID" != "null" ]]; then
        pass
    else
        skip "No service type unit of measure IDs available in rates"
    fi
else
    skip "Failed to list rates"
fi

test_name "Capture job production plan ID"
xbe_json view job-production-plans list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_JPP_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_JPP_ID" && "$SAMPLE_JPP_ID" != "null" ]]; then
        pass
    else
        skip "No job production plans available"
    fi
else
    skip "Failed to list job production plans"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List equipment movement trips"
xbe_json view equipment-movement-trips list --limit 5
assert_success

test_name "List equipment movement trips returns array"
xbe_json view equipment-movement-trips list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment movement trips"
fi

test_name "List trips with --broker filter"
xbe_json view equipment-movement-trips list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List trips with --max-origin-at-min-min filter"
xbe_json view equipment-movement-trips list --max-origin-at-min-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trips with --max-origin-at-min-max filter"
xbe_json view equipment-movement-trips list --max-origin-at-min-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List trips with --job-production-plan filter"
if [[ -n "$SAMPLE_JPP_ID" && "$SAMPLE_JPP_ID" != "null" ]]; then
    xbe_json view equipment-movement-trips list --job-production-plan "$SAMPLE_JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List trips with --job-production-plan-id filter"
if [[ -n "$SAMPLE_JPP_ID" && "$SAMPLE_JPP_ID" != "null" ]]; then
    xbe_json view equipment-movement-trips list --job-production-plan-id "$SAMPLE_JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment movement trip"
JOB_NUMBER=$(unique_name "EMT")

xbe_json do equipment-movement-trips create \
    --broker "$CREATED_BROKER_ID" \
    --job-number "$JOB_NUMBER"

if [[ $status -eq 0 ]]; then
    CREATED_TRIP_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRIP_ID" && "$CREATED_TRIP_ID" != "null" ]]; then
        register_cleanup "equipment-movement-trips" "$CREATED_TRIP_ID"
        pass
    else
        fail "Created trip but no ID returned"
    fi
else
    fail "Failed to create equipment movement trip"
fi

# Only continue if we successfully created a trip
if [[ -z "$CREATED_TRIP_ID" || "$CREATED_TRIP_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment movement trip ID"
    run_tests
fi

test_name "Create trip without broker fails"
xbe_run do equipment-movement-trips create
assert_failure

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment movement trip"
xbe_json view equipment-movement-trips show "$CREATED_TRIP_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trip --job-number"
NEW_JOB_NUMBER=$(unique_name "EMT-Updated")
xbe_json do equipment-movement-trips update "$CREATED_TRIP_ID" --job-number "$NEW_JOB_NUMBER"
assert_success

test_name "Update trip --explicit-driver-day-mobilization-before-minutes"
xbe_json do equipment-movement-trips update "$CREATED_TRIP_ID" --explicit-driver-day-mobilization-before-minutes 30
assert_success

test_name "Update trip --trailer-classification"
if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do equipment-movement-trips update "$CREATED_TRIP_ID" --trailer-classification "$TRAILER_CLASSIFICATION_ID"
    assert_success
else
    skip "No trailer classification ID available"
fi

test_name "Update trip --trailer-classification-equivalent-ids"
if [[ -n "$TRAILER_CLASSIFICATION_EQUIV_ID" && "$TRAILER_CLASSIFICATION_EQUIV_ID" != "null" ]]; then
    xbe_json do equipment-movement-trips update "$CREATED_TRIP_ID" --trailer-classification-equivalent-ids "$TRAILER_CLASSIFICATION_EQUIV_ID"
    assert_success
else
    skip "No second heavy equipment transport trailer classification ID available"
fi

test_name "Update trip --service-type-unit-of-measure-ids"
if [[ -n "$STUOM_ID" && "$STUOM_ID" != "null" ]]; then
    xbe_json do equipment-movement-trips update "$CREATED_TRIP_ID" --service-type-unit-of-measure-ids "$STUOM_ID"
    assert_success
else
    skip "No service type unit of measure ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment movement trip"
xbe_run do equipment-movement-trips delete "$CREATED_TRIP_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
