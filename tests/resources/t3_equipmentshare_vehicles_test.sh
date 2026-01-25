#!/bin/bash
#
# XBE CLI Integration Tests: T3 EquipmentShare Vehicles
#
# Tests operations for the t3-equipmentshare-vehicles resource.
# Note: T3 EquipmentShare vehicles are read-only; only update assignments are available.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

T3_EQUIPMENTSHARE_VEHICLE_ID=""
BROKER_ID=""
TRUCKER_ID=""
TRAILER_ID=""
TRACTOR_ID=""
INTEGRATION_IDENTIFIER=""

DATE_MIN="2000-01-01T00:00:00Z"
DATE_MAX="2100-01-01T00:00:00Z"

describe "Resource: t3-equipmentshare-vehicles"

# ============================================================================
# LIST Tests - Get an ID for further tests
# ============================================================================

test_name "List T3 EquipmentShare vehicles"
xbe_json view t3-equipmentshare-vehicles list --limit 5
assert_success

test_name "List T3 EquipmentShare vehicles returns array"
xbe_json view t3-equipmentshare-vehicles list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list T3 EquipmentShare vehicles"
fi

test_name "Get a T3 EquipmentShare vehicle ID for tests"
xbe_json view t3-equipmentshare-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    T3_EQUIPMENTSHARE_VEHICLE_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    TRUCKER_ID=$(json_get ".[0].trucker_id")
    TRAILER_ID=$(json_get ".[0].trailer_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    INTEGRATION_IDENTIFIER=$(json_get ".[0].integration_identifier")
    if [[ -n "$T3_EQUIPMENTSHARE_VEHICLE_ID" && "$T3_EQUIPMENTSHARE_VEHICLE_ID" != "null" ]]; then
        pass
    else
        skip "No T3 EquipmentShare vehicles found in the system"
        run_tests
    fi
else
    fail "Failed to list T3 EquipmentShare vehicles"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show T3 EquipmentShare vehicle"
xbe_json view t3-equipmentshare-vehicles show "$T3_EQUIPMENTSHARE_VEHICLE_ID"
assert_success

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter T3 EquipmentShare vehicles by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter T3 EquipmentShare vehicles by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "Filter T3 EquipmentShare vehicles by trailer"
if [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --trailer "$TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "Filter T3 EquipmentShare vehicles by tractor"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "Filter T3 EquipmentShare vehicles with trailer"
xbe_json view t3-equipmentshare-vehicles list --has-trailer true --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles with tractor"
xbe_json view t3-equipmentshare-vehicles list --has-tractor true --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by assigned-at-min"
xbe_json view t3-equipmentshare-vehicles list --assigned-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by integration identifier"
if [[ -n "$INTEGRATION_IDENTIFIER" && "$INTEGRATION_IDENTIFIER" != "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --integration-identifier "$INTEGRATION_IDENTIFIER" --limit 5
    assert_success
else
    skip "No integration identifier available"
fi

test_name "Filter T3 EquipmentShare vehicles by trailer-set-at-min"
xbe_json view t3-equipmentshare-vehicles list --trailer-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by trailer-set-at-max"
xbe_json view t3-equipmentshare-vehicles list --trailer-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by is-trailer-set-at"
xbe_json view t3-equipmentshare-vehicles list --is-trailer-set-at true --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by tractor-set-at-min"
xbe_json view t3-equipmentshare-vehicles list --tractor-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by tractor-set-at-max"
xbe_json view t3-equipmentshare-vehicles list --tractor-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by is-tractor-set-at"
xbe_json view t3-equipmentshare-vehicles list --is-tractor-set-at true --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by created-at-min"
xbe_json view t3-equipmentshare-vehicles list --created-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by created-at-max"
xbe_json view t3-equipmentshare-vehicles list --created-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by is-created-at"
xbe_json view t3-equipmentshare-vehicles list --is-created-at true --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by updated-at-min"
xbe_json view t3-equipmentshare-vehicles list --updated-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by updated-at-max"
xbe_json view t3-equipmentshare-vehicles list --updated-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter T3 EquipmentShare vehicles by is-updated-at"
xbe_json view t3-equipmentshare-vehicles list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List T3 EquipmentShare vehicles with --limit"
xbe_json view t3-equipmentshare-vehicles list --limit 3
assert_success

test_name "List T3 EquipmentShare vehicles with --offset"
xbe_json view t3-equipmentshare-vehicles list --limit 3 --offset 1
assert_success

# ============================================================================
# UPDATE Tests - Assignment Relationships
# ============================================================================

TRAILER_UPDATE_VEHICLE_ID="$T3_EQUIPMENTSHARE_VEHICLE_ID"
TRAILER_UPDATE_ID="$TRAILER_ID"

if [[ -z "$TRAILER_UPDATE_ID" || "$TRAILER_UPDATE_ID" == "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --has-trailer true --limit 1
    if [[ $status -eq 0 ]]; then
        TRAILER_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRAILER_UPDATE_ID=$(json_get ".[0].trailer_id")
    fi
fi

test_name "Update T3 EquipmentShare vehicle trailer assignment"
if [[ -n "$TRAILER_UPDATE_ID" && "$TRAILER_UPDATE_ID" != "null" && -n "$TRAILER_UPDATE_VEHICLE_ID" && "$TRAILER_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do t3-equipmentshare-vehicles update "$TRAILER_UPDATE_VEHICLE_ID" --trailer "$TRAILER_UPDATE_ID"
    assert_success
else
    skip "No T3 EquipmentShare vehicle with trailer available"
fi

TRACTOR_UPDATE_VEHICLE_ID="$T3_EQUIPMENTSHARE_VEHICLE_ID"
TRACTOR_UPDATE_ID="$TRACTOR_ID"

if [[ -z "$TRACTOR_UPDATE_ID" || "$TRACTOR_UPDATE_ID" == "null" ]]; then
    xbe_json view t3-equipmentshare-vehicles list --has-tractor true --limit 1
    if [[ $status -eq 0 ]]; then
        TRACTOR_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRACTOR_UPDATE_ID=$(json_get ".[0].tractor_id")
    fi
fi

test_name "Update T3 EquipmentShare vehicle tractor assignment"
if [[ -n "$TRACTOR_UPDATE_ID" && "$TRACTOR_UPDATE_ID" != "null" && -n "$TRACTOR_UPDATE_VEHICLE_ID" && "$TRACTOR_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do t3-equipmentshare-vehicles update "$TRACTOR_UPDATE_VEHICLE_ID" --tractor "$TRACTOR_UPDATE_ID"
    assert_success
else
    skip "No T3 EquipmentShare vehicle with tractor available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Update without any fields fails"
xbe_json do t3-equipmentshare-vehicles update "$T3_EQUIPMENTSHARE_VEHICLE_ID"
assert_failure

test_name "Update non-existent T3 EquipmentShare vehicle fails"
xbe_json do t3-equipmentshare-vehicles update "99999999" --trailer "1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
