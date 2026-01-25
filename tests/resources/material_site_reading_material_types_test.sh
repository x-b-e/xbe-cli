#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Reading Material Types
#
# Tests CRUD operations for the material_site_reading_material_types resource.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_BROKER_ID=""

describe "Resource: material-site-reading-material-types"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for material site reading material type tests"
BROKER_NAME=$(unique_name "MSRMTBroker")

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

test_name "Create prerequisite material supplier for material site reading material type tests"
SUPPLIER_NAME=$(unique_name "MSRMTSupplier")

xbe_json do material-suppliers create \
    --name "$SUPPLIER_NAME" \
    --broker "$CREATED_BROKER_ID"

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

test_name "Create prerequisite material site for material site reading material type tests"
MATERIAL_SITE_NAME=$(unique_name "MSRMTMaterialSite")

xbe_json do material-sites create \
    --name "$MATERIAL_SITE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "150 Test Quarry Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Created material site but no ID returned"
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_SITE_ID" ]]; then
        CREATED_MATERIAL_SITE_ID="$XBE_TEST_MATERIAL_SITE_ID"
        echo "    Using XBE_TEST_MATERIAL_SITE_ID: $CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Failed to create material site and XBE_TEST_MATERIAL_SITE_ID not set"
        echo "Cannot continue without a material site"
        run_tests
    fi
fi

test_name "Create prerequisite material type"
MATERIAL_TYPE_NAME=$(unique_name "MSRMTMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

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
    if [[ -n "$XBE_TEST_MATERIAL_TYPE_ID" ]]; then
        CREATED_MATERIAL_TYPE_ID="$XBE_TEST_MATERIAL_TYPE_ID"
        echo "    Using XBE_TEST_MATERIAL_TYPE_ID: $CREATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Failed to create material type and XBE_TEST_MATERIAL_TYPE_ID not set"
        echo "Cannot continue without a material type"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material site reading material type with required fields"
EXTERNAL_ID=$(unique_name "MSRMTExternal")

xbe_json do material-site-reading-material-types create \
    --material-site "$CREATED_MATERIAL_SITE_ID" \
    --external-id "$EXTERNAL_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "material-site-reading-material-types" "$CREATED_ID"
        pass
    else
        fail "Created material site reading material type but no ID returned"
        run_tests
    fi
else
    fail "Failed to create material site reading material type: $output"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material site reading material type external-id"
UPDATED_EXTERNAL_ID=$(unique_name "MSRMTExternalUpdated")
xbe_json do material-site-reading-material-types update "$CREATED_ID" --external-id "$UPDATED_EXTERNAL_ID"
assert_success

test_name "Clear material type relationship"
xbe_json do material-site-reading-material-types update "$CREATED_ID" --material-type ""
assert_success

test_name "Set material type relationship"
xbe_json do material-site-reading-material-types update "$CREATED_ID" --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material site reading material type"
xbe_json view material-site-reading-material-types show "$CREATED_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show material site reading material type"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List material site reading material types"
xbe_json view material-site-reading-material-types list
assert_success

test_name "List material site reading material types with --material-site filter"
xbe_json view material-site-reading-material-types list --material-site "$CREATED_MATERIAL_SITE_ID"
assert_success

test_name "List material site reading material types with --material-type filter"
xbe_json view material-site-reading-material-types list --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

test_name "List material site reading material types with --external-id filter"
xbe_json view material-site-reading-material-types list --external-id "$UPDATED_EXTERNAL_ID"
assert_success

test_name "List material site reading material types with --limit"
xbe_json view material-site-reading-material-types list --limit 5
assert_success

test_name "List material site reading material types with --offset"
xbe_json view material-site-reading-material-types list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create requires --material-site"
xbe_run do material-site-reading-material-types create --external-id "EXT-FAIL"
assert_failure

test_name "Create requires --external-id"
xbe_run do material-site-reading-material-types create --material-site "$CREATED_MATERIAL_SITE_ID"
assert_failure

test_name "Update without fields fails"
xbe_run do material-site-reading-material-types update "$CREATED_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
xbe_run do material-site-reading-material-types delete "$CREATED_ID"
assert_failure

test_name "Delete material site reading material type with --confirm"
xbe_json do material-site-reading-material-types delete "$CREATED_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
