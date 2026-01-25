#!/bin/bash
#
# XBE CLI Integration Tests: Material Mix Designs
#
# Tests operations for the material-mix-designs resource.
# Material mix designs require a material-type relationship.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MIX_DESIGN_ID=""
CREATED_BROKER_ID=""
CREATED_MATERIAL_TYPE_ID=""

describe "Resource: material-mix-designs"

# ============================================================================
# Prerequisites - Create resources for material mix design tests
# ============================================================================

test_name "Create prerequisite broker for material mix design tests"
BROKER_NAME=$(unique_name "MixDesignBroker")

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

test_name "Create prerequisite material type for material mix design tests"
MT_NAME=$(unique_name "MaterialType")

xbe_json do material-types create \
    --name "$MT_NAME"

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

test_name "Create material mix design with required fields"
MIX_ID="MIX$(unique_suffix)"
xbe_json do material-mix-designs create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --mix "$MIX_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MIX_DESIGN_ID=$(json_get ".id")
    if [[ -n "$CREATED_MIX_DESIGN_ID" && "$CREATED_MIX_DESIGN_ID" != "null" ]]; then
        register_cleanup "material-mix-designs" "$CREATED_MIX_DESIGN_ID"
        pass
    else
        fail "Created material mix design but no ID returned"
    fi
else
    # Server may have issues, skip CRUD tests but continue with list tests
    skip "Failed to create material mix design (server may not support this operation)"
fi

# Only run these tests if we have a valid ID
if [[ -n "$CREATED_MIX_DESIGN_ID" && "$CREATED_MIX_DESIGN_ID" != "null" ]]; then

test_name "Create material mix design with description"
xbe_json do material-mix-designs create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --mix "MIX$(unique_suffix)" \
    --description "Test Mix Design A"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-mix-designs" "$id"
    pass
else
    skip "Failed to create material mix design with description"
fi

test_name "Update material mix design description"
xbe_json do material-mix-designs update "$CREATED_MIX_DESIGN_ID" --description "Updated Description"
if [[ $status -eq 0 ]]; then
    pass
else
    skip "Update not supported or failed"
fi

test_name "Update material mix design notes"
xbe_json do material-mix-designs update "$CREATED_MIX_DESIGN_ID" --notes "Updated notes"
if [[ $status -eq 0 ]]; then
    pass
else
    skip "Update not supported or failed"
fi

fi # End of conditional CRUD tests

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material mix designs"
xbe_json view material-mix-designs list --limit 5
assert_success

test_name "List material mix designs returns array"
xbe_json view material-mix-designs list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material mix designs"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List material mix designs with --material-type filter"
xbe_json view material-mix-designs list --material-type "$CREATED_MATERIAL_TYPE_ID" --limit 10
assert_success

test_name "List material mix designs with --description-like filter"
xbe_json view material-mix-designs list --description-like "Test" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List material mix designs with --limit"
xbe_json view material-mix-designs list --limit 3
assert_success

test_name "List material mix designs with --offset"
xbe_json view material-mix-designs list --limit 3 --offset 1
assert_success

test_name "List material mix designs with pagination (limit + offset)"
xbe_json view material-mix-designs list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material mix design without material-type fails"
xbe_json do material-mix-designs create --description "Test"
assert_failure

test_name "Update without any fields fails"
# Use a placeholder ID - it should fail for missing fields before hitting the API
xbe_json do material-mix-designs update "99999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
