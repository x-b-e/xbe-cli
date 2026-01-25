#!/bin/bash
#
# XBE CLI Integration Tests: Geofences
#
# Tests CRUD operations for the geofences resource.
# Geofences represent geographic boundaries.
#
# NOTE: This test requires creating a prerequisite broker.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_GEOFENCE_ID=""
CREATED_BROKER_ID=""

describe "Resource: geofences"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for geofence tests"
BROKER_NAME=$(unique_name "GeofenceTestBroker")

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

test_name "Create geofence with required fields"
GEOFENCE_NAME=$(unique_name "TestGeofence")
# Polygon must have at least 4 points and be closed (first and last points identical)
POLYGON='[[-122.4,37.8],[-122.4,37.7],[-122.3,37.7],[-122.3,37.8],[-122.4,37.8]]'

xbe_json do geofences create \
    --name "$GEOFENCE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --polygon-coordinates "$POLYGON"

if [[ $status -eq 0 ]]; then
    CREATED_GEOFENCE_ID=$(json_get ".id")
    if [[ -n "$CREATED_GEOFENCE_ID" && "$CREATED_GEOFENCE_ID" != "null" ]]; then
        register_cleanup "geofences" "$CREATED_GEOFENCE_ID"
        pass
    else
        fail "Created geofence but no ID returned"
    fi
else
    fail "Failed to create geofence"
fi

# Only continue if we successfully created a geofence
if [[ -z "$CREATED_GEOFENCE_ID" || "$CREATED_GEOFENCE_ID" == "null" ]]; then
    echo "Cannot continue without a valid geofence ID"
    run_tests
fi

test_name "Create geofence with description"
DESC_GEOFENCE_NAME=$(unique_name "DescGeofence")
DESC_POLYGON='[[-122.5,37.9],[-122.5,37.85],[-122.45,37.85],[-122.45,37.9],[-122.5,37.9]]'

xbe_json do geofences create \
    --name "$DESC_GEOFENCE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --polygon-coordinates "$DESC_POLYGON" \
    --description "Test description"

if [[ $status -eq 0 ]]; then
    DESC_GEOFENCE_ID=$(json_get ".id")
    if [[ -n "$DESC_GEOFENCE_ID" && "$DESC_GEOFENCE_ID" != "null" ]]; then
        register_cleanup "geofences" "$DESC_GEOFENCE_ID"
        pass
    else
        fail "Created geofence but no ID returned"
    fi
else
    fail "Failed to create geofence with description"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update geofence --name"
UPDATED_NAME=$(unique_name "UpdatedGeofence")
xbe_json do geofences update "$CREATED_GEOFENCE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update geofence --description"
xbe_json do geofences update "$CREATED_GEOFENCE_ID" --description "Updated description"
assert_success

test_name "Update geofence --status"
xbe_json do geofences update "$CREATED_GEOFENCE_ID" --status "active"
assert_success

test_name "Update geofence --restriction-mode"
# Valid values: all_truckers, custom_truckers
xbe_json do geofences update "$CREATED_GEOFENCE_ID" --restriction-mode "all_truckers"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List geofences"
xbe_json view geofences list --limit 5
assert_success

test_name "List geofences returns array"
xbe_json view geofences list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list geofences"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List geofences with --broker filter"
xbe_json view geofences list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List geofences with --name filter"
xbe_json view geofences list --name "$UPDATED_NAME" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List geofences with --limit"
xbe_json view geofences list --limit 3
assert_success

test_name "List geofences with --offset"
xbe_json view geofences list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete geofence requires --confirm flag"
xbe_run do geofences delete "$CREATED_GEOFENCE_ID"
assert_failure

test_name "Delete geofence with --confirm"
# Create a geofence specifically for deletion
DEL_GEOFENCE_NAME=$(unique_name "DeleteGeofence")
DEL_POLYGON='[[-122.6,38.0],[-122.6,37.95],[-122.55,37.95],[-122.55,38.0],[-122.6,38.0]]'
xbe_json do geofences create \
    --name "$DEL_GEOFENCE_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --polygon-coordinates "$DEL_POLYGON"

if [[ $status -eq 0 ]]; then
    DEL_GEOFENCE_ID=$(json_get ".id")
    if [[ -n "$DEL_GEOFENCE_ID" && "$DEL_GEOFENCE_ID" != "null" ]]; then
        xbe_run do geofences delete "$DEL_GEOFENCE_ID" --confirm
        assert_success
    else
        skip "Could not create geofence for deletion test"
    fi
else
    skip "Could not create geofence for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create geofence without --name fails"
xbe_json do geofences create \
    --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create geofence without --broker fails"
xbe_json do geofences create \
    --name "NobrokerGeofence"
assert_failure

test_name "Update geofence without any fields fails"
xbe_run do geofences update "$CREATED_GEOFENCE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
