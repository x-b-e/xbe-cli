#!/bin/bash
#
# XBE CLI Integration Tests: Certification Types
#
# Tests CRUD operations for the certification_types resource.
# Certification types define the types of certifications that can be tracked
# for drivers, truckers, or equipment.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CERTIFICATION_TYPE_ID=""
CREATED_BROKER_ID=""

describe "Resource: certification_types"

# ============================================================================
# Prerequisites - Create broker for certification type tests
# ============================================================================

test_name "Create prerequisite broker for certification type tests"
BROKER_NAME=$(unique_name "CTTestBroker")

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

test_name "Create certification type with required fields"
TEST_NAME=$(unique_name "CertType")

xbe_json do certification-types create \
    --name "$TEST_NAME" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CERTIFICATION_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_CERTIFICATION_TYPE_ID" && "$CREATED_CERTIFICATION_TYPE_ID" != "null" ]]; then
        register_cleanup "certification-types" "$CREATED_CERTIFICATION_TYPE_ID"
        pass
    else
        fail "Created certification type but no ID returned"
    fi
else
    fail "Failed to create certification type"
fi

# Only continue if we successfully created a certification type
if [[ -z "$CREATED_CERTIFICATION_TYPE_ID" || "$CREATED_CERTIFICATION_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid certification type ID"
    run_tests
fi

test_name "Create certification type with requires-expiration"
TEST_NAME2=$(unique_name "CertType2")
xbe_json do certification-types create \
    --name "$TEST_NAME2" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID" \
    --requires-expiration
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "certification-types" "$id"
    pass
else
    fail "Failed to create certification type with requires-expiration"
fi

test_name "Create certification type with User can-apply-to"
TEST_NAME3=$(unique_name "CertType3")
xbe_json do certification-types create \
    --name "$TEST_NAME3" \
    --can-apply-to "User" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "certification-types" "$id"
    pass
else
    fail "Failed to create certification type with User can-apply-to"
fi

test_name "Create certification type with can-be-requirement-of"
TEST_NAME4=$(unique_name "CertType4")
xbe_json do certification-types create \
    --name "$TEST_NAME4" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID" \
    --can-be-requirement-of "Job"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "certification-types" "$id"
    pass
else
    fail "Failed to create certification type with can-be-requirement-of"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update certification type name"
UPDATED_NAME=$(unique_name "UpdatedCertType")
xbe_json do certification-types update "$CREATED_CERTIFICATION_TYPE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update certification type requires-expiration"
xbe_json do certification-types update "$CREATED_CERTIFICATION_TYPE_ID" --requires-expiration
assert_success

test_name "Update certification type can-be-requirement-of"
xbe_json do certification-types update "$CREATED_CERTIFICATION_TYPE_ID" --can-be-requirement-of "Job"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List certification types"
xbe_json view certification-types list --limit 5
assert_success

test_name "List certification types returns array"
xbe_json view certification-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list certification types"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List certification types with --broker filter"
xbe_json view certification-types list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List certification types with --name filter"
xbe_json view certification-types list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List certification types with --can-apply-to filter"
xbe_json view certification-types list --can-apply-to "Trucker" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List certification types with --limit"
xbe_json view certification-types list --limit 3
assert_success

test_name "List certification types with --offset"
xbe_json view certification-types list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete certification type requires --confirm flag"
xbe_run do certification-types delete "$CREATED_CERTIFICATION_TYPE_ID"
assert_failure

test_name "Delete certification type with --confirm"
# Create a certification type specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteCertType")
xbe_json do certification-types create \
    --name "$TEST_DEL_NAME" \
    --can-apply-to "Trucker" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do certification-types delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create certification type for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create certification type without name fails"
xbe_json do certification-types create --can-apply-to "Trucker" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create certification type without can-apply-to fails"
xbe_json do certification-types create --name "Test" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create certification type without broker fails"
xbe_json do certification-types create --name "Test" --can-apply-to "Trucker"
assert_failure

test_name "Update without any fields fails"
xbe_json do certification-types update "$CREATED_CERTIFICATION_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
