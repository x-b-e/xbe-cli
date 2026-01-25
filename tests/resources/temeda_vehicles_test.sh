#!/bin/bash
#
# XBE CLI Integration Tests: Temeda Vehicles
#
# Tests operations for the temeda-vehicles resource.
# Note: Temeda vehicles are read-only; only update assignments are available.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TEMEDA_VEHICLE_ID=""
BROKER_ID=""
TRUCKER_ID=""
TRAILER_ID=""
TRACTOR_ID=""
INTEGRATION_IDENTIFIER=""

DATE_MIN="2000-01-01T00:00:00Z"
DATE_MAX="2100-01-01T00:00:00Z"

describe "Resource: temeda-vehicles"

# ============================================================================
# LIST Tests - Get an ID for further tests
# ============================================================================

test_name "List temeda vehicles"
xbe_json view temeda-vehicles list --limit 5
assert_success

test_name "List temeda vehicles returns array"
xbe_json view temeda-vehicles list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list temeda vehicles"
fi

test_name "Get a temeda vehicle ID for tests"
xbe_json view temeda-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    TEMEDA_VEHICLE_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    TRUCKER_ID=$(json_get ".[0].trucker_id")
    TRAILER_ID=$(json_get ".[0].trailer_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    INTEGRATION_IDENTIFIER=$(json_get ".[0].integration_identifier")
    if [[ -n "$TEMEDA_VEHICLE_ID" && "$TEMEDA_VEHICLE_ID" != "null" ]]; then
        pass
    else
        skip "No temeda vehicles found in the system"
        run_tests
    fi
else
    fail "Failed to list temeda vehicles"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show temeda vehicle"
xbe_json view temeda-vehicles show "$TEMEDA_VEHICLE_ID"
assert_success

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter temeda vehicles by broker"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view temeda-vehicles list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "Filter temeda vehicles by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view temeda-vehicles list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "Filter temeda vehicles by trailer"
if [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
    xbe_json view temeda-vehicles list --trailer "$TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "Filter temeda vehicles by tractor"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view temeda-vehicles list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "Filter temeda vehicles with trailer"
xbe_json view temeda-vehicles list --has-trailer true --limit 5
assert_success

test_name "Filter temeda vehicles with tractor"
xbe_json view temeda-vehicles list --has-tractor true --limit 5
assert_success

test_name "Filter temeda vehicles by assigned-at-min"
xbe_json view temeda-vehicles list --assigned-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter temeda vehicles by integration identifier"
if [[ -n "$INTEGRATION_IDENTIFIER" && "$INTEGRATION_IDENTIFIER" != "null" ]]; then
    xbe_json view temeda-vehicles list --integration-identifier "$INTEGRATION_IDENTIFIER" --limit 5
    assert_success
else
    skip "No integration identifier available"
fi

test_name "Filter temeda vehicles by trailer-set-at-min"
xbe_json view temeda-vehicles list --trailer-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter temeda vehicles by trailer-set-at-max"
xbe_json view temeda-vehicles list --trailer-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter temeda vehicles by is-trailer-set-at"
xbe_json view temeda-vehicles list --is-trailer-set-at true --limit 5
assert_success

test_name "Filter temeda vehicles by tractor-set-at-min"
xbe_json view temeda-vehicles list --tractor-set-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter temeda vehicles by tractor-set-at-max"
xbe_json view temeda-vehicles list --tractor-set-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter temeda vehicles by is-tractor-set-at"
xbe_json view temeda-vehicles list --is-tractor-set-at true --limit 5
assert_success

test_name "Filter temeda vehicles by created-at-min"
xbe_json view temeda-vehicles list --created-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter temeda vehicles by created-at-max"
xbe_json view temeda-vehicles list --created-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter temeda vehicles by is-created-at"
xbe_json view temeda-vehicles list --is-created-at true --limit 5
assert_success

test_name "Filter temeda vehicles by updated-at-min"
xbe_json view temeda-vehicles list --updated-at-min "$DATE_MIN" --limit 5
assert_success

test_name "Filter temeda vehicles by updated-at-max"
xbe_json view temeda-vehicles list --updated-at-max "$DATE_MAX" --limit 5
assert_success

test_name "Filter temeda vehicles by is-updated-at"
xbe_json view temeda-vehicles list --is-updated-at true --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List temeda vehicles with --limit"
xbe_json view temeda-vehicles list --limit 3
assert_success

test_name "List temeda vehicles with --offset"
xbe_json view temeda-vehicles list --limit 3 --offset 1
assert_success

# ============================================================================
# UPDATE Tests - Assignment Relationships
# ============================================================================

TRAILER_UPDATE_VEHICLE_ID="$TEMEDA_VEHICLE_ID"
TRAILER_UPDATE_ID="$TRAILER_ID"

if [[ -z "$TRAILER_UPDATE_ID" || "$TRAILER_UPDATE_ID" == "null" ]]; then
    xbe_json view temeda-vehicles list --has-trailer true --limit 1
    if [[ $status -eq 0 ]]; then
        TRAILER_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRAILER_UPDATE_ID=$(json_get ".[0].trailer_id")
    fi
fi

test_name "Update temeda vehicle trailer assignment"
if [[ -n "$TRAILER_UPDATE_ID" && "$TRAILER_UPDATE_ID" != "null" && -n "$TRAILER_UPDATE_VEHICLE_ID" && "$TRAILER_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do temeda-vehicles update "$TRAILER_UPDATE_VEHICLE_ID" --trailer "$TRAILER_UPDATE_ID"
    assert_success
else
    skip "No temeda vehicle with trailer available"
fi

TRACTOR_UPDATE_VEHICLE_ID="$TEMEDA_VEHICLE_ID"
TRACTOR_UPDATE_ID="$TRACTOR_ID"

if [[ -z "$TRACTOR_UPDATE_ID" || "$TRACTOR_UPDATE_ID" == "null" ]]; then
    xbe_json view temeda-vehicles list --has-tractor true --limit 1
    if [[ $status -eq 0 ]]; then
        TRACTOR_UPDATE_VEHICLE_ID=$(json_get ".[0].id")
        TRACTOR_UPDATE_ID=$(json_get ".[0].tractor_id")
    fi
fi

test_name "Update temeda vehicle tractor assignment"
if [[ -n "$TRACTOR_UPDATE_ID" && "$TRACTOR_UPDATE_ID" != "null" && -n "$TRACTOR_UPDATE_VEHICLE_ID" && "$TRACTOR_UPDATE_VEHICLE_ID" != "null" ]]; then
    xbe_json do temeda-vehicles update "$TRACTOR_UPDATE_VEHICLE_ID" --tractor "$TRACTOR_UPDATE_ID"
    assert_success
else
    skip "No temeda vehicle with tractor available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Update without any fields fails"
xbe_json do temeda-vehicles update "$TEMEDA_VEHICLE_ID"
assert_failure

test_name "Update non-existent temeda vehicle fails"
xbe_json do temeda-vehicles update "99999999" --trailer "1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
