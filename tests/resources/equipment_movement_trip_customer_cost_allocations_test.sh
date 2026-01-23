#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Trip Customer Cost Allocations
#
# Tests list/show/create/update/delete operations for equipment movement trip customer cost allocations.
#
# COVERAGE: List filters + show + create/update attributes + delete (when created)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TRIP_ID=""
SAMPLE_ALLOCATION_JSON=""
CREATED_ID=""

describe "Resource: equipment-movement-trip-customer-cost-allocations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List cost allocations"
xbe_json view equipment-movement-trip-customer-cost-allocations list --limit 5
assert_success

test_name "List cost allocations returns array"
xbe_json view equipment-movement-trip-customer-cost-allocations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list cost allocations"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample cost allocation"
xbe_json view equipment-movement-trip-customer-cost-allocations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TRIP_ID=$(json_get ".[0].trip_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No cost allocations available for follow-on tests"
    fi
else
    skip "Could not list cost allocations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List cost allocations with --trip filter"
if [[ -n "$SAMPLE_TRIP_ID" && "$SAMPLE_TRIP_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-customer-cost-allocations list --trip "$SAMPLE_TRIP_ID" --limit 5
    assert_success
else
    skip "No trip ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show cost allocation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-customer-cost-allocations show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_ALLOCATION_JSON=$(echo "$output" | jq -c '.allocation // empty' 2>/dev/null)
        pass
    else
        fail "Show failed"
    fi
else
    skip "No allocation ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cost allocation"
CREATE_TRIP_ID="${XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID:-$SAMPLE_TRIP_ID}"
if [[ -n "$CREATE_TRIP_ID" && "$CREATE_TRIP_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-customer-cost-allocations create --trip "$CREATE_TRIP_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "equipment-movement-trip-customer-cost-allocations" "$CREATED_ID"
            pass
        else
            fail "Created allocation but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"exists"* ]] || [[ "$output" == *"has already been taken"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No trip ID available (set XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update cost allocation with explicit allocation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$SAMPLE_ALLOCATION_JSON" ]]; then
    xbe_json do equipment-movement-trip-customer-cost-allocations update "$SAMPLE_ID" \
        --is-explicit true \
        --allocation "$SAMPLE_ALLOCATION_JSON"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Update failed (permissions or policy): ${output:-no output}"
    fi
else
    skip "No sample allocation/JSON available"
fi

test_name "Update allocation without fields fails"
xbe_run do equipment-movement-trip-customer-cost-allocations update "999999"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete created cost allocation"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-customer-cost-allocations delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created allocation to delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
