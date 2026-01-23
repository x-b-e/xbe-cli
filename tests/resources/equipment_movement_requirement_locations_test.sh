#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Requirement Locations
#
# Tests CRUD operations for the equipment-movement-requirement-locations resource.
#
# NOTE: This test requires creating a prerequisite broker.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_LOCATION_ID=""
CREATED_BROKER_ID=""

describe "Resource: equipment-movement-requirement-locations"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for location tests"
BROKER_NAME=$(unique_name "EquipmentMoveReqLocationBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create location with required fields"
LOCATION_NAME=$(unique_name "EquipmentMoveReqLocation")
LATITUDE="37.7749"
LONGITUDE="-122.4194"

xbe_json do equipment-movement-requirement-locations create \
    --broker "$CREATED_BROKER_ID" \
    --latitude "$LATITUDE" \
    --longitude "$LONGITUDE" \
    --name "$LOCATION_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_LOCATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
        register_cleanup "equipment-movement-requirement-locations" "$CREATED_LOCATION_ID"
        pass
    else
        fail "Created location but no ID returned"
    fi
else
    fail "Failed to create location"
fi

# Only continue if we successfully created a location
if [[ -z "$CREATED_LOCATION_ID" || "$CREATED_LOCATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid location ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update location --name"
UPDATED_NAME=$(unique_name "UpdatedEquipmentMoveReqLocation")
xbe_json do equipment-movement-requirement-locations update "$CREATED_LOCATION_ID" --name "$UPDATED_NAME"
assert_success

UPDATED_LATITUDE="37.7810"
UPDATED_LONGITUDE="-122.4100"

test_name "Update location --latitude and --longitude"
xbe_json do equipment-movement-requirement-locations update "$CREATED_LOCATION_ID" \
    --latitude "$UPDATED_LATITUDE" \
    --longitude "$UPDATED_LONGITUDE"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show location"
xbe_json view equipment-movement-requirement-locations show "$CREATED_LOCATION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List locations"
xbe_json view equipment-movement-requirement-locations list --limit 5
assert_success

test_name "List locations returns array"
xbe_json view equipment-movement-requirement-locations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list locations"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List locations with --broker filter"
xbe_json view equipment-movement-requirement-locations list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List locations with --name filter"
xbe_json view equipment-movement-requirement-locations list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List locations with --q filter"
xbe_json view equipment-movement-requirement-locations list --q "$UPDATED_NAME" --limit 10
assert_success

test_name "List locations with --near filter"
NEAR_FILTER="$UPDATED_LATITUDE|$UPDATED_LONGITUDE|10"
xbe_json view equipment-movement-requirement-locations list --near "$NEAR_FILTER" --limit 10
assert_success

test_name "List locations with --used-after filter"
xbe_json view equipment-movement-requirement-locations list --used-after "2000-01-01T00:00:00Z" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List locations with --limit"
xbe_json view equipment-movement-requirement-locations list --limit 3
assert_success

test_name "List locations with --offset"
xbe_json view equipment-movement-requirement-locations list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete location requires --confirm flag"
xbe_run do equipment-movement-requirement-locations delete "$CREATED_LOCATION_ID"
assert_failure

# Create a location specifically for deletion
DEL_LOCATION_NAME=$(unique_name "DeleteEquipmentMoveReqLocation")

test_name "Delete location with --confirm"
xbe_json do equipment-movement-requirement-locations create \
    --broker "$CREATED_BROKER_ID" \
    --latitude "37.7900" \
    --longitude "-122.4300" \
    --name "$DEL_LOCATION_NAME"

if [[ $status -eq 0 ]]; then
    DEL_LOCATION_ID=$(json_get ".id")
    if [[ -n "$DEL_LOCATION_ID" && "$DEL_LOCATION_ID" != "null" ]]; then
        xbe_run do equipment-movement-requirement-locations delete "$DEL_LOCATION_ID" --confirm
        assert_success
    else
        skip "Could not create location for deletion test"
    fi
else
    skip "Could not create location for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create location without --broker fails"
xbe_json do equipment-movement-requirement-locations create \
    --latitude "37.0" \
    --longitude "-122.0"
assert_failure

test_name "Create location without --latitude fails"
xbe_json do equipment-movement-requirement-locations create \
    --broker "$CREATED_BROKER_ID" \
    --longitude "-122.0"
assert_failure

test_name "Create location without --longitude fails"
xbe_json do equipment-movement-requirement-locations create \
    --broker "$CREATED_BROKER_ID" \
    --latitude "37.0"
assert_failure

test_name "Update location without any fields fails"
xbe_run do equipment-movement-requirement-locations update "$CREATED_LOCATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
