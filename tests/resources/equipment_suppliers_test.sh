#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Suppliers
#
# Tests CRUD operations for the equipment_suppliers resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EQUIPMENT_SUPPLIER_ID=""
CREATED_BROKER_ID=""

describe "Resource: equipment_suppliers"

# ============================================================================
# Prerequisites - Create broker for tests
# ============================================================================

test_name "Create prerequisite broker for equipment supplier tests"
BROKER_NAME=$(unique_name "EquipSupBroker")

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

test_name "Create equipment supplier with required fields"
TEST_NAME=$(unique_name "EquipSupplier")

xbe_json do equipment-suppliers create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_SUPPLIER_ID" && "$CREATED_EQUIPMENT_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "equipment-suppliers" "$CREATED_EQUIPMENT_SUPPLIER_ID"
        pass
    else
        fail "Created equipment supplier but no ID returned"
    fi
else
    fail "Failed to create equipment supplier"
fi

# Only continue if we successfully created an equipment supplier
if [[ -z "$CREATED_EQUIPMENT_SUPPLIER_ID" || "$CREATED_EQUIPMENT_SUPPLIER_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment supplier ID"
    run_tests
fi

test_name "Create equipment supplier with contract-number"
TEST_NAME2=$(unique_name "EquipSupplier2")
CONTRACT_NUMBER="EQ-$(unique_suffix)"
xbe_json do equipment-suppliers create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --contract-number "$CONTRACT_NUMBER"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-suppliers" "$id"
    pass
else
    fail "Failed to create equipment supplier with contract-number"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update equipment supplier name"
UPDATED_NAME=$(unique_name "UpdatedEquipSup")
xbe_json do equipment-suppliers update "$CREATED_EQUIPMENT_SUPPLIER_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update equipment supplier contract-number"
UPDATED_CONTRACT_NUMBER="EQ-$(unique_suffix)"
xbe_json do equipment-suppliers update "$CREATED_EQUIPMENT_SUPPLIER_ID" --contract-number "$UPDATED_CONTRACT_NUMBER"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment supplier"
xbe_json view equipment-suppliers show "$CREATED_EQUIPMENT_SUPPLIER_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment suppliers"
xbe_json view equipment-suppliers list --limit 5
assert_success

test_name "List equipment suppliers returns array"
xbe_json view equipment-suppliers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment suppliers"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List equipment suppliers with --name filter"
xbe_json view equipment-suppliers list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List equipment suppliers with --broker filter"
xbe_json view equipment-suppliers list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List equipment suppliers with --limit"
xbe_json view equipment-suppliers list --limit 3
assert_success

test_name "List equipment suppliers with --offset"
xbe_json view equipment-suppliers list --limit 3 --offset 3
assert_success

test_name "List equipment suppliers with pagination (limit + offset)"
xbe_json view equipment-suppliers list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment supplier requires --confirm flag"
xbe_run do equipment-suppliers delete "$CREATED_EQUIPMENT_SUPPLIER_ID"
assert_failure

test_name "Delete equipment supplier with --confirm"
TEST_DEL_NAME=$(unique_name "DeleteEquipSup")
xbe_json do equipment-suppliers create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do equipment-suppliers delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create equipment supplier for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment supplier without name fails"
xbe_json do equipment-suppliers create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create equipment supplier without broker fails"
xbe_json do equipment-suppliers create --name "Test Equipment Supplier"
assert_failure

test_name "Update without any fields fails"
xbe_json do equipment-suppliers update "$CREATED_EQUIPMENT_SUPPLIER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
