#!/bin/bash
#
# XBE CLI Integration Tests: Developer Reference Types
#
# Tests CRUD operations for the developer_reference_types resource.
# Developer reference types define custom reference fields for developers.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_REF_TYPE_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""

describe "Resource: developer_reference_types"

# ============================================================================
# Prerequisites - Create broker and developer
# ============================================================================

test_name "Create prerequisite broker for developer reference type tests"
BROKER_NAME=$(unique_name "DRTTestBroker")

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

test_name "Create prerequisite developer"
DEVELOPER_NAME=$(unique_name "DRTTestDev")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create developer reference type with required fields"
TEST_NAME=$(unique_name "DevRefType")

xbe_json do developer-reference-types create \
    --name "$TEST_NAME" \
    --developer "$CREATED_DEVELOPER_ID" \
    --subject-types "Project"

if [[ $status -eq 0 ]]; then
    CREATED_REF_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_REF_TYPE_ID" && "$CREATED_REF_TYPE_ID" != "null" ]]; then
        register_cleanup "developer-reference-types" "$CREATED_REF_TYPE_ID"
        pass
    else
        fail "Created developer reference type but no ID returned"
    fi
else
    fail "Failed to create developer reference type"
fi

# Only continue if we successfully created a developer reference type
if [[ -z "$CREATED_REF_TYPE_ID" || "$CREATED_REF_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid developer reference type ID"
    run_tests
fi

test_name "Create developer reference type with subject-types"
TEST_NAME2=$(unique_name "DevRefType2")
xbe_json do developer-reference-types create \
    --name "$TEST_NAME2" \
    --developer "$CREATED_DEVELOPER_ID" \
    --subject-types "Project"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "developer-reference-types" "$id"
    pass
else
    fail "Failed to create developer reference type with subject-types"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update developer reference type name"
UPDATED_NAME=$(unique_name "UpdatedDRT")
xbe_json do developer-reference-types update "$CREATED_REF_TYPE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update developer reference type subject-types"
xbe_json do developer-reference-types update "$CREATED_REF_TYPE_ID" --subject-types "Project"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List developer reference types"
xbe_json view developer-reference-types list --limit 5
assert_success

test_name "List developer reference types returns array"
xbe_json view developer-reference-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list developer reference types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List developer reference types with --developer filter"
xbe_json view developer-reference-types list --developer "$CREATED_DEVELOPER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List developer reference types with --limit"
xbe_json view developer-reference-types list --limit 3
assert_success

test_name "List developer reference types with --offset"
xbe_json view developer-reference-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete developer reference type requires --confirm flag"
xbe_run do developer-reference-types delete "$CREATED_REF_TYPE_ID"
assert_failure

test_name "Delete developer reference type with --confirm"
# Create a developer reference type specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteDRT")
xbe_json do developer-reference-types create \
    --name "$TEST_DEL_NAME" \
    --developer "$CREATED_DEVELOPER_ID" \
    --subject-types "Project"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do developer-reference-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create developer reference type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create developer reference type without name fails"
xbe_json do developer-reference-types create --developer "$CREATED_DEVELOPER_ID"
assert_failure

test_name "Create developer reference type without developer fails"
xbe_json do developer-reference-types create --name "NoDeveloper"
assert_failure

test_name "Update without any fields fails"
xbe_json do developer-reference-types update "$CREATED_REF_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
