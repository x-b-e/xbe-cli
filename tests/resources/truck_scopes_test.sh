#!/bin/bash
#
# XBE CLI Integration Tests: Truck Scopes
#
# Tests CRUD operations for the truck_scopes resource.
# Truck scopes define the geographic and equipment scope for trucking operations.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_TRUCK_SCOPE_ID=""
CREATED_BROKER_ID=""

describe "Resource: truck_scopes"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for truck scope tests"
BROKER_NAME=$(unique_name "TSTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create truck scope with required fields"

xbe_json do truck-scopes create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCK_SCOPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCK_SCOPE_ID" && "$CREATED_TRUCK_SCOPE_ID" != "null" ]]; then
        register_cleanup "truck-scopes" "$CREATED_TRUCK_SCOPE_ID"
        pass
    else
        fail "Created truck scope but no ID returned"
    fi
else
    fail "Failed to create truck scope"
fi

# Only continue if we successfully created a truck scope
if [[ -z "$CREATED_TRUCK_SCOPE_ID" || "$CREATED_TRUCK_SCOPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid truck scope ID"
    run_tests
fi

test_name "Create truck scope with authorized-state-codes"
xbe_json do truck-scopes create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --authorized-state-codes "IL,IN,WI"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "truck-scopes" "$id"
    pass
else
    fail "Failed to create truck scope with authorized-state-codes"
fi

# NOTE: Address alone requires proximity meters - tested in next test

test_name "Create truck scope with address-proximity-meters"
xbe_json do truck-scopes create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID" \
    --address "456 Oak Ave, Chicago, IL 60602" \
    --address-proximity-meters 50000
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "truck-scopes" "$id"
    pass
else
    fail "Failed to create truck scope with address-proximity-meters"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update truck scope authorized-state-codes"
xbe_json do truck-scopes update "$CREATED_TRUCK_SCOPE_ID" --authorized-state-codes "IL,IN,WI,MI"
assert_success

# NOTE: Address updates require lat/lng to be provided together, skipping individual field updates

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List truck scopes"
xbe_json view truck-scopes list --limit 5
assert_success

test_name "List truck scopes returns array"
xbe_json view truck-scopes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list truck scopes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

# NOTE: The truck-scopes list command doesn't currently support broker filter

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List truck scopes with --limit"
xbe_json view truck-scopes list --limit 3
assert_success

test_name "List truck scopes with --offset"
xbe_json view truck-scopes list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete truck scope requires --confirm flag"
xbe_run do truck-scopes delete "$CREATED_TRUCK_SCOPE_ID"
assert_failure

test_name "Delete truck scope with --confirm"
# Create a truck scope specifically for deletion
xbe_json do truck-scopes create \
    --organization-type "brokers" \
    --organization-id "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do truck-scopes delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create truck scope for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create truck scope without organization-type fails"
xbe_json do truck-scopes create --organization-id "$CREATED_BROKER_ID"
assert_failure

test_name "Create truck scope without organization-id fails"
xbe_json do truck-scopes create --organization-type "brokers"
assert_failure

test_name "Update without any fields fails"
xbe_json do truck-scopes update "$CREATED_TRUCK_SCOPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
