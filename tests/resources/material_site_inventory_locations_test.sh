#!/bin/bash
#
# XBE CLI Integration Tests: Material Site Inventory Locations
#
# Tests list, show, create, update, and delete operations for the
# material-site-inventory-locations resource.
#
# COVERAGE: List filters + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
SAMPLE_ID=""
CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
UNIT_OF_MEASURE_ID=""
ALT_UNIT_OF_MEASURE_ID=""
QUALIFIED_NAME=""

describe "Resource: material_site_inventory_locations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material site inventory locations"
xbe_json view material-site-inventory-locations list --limit 5
assert_success

test_name "List material site inventory locations returns array"
xbe_json view material-site-inventory-locations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material site inventory locations"
fi

# ============================================================================
# Sample Record (used for show fallback)
# ============================================================================

test_name "Capture sample material site inventory location"
xbe_json view material-site-inventory-locations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No material site inventory locations available for sample"
    fi
else
    skip "Could not list material site inventory locations to capture sample"
fi

# ============================================================================
# Prerequisites - Create broker, supplier, and material site
# ============================================================================

test_name "Create prerequisite broker for inventory location tests"
BROKER_NAME=$(unique_name "MSILTestBroker")

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

test_name "Create prerequisite material supplier for inventory location tests"
SUPPLIER_NAME=$(unique_name "MSILTestSupplier")

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
    fail "Failed to create material supplier"
    echo "Cannot continue without a material supplier"
    run_tests
fi

test_name "Create prerequisite material site for inventory location tests"
SITE_NAME=$(unique_name "MSILTestSite")

xbe_json do material-sites create \
    --name "$SITE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Test Quarry Rd, Chicago, IL 60601"

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
    fail "Failed to create material site"
    echo "Cannot continue without a material site"
    run_tests
fi

test_name "Find unit of measure"
xbe_json view unit-of-measures list --limit 2
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    ALT_UNIT_OF_MEASURE_ID=$(json_get ".[1].id")
    if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
        if [[ -z "$ALT_UNIT_OF_MEASURE_ID" || "$ALT_UNIT_OF_MEASURE_ID" == "null" ]]; then
            ALT_UNIT_OF_MEASURE_ID="$UNIT_OF_MEASURE_ID"
        fi
        pass
    else
        skip "No unit of measure available"
    fi
else
    skip "Could not list unit of measures"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material site inventory location with required fields"
QUALIFIED_NAME=$(unique_name "MSIL")
DISPLAY_NAME=$(unique_name "Inventory Location")
LATITUDE="41.881"
LONGITUDE="-87.623"

if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do material-site-inventory-locations create \
        --material-site "$CREATED_MATERIAL_SITE_ID" \
        --qualified-name "$QUALIFIED_NAME" \
        --display-name-explicit "$DISPLAY_NAME" \
        --latitude "$LATITUDE" \
        --longitude "$LONGITUDE" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID"
else
    xbe_json do material-site-inventory-locations create \
        --material-site "$CREATED_MATERIAL_SITE_ID" \
        --qualified-name "$QUALIFIED_NAME" \
        --display-name-explicit "$DISPLAY_NAME" \
        --latitude "$LATITUDE" \
        --longitude "$LONGITUDE"
fi

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "material-site-inventory-locations" "$CREATED_ID"
        pass
    else
        fail "Created material site inventory location but no ID returned"
    fi
else
    fail "Failed to create material site inventory location"
fi

if [[ -z "$CREATED_ID" || "$CREATED_ID" == "null" ]]; then
    echo "Cannot continue without a material site inventory location"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material site inventory location"
SHOW_ID="$CREATED_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$SAMPLE_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations show "$SHOW_ID"
    assert_success
else
    skip "No material site inventory location ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material site inventory location"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    UPDATED_NAME=$(unique_name "Inventory Location Updated")
    UPDATED_LATITUDE="41.882"
    UPDATED_LONGITUDE="-87.624"

    if [[ -n "$ALT_UNIT_OF_MEASURE_ID" && "$ALT_UNIT_OF_MEASURE_ID" != "null" ]]; then
        xbe_json do material-site-inventory-locations update "$CREATED_ID" \
            --display-name-explicit "$UPDATED_NAME" \
            --latitude "$UPDATED_LATITUDE" \
            --longitude "$UPDATED_LONGITUDE" \
            --unit-of-measure "$ALT_UNIT_OF_MEASURE_ID"
    else
        xbe_json do material-site-inventory-locations update "$CREATED_ID" \
            --display-name-explicit "$UPDATED_NAME" \
            --latitude "$UPDATED_LATITUDE" \
            --longitude "$UPDATED_LONGITUDE"
    fi

    assert_success
else
    skip "No material site inventory location ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List inventory locations with --qualified-name filter"
if [[ -n "$QUALIFIED_NAME" && "$QUALIFIED_NAME" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --qualified-name "$QUALIFIED_NAME" --limit 5
    assert_success
else
    skip "No qualified name available"
fi

test_name "List inventory locations with --material-site filter"
if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --material-site "$CREATED_MATERIAL_SITE_ID" --limit 5
    assert_success
else
    skip "No material site ID available"
fi

test_name "List inventory locations with --unit-of-measure filter"
if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --unit-of-measure "$UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No unit of measure ID available"
fi

test_name "List inventory locations with --broker-id filter"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --broker-id "$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List inventory locations with --broker filter"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --broker "$CREATED_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List inventory locations with --material-supplier-id filter"
if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --material-supplier-id "$CREATED_MATERIAL_SUPPLIER_ID" --limit 5
    assert_success
else
    skip "No material supplier ID available"
fi

test_name "List inventory locations with --material-supplier filter"
if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-site-inventory-locations list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --limit 5
    assert_success
else
    skip "No material supplier ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material site inventory location requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do material-site-inventory-locations delete "$CREATED_ID"
    assert_failure
else
    skip "No created material site inventory location ID available"
fi

test_name "Delete material site inventory location with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do material-site-inventory-locations delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created material site inventory location ID available"
fi

run_tests
