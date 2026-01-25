#!/bin/bash
#
# XBE CLI Integration Tests: KeepTruckin Vehicles
#
# Tests list and update operations for the keep-truckin-vehicles resource.
# KeepTruckin vehicles are created by integrations and cannot be created or deleted via the API.
#
# COVERAGE: List filters + update relationships (if assignments exist)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: keep-truckin-vehicles"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List KeepTruckin vehicles"
xbe_json view keep-truckin-vehicles list --limit 5
assert_success

test_name "List KeepTruckin vehicles returns array"
xbe_json view keep-truckin-vehicles list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list KeepTruckin vehicles"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List KeepTruckin vehicles with --broker filter"
# Use a likely non-existent broker ID to test filter works without errors
xbe_json view keep-truckin-vehicles list --broker 1 --limit 10
assert_success

test_name "List KeepTruckin vehicles with --trucker filter"
# Use a likely non-existent trucker ID to test filter works without errors
xbe_json view keep-truckin-vehicles list --trucker 1 --limit 10
assert_success

test_name "List KeepTruckin vehicles with --trucker-id filter"
# Use a likely non-existent trucker ID to test filter works without errors
xbe_json view keep-truckin-vehicles list --trucker-id 1 --limit 10
assert_success

test_name "List KeepTruckin vehicles with --tractor filter"
# Use a likely non-existent tractor ID to test filter works without errors
xbe_json view keep-truckin-vehicles list --tractor 1 --limit 10
assert_success

test_name "List KeepTruckin vehicles with --trailer filter"
# Use a likely non-existent trailer ID to test filter works without errors
xbe_json view keep-truckin-vehicles list --trailer 1 --limit 10
assert_success

test_name "List KeepTruckin vehicles with --has-tractor filter"
xbe_json view keep-truckin-vehicles list --has-tractor true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --has-trailer filter"
xbe_json view keep-truckin-vehicles list --has-trailer true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --assigned-at-min filter"
xbe_json view keep-truckin-vehicles list --assigned-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --integration-identifier filter"
# Use a dummy integration identifier to test filter works without errors
xbe_json view keep-truckin-vehicles list --integration-identifier "test-integration-id" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --active filter"
xbe_json view keep-truckin-vehicles list --active true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --trailer-set-at-min filter"
xbe_json view keep-truckin-vehicles list --trailer-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --trailer-set-at-max filter"
xbe_json view keep-truckin-vehicles list --trailer-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --is-trailer-set-at filter"
xbe_json view keep-truckin-vehicles list --is-trailer-set-at true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --tractor-set-at-min filter"
xbe_json view keep-truckin-vehicles list --tractor-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --tractor-set-at-max filter"
xbe_json view keep-truckin-vehicles list --tractor-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --is-tractor-set-at filter"
xbe_json view keep-truckin-vehicles list --is-tractor-set-at true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --created-at-min filter"
xbe_json view keep-truckin-vehicles list --created-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --created-at-max filter"
xbe_json view keep-truckin-vehicles list --created-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --is-created-at filter"
xbe_json view keep-truckin-vehicles list --is-created-at true --limit 10
assert_success

test_name "List KeepTruckin vehicles with --updated-at-min filter"
xbe_json view keep-truckin-vehicles list --updated-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --updated-at-max filter"
xbe_json view keep-truckin-vehicles list --updated-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List KeepTruckin vehicles with --is-updated-at filter"
xbe_json view keep-truckin-vehicles list --is-updated-at true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List KeepTruckin vehicles with --limit"
xbe_json view keep-truckin-vehicles list --limit 3
assert_success

test_name "List KeepTruckin vehicles with --offset"
xbe_json view keep-truckin-vehicles list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Attempt to update KeepTruckin vehicle assignment (may skip if unavailable)"
xbe_json view keep-truckin-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    VEHICLE_ID=$(json_get ".[0].id")
    if [[ -n "$VEHICLE_ID" && "$VEHICLE_ID" != "null" ]]; then
        xbe_json view keep-truckin-vehicles show "$VEHICLE_ID"
        if [[ $status -eq 0 ]]; then
            TRACTOR_ID=$(json_get ".tractor_id")
            TRAILER_ID=$(json_get ".trailer_id")
            if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
                xbe_json do keep-truckin-vehicles update "$VEHICLE_ID" --tractor "$TRACTOR_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update KeepTruckin vehicle tractor - may not have permission"
                fi
            elif [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
                xbe_json do keep-truckin-vehicles update "$VEHICLE_ID" --trailer "$TRAILER_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update KeepTruckin vehicle trailer - may not have permission"
                fi
            else
                skip "No tractor or trailer assignment available to test update"
            fi
        else
            skip "Could not load KeepTruckin vehicle details"
        fi
    else
        skip "No KeepTruckin vehicles available to test update"
    fi
else
    skip "Could not list KeepTruckin vehicles to find one for update"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update KeepTruckin vehicle without any fields fails"
xbe_json view keep-truckin-vehicles list --limit 1
if [[ $status -eq 0 ]]; then
    VEHICLE_ID=$(json_get ".[0].id")
    if [[ -n "$VEHICLE_ID" && "$VEHICLE_ID" != "null" ]]; then
        xbe_run do keep-truckin-vehicles update "$VEHICLE_ID"
        assert_failure
    else
        skip "No KeepTruckin vehicles available to test error case"
    fi
else
    skip "Could not list KeepTruckin vehicles for error case test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
