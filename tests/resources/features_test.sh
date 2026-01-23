#!/bin/bash
#
# XBE CLI Integration Tests: Features
#
# Tests CRUD operations for the features resource.
# Features are product capabilities tracked in the system.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FEATURE_ID=""

describe "Resource: features"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create feature with required fields"
TEST_NAME=$(unique_name "Feature")
xbe_json do features create --name-generic "$TEST_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_FEATURE_ID=$(json_get ".id")
    if [[ -n "$CREATED_FEATURE_ID" && "$CREATED_FEATURE_ID" != "null" ]]; then
        register_cleanup "features" "$CREATED_FEATURE_ID"
        pass
    else
        fail "Created feature but no ID returned"
    fi
else
    fail "Failed to create feature"
fi

# Only continue if we successfully created a feature
if [[ -z "$CREATED_FEATURE_ID" || "$CREATED_FEATURE_ID" == "null" ]]; then
    echo "Cannot continue without a valid feature ID"
    run_tests
fi

test_name "Create feature with all attributes"
TEST_NAME2=$(unique_name "FeatureFull")
xbe_json do features create \
    --name-generic "$TEST_NAME2" \
    --name-branded "Branded $TEST_NAME2" \
    --description "Test feature description" \
    --pdca-stage "plan" \
    --differentiation-degree "2" \
    --scale "3"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "features" "$id"
    pass
else
    fail "Failed to create feature with all attributes"
fi

# ============================================================================
# UPDATE Tests - String Attributes
# ============================================================================

test_name "Update feature name-generic"
UPDATED_NAME=$(unique_name "UpdatedFeature")
xbe_json do features update "$CREATED_FEATURE_ID" --name-generic "$UPDATED_NAME"
assert_success

test_name "Update feature name-branded"
xbe_json do features update "$CREATED_FEATURE_ID" --name-branded "Updated Branded Name"
assert_success

test_name "Update feature description"
xbe_json do features update "$CREATED_FEATURE_ID" --description "Updated description"
assert_success

test_name "Update feature pdca-stage to plan"
xbe_json do features update "$CREATED_FEATURE_ID" --pdca-stage "plan"
assert_success

test_name "Update feature pdca-stage to do"
xbe_json do features update "$CREATED_FEATURE_ID" --pdca-stage "do"
assert_success

test_name "Update feature pdca-stage to check"
xbe_json do features update "$CREATED_FEATURE_ID" --pdca-stage "check"
assert_success

test_name "Update feature pdca-stage to act"
xbe_json do features update "$CREATED_FEATURE_ID" --pdca-stage "act"
assert_success

# ============================================================================
# UPDATE Tests - Integer Attributes
# ============================================================================

test_name "Update feature differentiation-degree"
xbe_json do features update "$CREATED_FEATURE_ID" --differentiation-degree "3"
assert_success

test_name "Update feature scale"
xbe_json do features update "$CREATED_FEATURE_ID" --scale "5"
assert_success

# ============================================================================
# UPDATE Tests - Date Attributes
# ============================================================================

test_name "Update feature released-on"
xbe_json do features update "$CREATED_FEATURE_ID" --released-on "2024-06-01"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List features"
xbe_json view features list
assert_success

test_name "List features returns array"
xbe_json view features list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list features"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List features with --pdca-stage filter"
xbe_json view features list --pdca-stage "do"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List features with --limit"
xbe_json view features list --limit 5
assert_success

test_name "List features with --offset"
xbe_json view features list --limit 5 --offset 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete feature requires --confirm flag"
xbe_json do features delete "$CREATED_FEATURE_ID"
assert_failure

test_name "Delete feature with --confirm"
# Create a feature specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMe")
xbe_json do features create --name-generic "$TEST_DEL_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do features delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create feature for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create feature without name-generic fails"
xbe_json do features create
assert_failure

test_name "Update without any fields fails"
xbe_json do features update "$CREATED_FEATURE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
