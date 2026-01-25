#!/bin/bash
#
# XBE CLI Integration Tests: Material Type Unavailabilities
#
# Tests CRUD operations for the material-type-unavailabilities resource.
# Material type unavailabilities require a supplier-specific material type.
#
# COVERAGE: create/update/delete + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_PARENT_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_UNAVAILABILITY_ID=""

START_AT="2025-01-01T00:00:00Z"
END_AT="2025-01-02T00:00:00Z"
UPDATED_START_AT="2024-12-31T12:00:00Z"
UPDATED_END_AT="2025-01-03T00:00:00Z"

DESCR_INITIAL="Initial unavailability"
DESCR_UPDATED="Updated unavailability"

describe "Resource: material-type-unavailabilities"

# ============================================================================
# Prerequisites - Create broker, material supplier, material type
# ============================================================================

test_name "Create prerequisite broker for material type unavailability tests"
BROKER_NAME=$(unique_name "MTUnavailBroker")

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

test_name "Create prerequisite material supplier for material type unavailability tests"
SUPPLIER_NAME=$(unique_name "MTUnavailSupplier")

xbe_json do material-suppliers create --name "$SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_SUPPLIER_ID" ]]; then
        CREATED_MATERIAL_SUPPLIER_ID="$XBE_TEST_MATERIAL_SUPPLIER_ID"
        echo "    Using XBE_TEST_MATERIAL_SUPPLIER_ID: $CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        fail "Failed to create material supplier and XBE_TEST_MATERIAL_SUPPLIER_ID not set"
        echo "Cannot continue without a material supplier"
        run_tests
    fi
fi

test_name "Create prerequisite material type for material type unavailability tests"
PARENT_MT_NAME=$(unique_name "MTUnavailParent")
MT_NAME=$(unique_name "MTUnavailType")

xbe_json do material-types create --name "$PARENT_MT_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_PARENT_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PARENT_MATERIAL_TYPE_ID" && "$CREATED_PARENT_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_PARENT_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created parent material type but no ID returned"
        echo "Cannot continue without a parent material type"
        run_tests
    fi
else
    fail "Failed to create parent material type"
    echo "Cannot continue without a parent material type"
    run_tests
fi

test_name "Create supplier-specific material type for material type unavailability tests"

xbe_json do material-types create \
    --name "$MT_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --parent-material-type "$CREATED_PARENT_MATERIAL_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create material type"
    echo "Cannot continue without a material type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material type unavailability with required fields"
xbe_json do material-type-unavailabilities create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --start-at "$START_AT" \
    --end-at "$END_AT" \
    --description "$DESCR_INITIAL"

if [[ $status -eq 0 ]]; then
    CREATED_UNAVAILABILITY_ID=$(json_get ".id")
    if [[ -n "$CREATED_UNAVAILABILITY_ID" && "$CREATED_UNAVAILABILITY_ID" != "null" ]]; then
        register_cleanup "material-type-unavailabilities" "$CREATED_UNAVAILABILITY_ID"
        pass
    else
        fail "Created unavailability but no ID returned"
    fi
else
    fail "Failed to create material type unavailability"
fi

# Only continue if we successfully created an unavailability
if [[ -z "$CREATED_UNAVAILABILITY_ID" || "$CREATED_UNAVAILABILITY_ID" == "null" ]]; then
    echo "Cannot continue without a valid material type unavailability ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material type unavailability by ID"
xbe_json view material-type-unavailabilities show "$CREATED_UNAVAILABILITY_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material type unavailability description"
xbe_json do material-type-unavailabilities update "$CREATED_UNAVAILABILITY_ID" --description "$DESCR_UPDATED"
assert_success

test_name "Update material type unavailability start-at"
xbe_json do material-type-unavailabilities update "$CREATED_UNAVAILABILITY_ID" --start-at "$UPDATED_START_AT"
assert_success

test_name "Update material type unavailability end-at"
xbe_json do material-type-unavailabilities update "$CREATED_UNAVAILABILITY_ID" --end-at "$UPDATED_END_AT"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material type unavailabilities"
xbe_json view material-type-unavailabilities list --limit 5
assert_success

test_name "List material type unavailabilities returns array"
xbe_json view material-type-unavailabilities list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material type unavailabilities"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List unavailabilities with --material-type filter"
if [[ -n "$CREATED_MATERIAL_TYPE_ID" && "$CREATED_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view material-type-unavailabilities list --material-type "$CREATED_MATERIAL_TYPE_ID" --limit 10
    assert_success
else
    skip "No material type ID available for filter test"
fi

test_name "List unavailabilities with --start-at-min filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view material-type-unavailabilities list --start-at-min "$START_AT" --limit 10
    assert_success
else
    skip "No start-at value available for filter test"
fi

test_name "List unavailabilities with --start-at-max filter"
if [[ -n "$UPDATED_START_AT" && "$UPDATED_START_AT" != "null" ]]; then
    xbe_json view material-type-unavailabilities list --start-at-max "$UPDATED_START_AT" --limit 10
    assert_success
else
    skip "No start-at value available for filter test"
fi

test_name "List unavailabilities with --end-at-min filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view material-type-unavailabilities list --end-at-min "$END_AT" --limit 10
    assert_success
else
    skip "No end-at value available for filter test"
fi

test_name "List unavailabilities with --end-at-max filter"
if [[ -n "$UPDATED_END_AT" && "$UPDATED_END_AT" != "null" ]]; then
    xbe_json view material-type-unavailabilities list --end-at-max "$UPDATED_END_AT" --limit 10
    assert_success
else
    skip "No end-at value available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material type unavailability"
xbe_json do material-type-unavailabilities delete "$CREATED_UNAVAILABILITY_ID" --confirm
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material type unavailability without time bounds fails"
xbe_json do material-type-unavailabilities create --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_failure

test_name "Update material type unavailability without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do material-type-unavailabilities update "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
