#!/bin/bash
#
# XBE CLI Integration Tests: Digital Fleet Trucks
#
# Tests list and update operations for the digital-fleet-trucks resource.
# Digital fleet trucks are created by integrations and cannot be created or deleted via the API.
#
# COVERAGE: List filters + update relationships (if assignments exist)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: digital-fleet-trucks"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List digital fleet trucks"
xbe_json view digital-fleet-trucks list --limit 5
assert_success

test_name "List digital fleet trucks returns array"
xbe_json view digital-fleet-trucks list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list digital fleet trucks"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List digital fleet trucks with --broker filter"
# Use a likely non-existent broker ID to test filter works without errors
xbe_json view digital-fleet-trucks list --broker 1 --limit 10
assert_success

test_name "List digital fleet trucks with --trucker filter"
# Use a likely non-existent trucker ID to test filter works without errors
xbe_json view digital-fleet-trucks list --trucker 1 --limit 10
assert_success

test_name "List digital fleet trucks with --tractor filter"
# Use a likely non-existent tractor ID to test filter works without errors
xbe_json view digital-fleet-trucks list --tractor 1 --limit 10
assert_success

test_name "List digital fleet trucks with --trailer filter"
# Use a likely non-existent trailer ID to test filter works without errors
xbe_json view digital-fleet-trucks list --trailer 1 --limit 10
assert_success

test_name "List digital fleet trucks with --has-tractor filter"
xbe_json view digital-fleet-trucks list --has-tractor true --limit 10
assert_success

test_name "List digital fleet trucks with --has-trailer filter"
xbe_json view digital-fleet-trucks list --has-trailer true --limit 10
assert_success

test_name "List digital fleet trucks with --assigned-at-min filter"
xbe_json view digital-fleet-trucks list --assigned-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --integration-identifier filter"
# Use a dummy integration identifier to test filter works without errors
xbe_json view digital-fleet-trucks list --integration-identifier "test-integration-id" --limit 10
assert_success

test_name "List digital fleet trucks with --is-active filter"
xbe_json view digital-fleet-trucks list --is-active true --limit 10
assert_success

test_name "List digital fleet trucks with --trailer-set-at-min filter"
xbe_json view digital-fleet-trucks list --trailer-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --trailer-set-at-max filter"
xbe_json view digital-fleet-trucks list --trailer-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --is-trailer-set-at filter"
xbe_json view digital-fleet-trucks list --is-trailer-set-at true --limit 10
assert_success

test_name "List digital fleet trucks with --tractor-set-at-min filter"
xbe_json view digital-fleet-trucks list --tractor-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --tractor-set-at-max filter"
xbe_json view digital-fleet-trucks list --tractor-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --is-tractor-set-at filter"
xbe_json view digital-fleet-trucks list --is-tractor-set-at true --limit 10
assert_success

test_name "List digital fleet trucks with --created-at-min filter"
xbe_json view digital-fleet-trucks list --created-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --created-at-max filter"
xbe_json view digital-fleet-trucks list --created-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --is-created-at filter"
xbe_json view digital-fleet-trucks list --is-created-at true --limit 10
assert_success

test_name "List digital fleet trucks with --updated-at-min filter"
xbe_json view digital-fleet-trucks list --updated-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --updated-at-max filter"
xbe_json view digital-fleet-trucks list --updated-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List digital fleet trucks with --is-updated-at filter"
xbe_json view digital-fleet-trucks list --is-updated-at true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List digital fleet trucks with --limit"
xbe_json view digital-fleet-trucks list --limit 3
assert_success

test_name "List digital fleet trucks with --offset"
xbe_json view digital-fleet-trucks list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Attempt to update digital fleet truck assignment (may skip if unavailable)"
xbe_json view digital-fleet-trucks list --limit 1
if [[ $status -eq 0 ]]; then
    TRUCK_ID=$(json_get ".[0].id")
    if [[ -n "$TRUCK_ID" && "$TRUCK_ID" != "null" ]]; then
        xbe_json view digital-fleet-trucks show "$TRUCK_ID"
        if [[ $status -eq 0 ]]; then
            TRACTOR_ID=$(json_get ".tractor_id")
            TRAILER_ID=$(json_get ".trailer_id")
            if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
                xbe_json do digital-fleet-trucks update "$TRUCK_ID" --tractor "$TRACTOR_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update digital fleet truck tractor - may not have permission"
                fi
            elif [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
                xbe_json do digital-fleet-trucks update "$TRUCK_ID" --trailer "$TRAILER_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update digital fleet truck trailer - may not have permission"
                fi
            else
                skip "No tractor or trailer assignment available to test update"
            fi
        else
            skip "Could not load digital fleet truck details"
        fi
    else
        skip "No digital fleet trucks available to test update"
    fi
else
    skip "Could not list digital fleet trucks to find one for update"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update digital fleet truck without any fields fails"
xbe_json view digital-fleet-trucks list --limit 1
if [[ $status -eq 0 ]]; then
    TRUCK_ID=$(json_get ".[0].id")
    if [[ -n "$TRUCK_ID" && "$TRUCK_ID" != "null" ]]; then
        xbe_run do digital-fleet-trucks update "$TRUCK_ID"
        assert_failure
    else
        skip "No digital fleet trucks available to test error case"
    fi
else
    skip "Could not list digital fleet trucks for error case test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
