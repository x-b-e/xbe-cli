#!/bin/bash
#
# XBE CLI Integration Tests: Deere Equipments
#
# Tests list and update operations for the deere-equipments resource.
# Deere equipment is created by integrations and cannot be created or deleted via the API.
#
# COVERAGE: List filters + update attributes/relationships (if equipment exists)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: deere-equipments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Deere equipment"
xbe_json view deere-equipments list --limit 5
assert_success

test_name "List Deere equipment returns array"
xbe_json view deere-equipments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list Deere equipment"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List Deere equipment with --broker filter"
# Use a likely non-existent broker ID to test filter works without errors
xbe_json view deere-equipments list --broker 1 --limit 10
assert_success

test_name "List Deere equipment with --equipment filter"
# Use a likely non-existent equipment ID to test filter works without errors
xbe_json view deere-equipments list --equipment 1 --limit 10
assert_success

test_name "List Deere equipment with --has-equipment filter"
xbe_json view deere-equipments list --has-equipment true --limit 10
assert_success

test_name "List Deere equipment with --equipment-assigned-at-min filter"
xbe_json view deere-equipments list --equipment-assigned-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --integration-identifier filter"
# Use a dummy integration identifier to test filter works without errors
xbe_json view deere-equipments list --integration-identifier "test-integration-id" --limit 10
assert_success

test_name "List Deere equipment with --equipment-set-at-min filter"
xbe_json view deere-equipments list --equipment-set-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --equipment-set-at-max filter"
xbe_json view deere-equipments list --equipment-set-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --is-equipment-set-at filter"
xbe_json view deere-equipments list --is-equipment-set-at true --limit 10
assert_success

test_name "List Deere equipment with --created-at-min filter"
xbe_json view deere-equipments list --created-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --created-at-max filter"
xbe_json view deere-equipments list --created-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --is-created-at filter"
xbe_json view deere-equipments list --is-created-at true --limit 10
assert_success

test_name "List Deere equipment with --updated-at-min filter"
xbe_json view deere-equipments list --updated-at-min "2020-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --updated-at-max filter"
xbe_json view deere-equipments list --updated-at-max "2030-01-01T00:00:00Z" --limit 10
assert_success

test_name "List Deere equipment with --is-updated-at filter"
xbe_json view deere-equipments list --is-updated-at true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List Deere equipment with --limit"
xbe_json view deere-equipments list --limit 3
assert_success

test_name "List Deere equipment with --offset"
xbe_json view deere-equipments list --limit 3 --offset 3
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Attempt to update Deere equipment assignment (may skip if unavailable)"
xbe_json view deere-equipments list --limit 1
if [[ $status -eq 0 ]]; then
    EQUIPMENT_ID=$(json_get ".[0].id")
    if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
        xbe_json view deere-equipments show "$EQUIPMENT_ID"
        if [[ $status -eq 0 ]]; then
            ASSIGNED_ID=$(json_get ".equipment_id")
            if [[ -n "$ASSIGNED_ID" && "$ASSIGNED_ID" != "null" ]]; then
                xbe_json do deere-equipments update "$EQUIPMENT_ID" --equipment "$ASSIGNED_ID"
                if [[ $status -eq 0 ]]; then
                    pass
                else
                    skip "Could not update Deere equipment assignment - may not have permission"
                fi
            else
                skip "No assigned equipment available to test relationship update"
            fi
        else
            skip "Could not load Deere equipment details"
        fi
    else
        skip "No Deere equipment available to test assignment update"
    fi
else
    skip "Could not list Deere equipment to find one for assignment update"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update Deere equipment without any fields fails"
xbe_json view deere-equipments list --limit 1
if [[ $status -eq 0 ]]; then
    EQUIPMENT_ID=$(json_get ".[0].id")
    if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
        xbe_run do deere-equipments update "$EQUIPMENT_ID"
        assert_failure
    else
        skip "No Deere equipment available to test error case"
    fi
else
    skip "Could not list Deere equipment for error case test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
