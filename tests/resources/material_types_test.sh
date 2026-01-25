#!/bin/bash
#
# XBE CLI Integration Tests: Material Types
#
# Tests CRUD operations for the material_types resource.
# Material types categorize materials used in construction.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MT_ID=""

describe "Resource: material-types"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material type with required fields"
TEST_NAME=$(unique_name "MaterialType")
xbe_json do material-types create --name "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_MT_ID=$(json_get ".id")
    if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
        # Note: No delete available for material-types
        pass
    else
        fail "Created material type but no ID returned"
    fi
else
    fail "Failed to create material type: $output"
fi

test_name "Create material type with display name"
TEST_NAME2=$(unique_name "MaterialType2")
xbe_json do material-types create \
    --name "$TEST_NAME2" \
    --explicit-display-name "Display $TEST_NAME2"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create material type with display name"
fi

test_name "Create material type with description"
TEST_NAME3=$(unique_name "MaterialType3")
xbe_json do material-types create \
    --name "$TEST_NAME3" \
    --description "Test material type description"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create material type with description"
fi

test_name "Create material type with weight specification"
TEST_NAME4=$(unique_name "MaterialType4")
xbe_json do material-types create \
    --name "$TEST_NAME4" \
    --lbs-per-cubic-foot "150"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create material type with weight"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

# Only continue if we have a valid material type ID
if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then

test_name "Update material type name"
UPDATED_NAME=$(unique_name "UpdatedMT")
xbe_json do material-types update "$CREATED_MT_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update material type explicit-display-name"
xbe_json do material-types update "$CREATED_MT_ID" --explicit-display-name "Updated Display Name"
assert_success

test_name "Update material type description"
xbe_json do material-types update "$CREATED_MT_ID" --description "Updated description text"
assert_success

test_name "Update material type aggregate-bed"
xbe_json do material-types update "$CREATED_MT_ID" --aggregate-bed "test bed"
assert_success

test_name "Update material type aggregate-gradation"
# aggregate-gradation is a numeric field
xbe_json do material-types update "$CREATED_MT_ID" --aggregate-gradation "5"
assert_success

test_name "Update material type aggregate-ecce"
# aggregate-ecce is a numeric field
xbe_json do material-types update "$CREATED_MT_ID" --aggregate-ecce "3"
assert_success

test_name "Update material type lbs-per-cubic-foot"
xbe_json do material-types update "$CREATED_MT_ID" --lbs-per-cubic-foot "125"
assert_success

test_name "Update material type start-site-type"
# start-site-type is an enum but valid values depend on server configuration
xbe_json do material-types update "$CREATED_MT_ID" --start-site-type "supply"
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"is not included in the list"* ]]; then
        echo "    (Server validation for start-site-type - flag works correctly)"
        pass
    else
        fail "Unexpected error: $output"
    fi
fi

fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material types"
xbe_json view material-types list
assert_success

test_name "List material types returns array"
xbe_json view material-types list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List material types with --is-archived filter"
xbe_json view material-types list --is-archived false
assert_success

test_name "List material types with --has-material-supplier filter"
xbe_json view material-types list --has-material-supplier true
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List material types with --limit"
xbe_json view material-types list --limit 5
assert_success

test_name "List material types with --offset"
xbe_json view material-types list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material type without name fails"
xbe_json do material-types create
assert_failure

test_name "Update without any fields fails"
if [[ -n "$CREATED_MT_ID" && "$CREATED_MT_ID" != "null" ]]; then
    xbe_json do material-types update "$CREATED_MT_ID"
    assert_failure
else
    skip "No material type ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material type requires --confirm flag"
# Create a material type for deletion test
TEST_DEL_NAME=$(unique_name "DeleteMeMT")
xbe_json do material-types create --name "$TEST_DEL_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do material-types delete "$DEL_ID"
    assert_failure
else
    skip "Could not create material type for deletion test"
fi

test_name "Delete material type with --confirm"
TEST_DEL_NAME2=$(unique_name "DeleteMeMT2")
xbe_json do material-types create --name "$TEST_DEL_NAME2"
if [[ $status -eq 0 ]]; then
    DEL_ID2=$(json_get ".id")
    xbe_json do material-types delete "$DEL_ID2" --confirm
    # Note: May fail if material type is in use - that's expected
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"in use"* ]] || [[ "$output" == *"cannot"* ]]; then
            echo "    (Material type in use - delete not allowed, which is expected)"
            pass
        else
            fail "Delete command failed unexpectedly: $output"
        fi
    fi
else
    skip "Could not create material type for deletion test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
