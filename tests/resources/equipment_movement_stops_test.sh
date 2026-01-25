#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Stops
#
# Tests CRUD operations for the equipment-movement-stops resource.
#
# NOTE: This test requires an equipment movement trip ID and broker ID.
#
# COVERAGE: All create/update attributes + list filters
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_STOP_ID=""
CREATED_LOCATION_ID=""
UPDATED_LOCATION_ID=""

TRIP_ID="${XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID:-}"
BROKER_ID="${XBE_TEST_EQUIPMENT_MOVEMENT_BROKER_ID:-}"

describe "Resource: equipment-movement-stops"

# ============================================================================
# LIST Tests - Basic (no prerequisites)
# ============================================================================

test_name "List stops"
xbe_json view equipment-movement-stops list --limit 5
assert_success

test_name "List stops returns array"
xbe_json view equipment-movement-stops list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list stops"
fi

# ============================================================================
# CREATE/UPDATE/SHOW/DELETE Tests (requires trip + broker)
# ============================================================================

if [[ -z "$TRIP_ID" || -z "$BROKER_ID" ]]; then
    test_name "Skip create/update/delete tests (missing trip/broker)"
    skip "Set XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID and XBE_TEST_EQUIPMENT_MOVEMENT_BROKER_ID to run full tests"
else
    # ------------------------------------------------------------------------
    # Prerequisite - create location for stop
    # ------------------------------------------------------------------------

    test_name "Create location for stop"
    LOCATION_NAME=$(unique_name "EquipmentMoveStopLocation")
    xbe_json do equipment-movement-requirement-locations create \
        --broker "$BROKER_ID" \
        --latitude "37.7749" \
        --longitude "-122.4194" \
        --name "$LOCATION_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_LOCATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
            register_cleanup "equipment-movement-requirement-locations" "$CREATED_LOCATION_ID"
            pass
        else
            fail "Created location but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create location"
        run_tests
    fi

    # ------------------------------------------------------------------------
    # CREATE Tests
    # ------------------------------------------------------------------------

    test_name "Create stop with required fields"
    SCHEDULED_ARRIVAL="2025-01-01T08:00:00Z"
    xbe_json do equipment-movement-stops create \
        --trip "$TRIP_ID" \
        --location "$CREATED_LOCATION_ID" \
        --sequence-position 1 \
        --scheduled-arrival-at "$SCHEDULED_ARRIVAL"

    if [[ $status -eq 0 ]]; then
        CREATED_STOP_ID=$(json_get ".id")
        if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
            register_cleanup "equipment-movement-stops" "$CREATED_STOP_ID"
            pass
        else
            fail "Created stop but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create stop"
        run_tests
    fi

    # ------------------------------------------------------------------------
    # UPDATE Tests
    # ------------------------------------------------------------------------

    test_name "Update stop --sequence-position"
    xbe_json do equipment-movement-stops update "$CREATED_STOP_ID" --sequence-position 2
    assert_success

    test_name "Create location for stop update"
    UPDATED_LOCATION_NAME=$(unique_name "EquipmentMoveStopLocationUpdate")
    xbe_json do equipment-movement-requirement-locations create \
        --broker "$BROKER_ID" \
        --latitude "37.7810" \
        --longitude "-122.4100" \
        --name "$UPDATED_LOCATION_NAME"

    if [[ $status -eq 0 ]]; then
        UPDATED_LOCATION_ID=$(json_get ".id")
        if [[ -n "$UPDATED_LOCATION_ID" && "$UPDATED_LOCATION_ID" != "null" ]]; then
            register_cleanup "equipment-movement-requirement-locations" "$UPDATED_LOCATION_ID"
            pass
        else
            fail "Created update location but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create update location"
        run_tests
    fi

    test_name "Update stop --location"
    xbe_json do equipment-movement-stops update "$CREATED_STOP_ID" --location "$UPDATED_LOCATION_ID"
    assert_success

    test_name "Update stop --scheduled-arrival-at"
    UPDATED_ARRIVAL="2025-01-01T09:30:00Z"
    xbe_json do equipment-movement-stops update "$CREATED_STOP_ID" --scheduled-arrival-at "$UPDATED_ARRIVAL"
    assert_success

    test_name "Update stop without any fields fails"
    xbe_run do equipment-movement-stops update "$CREATED_STOP_ID"
    assert_failure

    # ------------------------------------------------------------------------
    # SHOW Tests
    # ------------------------------------------------------------------------

    test_name "Show stop"
    xbe_json view equipment-movement-stops show "$CREATED_STOP_ID"
    assert_success

    # ------------------------------------------------------------------------
    # LIST Tests - Filters (requires created stop)
    # ------------------------------------------------------------------------

    test_name "List stops with --trip filter"
    xbe_json view equipment-movement-stops list --trip "$TRIP_ID" --limit 10
    assert_success

    test_name "List stops with --location filter"
    xbe_json view equipment-movement-stops list --location "$UPDATED_LOCATION_ID" --limit 10
    assert_success

    test_name "List stops with --scheduled-arrival-at-min filter"
    xbe_json view equipment-movement-stops list --scheduled-arrival-at-min "2000-01-01T00:00:00Z" --limit 10
    assert_success

    test_name "List stops with --scheduled-arrival-at-max filter"
    xbe_json view equipment-movement-stops list --scheduled-arrival-at-max "2100-01-01T00:00:00Z" --limit 10
    assert_success

    test_name "List stops with --created-at-min filter"
    xbe_json view equipment-movement-stops list --created-at-min "2000-01-01T00:00:00Z" --limit 10
    assert_success

    test_name "List stops with --created-at-max filter"
    xbe_json view equipment-movement-stops list --created-at-max "2100-01-01T00:00:00Z" --limit 10
    assert_success

    test_name "List stops with --updated-at-min filter"
    xbe_json view equipment-movement-stops list --updated-at-min "2000-01-01T00:00:00Z" --limit 10
    assert_success

    test_name "List stops with --updated-at-max filter"
    xbe_json view equipment-movement-stops list --updated-at-max "2100-01-01T00:00:00Z" --limit 10
    assert_success

    # ------------------------------------------------------------------------
    # LIST Tests - Pagination
    # ------------------------------------------------------------------------

    test_name "List stops with --limit"
    xbe_json view equipment-movement-stops list --limit 3
    assert_success

    test_name "List stops with --offset"
    xbe_json view equipment-movement-stops list --limit 3 --offset 3
    assert_success

    # ------------------------------------------------------------------------
    # DELETE Tests
    # ------------------------------------------------------------------------

    test_name "Delete stop requires --confirm flag"
    xbe_run do equipment-movement-stops delete "$CREATED_STOP_ID"
    assert_failure

    test_name "Delete stop with --confirm"
    xbe_json do equipment-movement-stops create \
        --trip "$TRIP_ID" \
        --location "$UPDATED_LOCATION_ID"

    if [[ $status -eq 0 ]]; then
        DEL_STOP_ID=$(json_get ".id")
        if [[ -n "$DEL_STOP_ID" && "$DEL_STOP_ID" != "null" ]]; then
            xbe_run do equipment-movement-stops delete "$DEL_STOP_ID" --confirm
            assert_success
        else
            skip "Could not create stop for deletion test"
        fi
    else
        skip "Could not create stop for deletion test"
    fi
fi

# ============================================================================
# Error Cases (no prerequisites)
# ============================================================================

test_name "Create stop without --trip fails"
xbe_run do equipment-movement-stops create --location 123
assert_failure

test_name "Create stop without --location fails"
xbe_run do equipment-movement-stops create --trip 123
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
