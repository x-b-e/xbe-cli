#!/bin/bash
#
# XBE CLI Integration Tests: External Identification Types
#
# Tests CRUD operations for the external_identification_types resource.
# External identification types define kinds of external IDs that can be
# associated with entities (e.g., license numbers for truckers).
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EXT_ID_TYPE_ID=""

describe "Resource: external_identification_types"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create external identification type with required fields"
TEST_NAME=$(unique_name "ExtIDType")

xbe_json do external-identification-types create \
    --name "$TEST_NAME" \
    --can-apply-to "Trucker"

if [[ $status -eq 0 ]]; then
    CREATED_EXT_ID_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_EXT_ID_TYPE_ID" && "$CREATED_EXT_ID_TYPE_ID" != "null" ]]; then
        register_cleanup "external-identification-types" "$CREATED_EXT_ID_TYPE_ID"
        pass
    else
        fail "Created external identification type but no ID returned"
    fi
else
    fail "Failed to create external identification type"
fi

# Only continue if we successfully created an external identification type
if [[ -z "$CREATED_EXT_ID_TYPE_ID" || "$CREATED_EXT_ID_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid external identification type ID"
    run_tests
fi

# NOTE: Multiple can-apply-to values would be tested here but the CLI requires
# specific model names that must match exactly. Single values work well.

test_name "Create external identification type for Broker"
TEST_NAME2=$(unique_name "ExtIDType2")
xbe_json do external-identification-types create \
    --name "$TEST_NAME2" \
    --can-apply-to "Broker"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "external-identification-types" "$id"
    pass
else
    fail "Failed to create external identification type for Broker"
fi

test_name "Create external identification type with format-validation-regex"
TEST_NAME3=$(unique_name "ExtIDType3")
xbe_json do external-identification-types create \
    --name "$TEST_NAME3" \
    --can-apply-to "Trucker" \
    --format-validation-regex "^[A-Z]{2}[0-9]{6}$"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "external-identification-types" "$id"
    pass
else
    fail "Failed to create external identification type with format-validation-regex"
fi

test_name "Create external identification type with value-should-be-globally-unique"
TEST_NAME4=$(unique_name "ExtIDType4")
xbe_json do external-identification-types create \
    --name "$TEST_NAME4" \
    --can-apply-to "Broker" \
    --value-should-be-globally-unique
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "external-identification-types" "$id"
    pass
else
    fail "Failed to create external identification type with value-should-be-globally-unique"
fi

test_name "Create external identification type with all optional fields"
TEST_NAME5=$(unique_name "ExtIDType5")
xbe_json do external-identification-types create \
    --name "$TEST_NAME5" \
    --can-apply-to "Trucker" \
    --format-validation-regex "^[0-9]+$" \
    --value-should-be-globally-unique
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "external-identification-types" "$id"
    pass
else
    fail "Failed to create external identification type with all optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update external identification type name"
UPDATED_NAME=$(unique_name "UpdatedEIT")
xbe_json do external-identification-types update "$CREATED_EXT_ID_TYPE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update external identification type can-apply-to"
xbe_json do external-identification-types update "$CREATED_EXT_ID_TYPE_ID" --can-apply-to "Broker"
assert_success

test_name "Update external identification type format-validation-regex"
xbe_json do external-identification-types update "$CREATED_EXT_ID_TYPE_ID" --format-validation-regex "^[0-9]{4}$"
assert_success

test_name "Update external identification type value-should-be-globally-unique"
xbe_json do external-identification-types update "$CREATED_EXT_ID_TYPE_ID" --value-should-be-globally-unique=false
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List external identification types"
xbe_json view external-identification-types list --limit 5
assert_success

test_name "List external identification types returns array"
xbe_json view external-identification-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list external identification types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

# NOTE: --name filter is not supported by the server

test_name "List external identification types with --can-apply-to filter"
xbe_json view external-identification-types list --can-apply-to "Trucker" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List external identification types with --limit"
xbe_json view external-identification-types list --limit 3
assert_success

test_name "List external identification types with --offset"
xbe_json view external-identification-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete external identification type requires --confirm flag"
xbe_run do external-identification-types delete "$CREATED_EXT_ID_TYPE_ID"
assert_failure

test_name "Delete external identification type with --confirm"
# Create an external identification type specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteEIT")
xbe_json do external-identification-types create \
    --name "$TEST_DEL_NAME" \
    --can-apply-to "Trucker"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do external-identification-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create external identification type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create external identification type without name fails"
xbe_json do external-identification-types create --can-apply-to "Trucker"
assert_failure

test_name "Create external identification type without can-apply-to fails"
xbe_json do external-identification-types create --name "NoCanApplyTo"
assert_failure

test_name "Update without any fields fails"
xbe_json do external-identification-types update "$CREATED_EXT_ID_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
