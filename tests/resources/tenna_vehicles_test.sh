#!/bin/bash
#
# XBE CLI Integration Tests: Tenna Vehicles
#
# Tests list and update operations for the tenna-vehicles resource.
# Tenna vehicles are created by integrations and cannot be created or deleted via the API.
#
# COVERAGE: List filters + update relationships (if assignments exist)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: tenna-vehicles"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Tenna vehicles"
xbe_json view tenna-vehicles list --limit 5
assert_success

test_name "List Tenna vehicles returns array"
xbe_json view tenna-vehicles list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list Tenna vehicles"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List Tenna vehicles with --broker filter"
# Use a likely non-existent broker ID to test filter works without errors
xbe_json view tenna-vehicles list --broker 1 --limit 10
assert_success

test_name "List Tenna vehicles with --trucker filter"
# Use a likely non-existent trucker ID to test filter works without errors
xbe_json view tenna-vehicles list --trucker 1 --limit 10
assert_success

test_name "List Tenna vehicles with --tractor filter"
# Use a likely non-existent tractor ID to test filter works without errors
xbe_json view tenna-vehicles list --tractor 1 --limit 10
assert_success

test_name "List Tenna vehicles with --trailer filter"
# Use a likely non-existent trailer ID to test filter works without errors
xbe_json view tenna-vehicles list --trailer 1 --limit 10
assert_success

test_name "List Tenna vehicles with --equipment filter"
# Use a likely non-existent equipment ID to test filter works without errors
xbe_json view tenna-vehicles list --equipment 1 --limit 10
assert_success

test_name "List Tenna vehicles with --has-tractor filter"
xbe_json view tenna-vehicles list --has-tractor true --limit 10
assert_success

test_name "List Tenna vehicles with --has-trailer filter"
xbe_json view tenna-vehicles list --has-trailer true --limit 10
assert_success

test_name "List Tenna vehicles with --has-equipment filter"
xbe_json view tenna-vehicles list --has-equipment true --limit 10
assert_success

test_name "List Tenna vehicles with --assigned-at-min filter"
xbe_json view tenna-vehicles list --assigned-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --equipment-assigned-at-min filter"
xbe_json view tenna-vehicles list --equipment-assigned-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --integration-identifier filter"
# Use a dummy integration identifier to test filter works without errors
xbe_json view tenna-vehicles list --integration-identifier "test-integration-id" --limit 10
assert_success

test_name "List Tenna vehicles with --trailer-set-at-min filter"
xbe_json view tenna-vehicles list --trailer-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --trailer-set-at-max filter"
xbe_json view tenna-vehicles list --trailer-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --is-trailer-set-at filter"
xbe_json view tenna-vehicles list --is-trailer-set-at true --limit 10
assert_success

test_name "List Tenna vehicles with --tractor-set-at-min filter"
xbe_json view tenna-vehicles list --tractor-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --tractor-set-at-max filter"
xbe_json view tenna-vehicles list --tractor-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --is-tractor-set-at filter"
xbe_json view tenna-vehicles list --is-tractor-set-at true --limit 10
assert_success

test_name "List Tenna vehicles with --equipment-set-at-min filter"
xbe_json view tenna-vehicles list --equipment-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --equipment-set-at-max filter"
xbe_json view tenna-vehicles list --equipment-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --is-equipment-set-at filter"
xbe_json view tenna-vehicles list --is-equipment-set-at true --limit 10
assert_success

test_name "List Tenna vehicles with --created-at-min filter"
xbe_json view tenna-vehicles list --created-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --created-at-max filter"
xbe_json view tenna-vehicles list --created-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --is-created-at filter"
xbe_json view tenna-vehicles list --is-created-at true --limit 10
assert_success

test_name "List Tenna vehicles with --updated-at-min filter"
xbe_json view tenna-vehicles list --updated-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --updated-at-max filter"
xbe_json view tenna-vehicles list --updated-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Tenna vehicles with --is-updated-at filter"
xbe_json view tenna-vehicles list --is-updated-at true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List Tenna vehicles with --limit"
xbe_json view tenna-vehicles list --limit 3
assert_success

test_name "List Tenna vehicles with --offset"
xbe_json view tenna-vehicles list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Attempt to update Tenna vehicle assignment (may skip if unavailable)"
xbe_json view tenna-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    VEHICLE_ID=$(json_get ".[0].id")
    if [[ -n "$VEHICLE_ID" && "$VEHICLE_ID" != "null" ]]; then
        xbe_json view tenna-vehicles show "$VEHICLE_ID"
        if [[ $status -eq 0 ]]; then
            EQUIPMENT_ID=$(json_get ".equipment_id")
            TRACTOR_ID=$(json_get ".tractor_id")
            TRAILER_ID=$(json_get ".trailer_id")
            if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
                xbe_json do tenna-vehicles update "$VEHICLE_ID" --equipment "$EQUIPMENT_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update Tenna vehicle equipment - may not have permission"
                fi
            elif [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
                xbe_json do tenna-vehicles update "$VEHICLE_ID" --tractor "$TRACTOR_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update Tenna vehicle tractor - may not have permission"
                fi
            elif [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
                xbe_json do tenna-vehicles update "$VEHICLE_ID" --trailer "$TRAILER_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update Tenna vehicle trailer - may not have permission"
                fi
            else
                skip "No tractor, trailer, or equipment assignment available to test update"
            fi
        else
            skip "Could not load Tenna vehicle details"
        fi
    else
        skip "No Tenna vehicles available to test update"
    fi
else
    skip "Could not list Tenna vehicles to find one for update"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update Tenna vehicle without any fields fails"
xbe_json view tenna-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    VEHICLE_ID=$(json_get ".[0].id")
    if [[ -n "$VEHICLE_ID" && "$VEHICLE_ID" != "null" ]]; then
        xbe_run do tenna-vehicles update "$VEHICLE_ID"
        assert_failure
    else
        skip "No Tenna vehicles available to test error case"
    fi
else
    skip "Could not list Tenna vehicles for error case test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
