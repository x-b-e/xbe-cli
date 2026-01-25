#!/bin/bash
#
# XBE CLI Integration Tests: Material Suppliers
#
# Tests CRUD operations for the material_suppliers resource.
# Material suppliers are companies that supply materials.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_BROKER_ID=""

describe "Resource: material_suppliers"

# ============================================================================
# Prerequisites - Create broker for tests
# ============================================================================

test_name "Create prerequisite broker for material supplier tests"
BROKER_NAME=$(unique_name "MSupTestBroker")

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

test_name "Create material supplier with required fields"
TEST_NAME=$(unique_name "MatSupplier")

xbe_json do material-suppliers create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
    fi
else
    fail "Failed to create material supplier"
fi

# Only continue if we successfully created a material supplier
if [[ -z "$CREATED_MATERIAL_SUPPLIER_ID" || "$CREATED_MATERIAL_SUPPLIER_ID" == "null" ]]; then
    echo "Cannot continue without a valid material supplier ID"
    run_tests
fi

test_name "Create material supplier with url"
TEST_NAME2=$(unique_name "MatSupplier2")
xbe_json do material-suppliers create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --url "https://example.com/supplier"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-suppliers" "$id"
    pass
else
    fail "Failed to create material supplier with url"
fi

test_name "Create material supplier with phone-number"
TEST_NAME3=$(unique_name "MatSupplier3")
TEST_PHONE=$(unique_mobile)
xbe_json do material-suppliers create \
    --name "$TEST_NAME3" \
    --broker "$CREATED_BROKER_ID" \
    --phone-number "$TEST_PHONE"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-suppliers" "$id"
    pass
else
    fail "Failed to create material supplier with phone-number"
fi

test_name "Create material supplier with active=false"
TEST_NAME4=$(unique_name "MatSupplier4")
xbe_json do material-suppliers create \
    --name "$TEST_NAME4" \
    --broker "$CREATED_BROKER_ID" \
    --active=false
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-suppliers" "$id"
    pass
else
    fail "Failed to create material supplier with active=false"
fi

test_name "Create material supplier with is-controlled-by-broker"
TEST_NAME5=$(unique_name "MatSupplier5")
xbe_json do material-suppliers create \
    --name "$TEST_NAME5" \
    --broker "$CREATED_BROKER_ID" \
    --is-controlled-by-broker
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-suppliers" "$id"
    pass
else
    fail "Failed to create material supplier with is-controlled-by-broker"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material supplier name"
UPDATED_NAME=$(unique_name "UpdatedMSup")
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update material supplier url"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --url "https://updated-supplier.com"
assert_success

test_name "Update material supplier phone-number"
UPDATED_PHONE=$(unique_mobile)
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --phone-number "$UPDATED_PHONE"
assert_success

test_name "Update material supplier to inactive"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --no-active
assert_success

test_name "Update material supplier to active"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --active
assert_success

test_name "Update material supplier is-controlled-by-broker to true"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --is-controlled-by-broker
assert_success

test_name "Update material supplier is-controlled-by-broker to false"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID" --no-is-controlled-by-broker
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material suppliers"
xbe_json view material-suppliers list --limit 5
assert_success

test_name "List material suppliers returns array"
xbe_json view material-suppliers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material suppliers"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List material suppliers with --name filter"
xbe_json view material-suppliers list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List material suppliers with --active filter"
xbe_json view material-suppliers list --active --limit 10
assert_success

test_name "List material suppliers with --broker filter"
xbe_json view material-suppliers list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List material suppliers with --is-broker-active filter"
xbe_json view material-suppliers list --is-broker-active true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List material suppliers with --limit"
xbe_json view material-suppliers list --limit 3
assert_success

test_name "List material suppliers with --offset"
xbe_json view material-suppliers list --limit 3 --offset 3
assert_success

test_name "List material suppliers with pagination (limit + offset)"
xbe_json view material-suppliers list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material supplier requires --confirm flag"
xbe_run do material-suppliers delete "$CREATED_MATERIAL_SUPPLIER_ID"
assert_failure

test_name "Delete material supplier with --confirm"
# Create a material supplier specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMeMSup")
xbe_json do material-suppliers create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do material-suppliers delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create material supplier for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material supplier without name fails"
xbe_json do material-suppliers create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create material supplier without broker fails"
xbe_json do material-suppliers create --name "Test Material Supplier"
assert_failure

test_name "Update without any fields fails"
xbe_json do material-suppliers update "$CREATED_MATERIAL_SUPPLIER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
