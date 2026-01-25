#!/bin/bash
#
# XBE CLI Integration Tests: Verizon Reveal Vehicles
#
# Tests operations for the verizon-reveal-vehicles resource.
# Note: Verizon Reveal vehicles are read-only; only update assignments are available.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

VERIZON_REVEAL_VEHICLE_ID=""
BROKER_ID=""
TRUCKER_ID=""
TRAILER_ID=""
TRACTOR_ID=""
EQUIPMENT_ID=""
INTEGRATION_IDENTIFIER=""

DATE_MIN="2000-01-01T00:00:00Z"
DATE_MAX="2100-01-01T00:00:00Z"

describe "Resource: verizon-reveal-vehicles"

# ============================================================================
# LIST Tests - Get an ID for further tests
# ============================================================================

test_name "List Verizon Reveal vehicles"
xbe_json view verizon-reveal-vehicles list --limit 5
assert_success

test_name "List Verizon Reveal vehicles returns array"
xbe_json view verizon-reveal-vehicles list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list Verizon Reveal vehicles"
fi

test_name "Get a Verizon Reveal vehicle ID for tests"
xbe_json view verizon-reveal-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    VERIZON_REVEAL_VEHICLE_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    TRUCKER_ID=$(json_get ".[0].trucker_id")
    TRAILER_ID=$(json_get ".[0].trailer_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    EQUIPMENT_ID=$(json_get ".[0].equipment_id")
    INTEGRATION_IDENTIFIER=$(json_get ".[0].integration_identifier")
    if [[ -n "$VERIZON_REVEAL_VEHICLE_ID" && "$VERIZON_REVEAL_VEHICLE_ID" != "null" ]]; then
        pass
    else
        skip "No Verizon Reveal vehicles found in the system"
        run_tests
    fi
else
    fail "Failed to list Verizon Reveal vehicles"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show Verizon Reveal vehicle"
xbe_json view verizon-reveal-vehicles show "$VERIZON_REVEAL_VEHICLE_ID"
assert_success

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter Verizon Reveal vehicles by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter Verizon Reveal vehicles by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "Filter Verizon Reveal vehicles by trailer"
if [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --trailer "$TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "Filter Verizon Reveal vehicles by tractor"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "Filter Verizon Reveal vehicles by equipment"
if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --equipment "$EQUIPMENT_ID" --limit 5
    assert_success
else
    skip "No equipment ID available"
fi

test_name "Filter Verizon Reveal vehicles with trailer"
xbe_json view verizon-reveal-vehicles list --has-trailer true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles with tractor"
xbe_json view verizon-reveal-vehicles list --has-tractor true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles with equipment"
xbe_json view verizon-reveal-vehicles list --has-equipment true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by assigned-at-min"
xbe_json view verizon-reveal-vehicles list --assigned-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by equipment-assigned-at-min"
xbe_json view verizon-reveal-vehicles list --equipment-assigned-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by integration identifier"
if [[ -n "$INTEGRATION_IDENTIFIER" && "$INTEGRATION_IDENTIFIER" != "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --integration-identifier "$INTEGRATION_IDENTIFIER" --limit 5
    assert_success
else
    skip "No integration identifier available"
fi

test_name "Filter Verizon Reveal vehicles by trailer-set-at-min"
xbe_json view verizon-reveal-vehicles list --trailer-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by trailer-set-at-max"
xbe_json view verizon-reveal-vehicles list --trailer-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by is-trailer-set-at"
xbe_json view verizon-reveal-vehicles list --is-trailer-set-at true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by tractor-set-at-min"
xbe_json view verizon-reveal-vehicles list --tractor-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by tractor-set-at-max"
xbe_json view verizon-reveal-vehicles list --tractor-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by is-tractor-set-at"
xbe_json view verizon-reveal-vehicles list --is-tractor-set-at true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by equipment-set-at-min"
xbe_json view verizon-reveal-vehicles list --equipment-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by equipment-set-at-max"
xbe_json view verizon-reveal-vehicles list --equipment-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by is-equipment-set-at"
xbe_json view verizon-reveal-vehicles list --is-equipment-set-at true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by created-at-min"
xbe_json view verizon-reveal-vehicles list --created-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by created-at-max"
xbe_json view verizon-reveal-vehicles list --created-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by is-created-at"
xbe_json view verizon-reveal-vehicles list --is-created-at true --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by updated-at-min"
xbe_json view verizon-reveal-vehicles list --updated-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by updated-at-max"
xbe_json view verizon-reveal-vehicles list --updated-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter Verizon Reveal vehicles by is-updated-at"
xbe_json view verizon-reveal-vehicles list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List Verizon Reveal vehicles with --limit"
xbe_json view verizon-reveal-vehicles list --limit 3
assert_success

test_name "List Verizon Reveal vehicles with --offset"
xbe_json view verizon-reveal-vehicles list --limit 3 --offset 1
assert_success

# ============================================================================
# UPDATE Tests - Assignment Relationships
# ============================================================================

TRAILER_UPDATE_VEHICLE_ID="$VERIZON_REVEAL_VEHICLE_ID"
TRAILER_UPDATE_ID="$TRAILER_ID"

if [[ -z "$TRAILER_UPDATE_ID" || "$TRAILER_UPDATE_ID" == "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --has-trailer true --limit 1
    if [[ $status -eq 0 ]]; then
        TRAILER_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRAILER_UPDATE_ID=$(json_get ".[0].trailer_id")
    fi
fi

test_name "Update Verizon Reveal vehicle trailer assignment"
if [[ -n "$TRAILER_UPDATE_ID" && "$TRAILER_UPDATE_ID" != "null" && -n "$TRAILER_UPDATE_VEHICLE_ID" && "$TRAILER_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do verizon-reveal-vehicles update "$TRAILER_UPDATE_VEHICLE_ID" \
        --trailer "$TRAILER_UPDATE_ID"
    assert_success
else
    skip "No Verizon Reveal vehicle with trailer available"
fi

TRACTOR_UPDATE_VEHICLE_ID="$VERIZON_REVEAL_VEHICLE_ID"
TRACTOR_UPDATE_ID="$TRACTOR_ID"

if [[ -z "$TRACTOR_UPDATE_ID" || "$TRACTOR_UPDATE_ID" == "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --has-tractor true --limit 1
    if [[ $status -eq 0 ]]; then
        TRACTOR_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRACTOR_UPDATE_ID=$(json_get ".[0].tractor_id")
    fi
fi

test_name "Update Verizon Reveal vehicle tractor assignment"
if [[ -n "$TRACTOR_UPDATE_ID" && "$TRACTOR_UPDATE_ID" != "null" && -n "$TRACTOR_UPDATE_VEHICLE_ID" && "$TRACTOR_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do verizon-reveal-vehicles update "$TRACTOR_UPDATE_VEHICLE_ID" \
        --tractor "$TRACTOR_UPDATE_ID"
    assert_success
else
    skip "No Verizon Reveal vehicle with tractor available"
fi

EQUIPMENT_UPDATE_VEHICLE_ID="$VERIZON_REVEAL_VEHICLE_ID"
EQUIPMENT_UPDATE_ID="$EQUIPMENT_ID"

if [[ -z "$EQUIPMENT_UPDATE_ID" || "$EQUIPMENT_UPDATE_ID" == "null" ]]; then
    xbe_json view verizon-reveal-vehicles list --has-equipment true --limit 1
    if [[ $status -eq 0 ]]; then
        EQUIPMENT_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        EQUIPMENT_UPDATE_ID=$(json_get ".[0].equipment_id")
    fi
fi

test_name "Update Verizon Reveal vehicle equipment assignment"
if [[ -n "$EQUIPMENT_UPDATE_ID" && "$EQUIPMENT_UPDATE_ID" != "null" && -n "$EQUIPMENT_UPDATE_VEHICLE_ID" && "$EQUIPMENT_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do verizon-reveal-vehicles update "$EQUIPMENT_UPDATE_VEHICLE_ID" \
        --equipment "$EQUIPMENT_UPDATE_ID"
    assert_success
else
    skip "No Verizon Reveal vehicle with equipment available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Update without any fields fails"
xbe_json do verizon-reveal-vehicles update "$VERIZON_REVEAL_VEHICLE_ID"
assert_failure

test_name "Update non-existent Verizon Reveal vehicle fails"
xbe_json do verizon-reveal-vehicles update "99999999" --trailer "1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
