#!/bin/bash
#
# XBE CLI Integration Tests: Business Unit Equipments
#
# Tests list, show, create, and delete operations for business_unit_equipments.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_BUSINESS_UNIT_EQUIPMENT_ID=""
BUSINESS_UNIT_ID="${XBE_TEST_BUSINESS_UNIT_ID:-}"
EQUIPMENT_ID="${XBE_TEST_EQUIPMENT_ID:-}"
CREATE_BUSINESS_UNIT_ID="${XBE_TEST_BUSINESS_UNIT_ID:-}"
CREATE_EQUIPMENT_ID="${XBE_TEST_EQUIPMENT_ID:-}"
CREATED_BUSINESS_UNIT_EQUIPMENT_ID=""

describe "Resource: business-unit-equipments"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List business unit equipments"
xbe_json view business-unit-equipments list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_BUSINESS_UNIT_EQUIPMENT_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$BUSINESS_UNIT_ID" || "$BUSINESS_UNIT_ID" == "null" ]]; then
            BUSINESS_UNIT_ID=$(echo "$output" | jq -r '.[0].business_unit_id')
        fi
        if [[ -z "$EQUIPMENT_ID" || "$EQUIPMENT_ID" == "null" ]]; then
            EQUIPMENT_ID=$(echo "$output" | jq -r '.[0].equipment_id')
        fi
    fi
else
    fail "Failed to list business unit equipments"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show business unit equipment"
if [[ -n "$SEED_BUSINESS_UNIT_EQUIPMENT_ID" && "$SEED_BUSINESS_UNIT_EQUIPMENT_ID" != "null" ]]; then
    xbe_json view business-unit-equipments show "$SEED_BUSINESS_UNIT_EQUIPMENT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        if [[ -z "$BUSINESS_UNIT_ID" || "$BUSINESS_UNIT_ID" == "null" ]]; then
            BUSINESS_UNIT_ID=$(json_get ".business_unit_id")
        fi
        if [[ -z "$EQUIPMENT_ID" || "$EQUIPMENT_ID" == "null" ]]; then
            EQUIPMENT_ID=$(json_get ".equipment_id")
        fi
        pass
    else
        fail "Failed to show business unit equipment"
    fi
else
    skip "No business unit equipment available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create business unit equipment"
if [[ -n "$CREATE_BUSINESS_UNIT_ID" && "$CREATE_BUSINESS_UNIT_ID" != "null" && -n "$CREATE_EQUIPMENT_ID" && "$CREATE_EQUIPMENT_ID" != "null" ]]; then
    xbe_json do business-unit-equipments create --business-unit "$CREATE_BUSINESS_UNIT_ID" --equipment "$CREATE_EQUIPMENT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BUSINESS_UNIT_EQUIPMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID" && "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID" != "null" ]]; then
            register_cleanup "business-unit-equipments" "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID"
            pass
        else
            fail "Created business unit equipment but no ID returned"
        fi
    else
        fail "Failed to create business unit equipment"
    fi
else
    skip "No business unit or equipment ID available for creation (set XBE_TEST_BUSINESS_UNIT_ID and XBE_TEST_EQUIPMENT_ID to a new pair)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete business unit equipment"
if [[ -n "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID" && "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID" != "null" ]]; then
    xbe_run do business-unit-equipments delete "$CREATED_BUSINESS_UNIT_EQUIPMENT_ID" --confirm
    assert_success
else
    skip "No created business unit equipment to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by business unit"
if [[ -n "$BUSINESS_UNIT_ID" && "$BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view business-unit-equipments list --business-unit "$BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available for filter"
fi

test_name "Filter by equipment"
if [[ -n "$EQUIPMENT_ID" && "$EQUIPMENT_ID" != "null" ]]; then
    xbe_json view business-unit-equipments list --equipment "$EQUIPMENT_ID" --limit 5
    assert_success
else
    skip "No equipment ID available for filter"
fi

run_tests
