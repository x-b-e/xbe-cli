#!/bin/bash
#
# XBE CLI Integration Tests: Inventory Capacities
#
# Tests CRUD operations for the inventory-capacities resource.
# Inventory capacities define min/max levels and threshold alerts for
# material sites and material types.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID2=""
CREATED_INVENTORY_CAPACITY_ID=""
DEL_ID=""

describe "Resource: inventory-capacities"

# ============================================================================
# Prerequisites - Create broker, material supplier, material site, material types
# ============================================================================

test_name "Create prerequisite broker for inventory capacity tests"
BROKER_NAME=$(unique_name "InvCapBroker")

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

test_name "Create prerequisite material supplier for inventory capacity tests"
SUPPLIER_NAME=$(unique_name "InvCapSupplier")

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

test_name "Create prerequisite material site for inventory capacity tests"
SITE_NAME=$(unique_name "InvCapSite")

xbe_json do material-sites create \
    --name "$SITE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Inventory Rd, Chicago, IL 60601"

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

test_name "Create prerequisite material type for inventory capacity tests"
MT_NAME=$(unique_name "InvCapType")

xbe_json do material-types create --name "$MT_NAME"

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

test_name "Create secondary material type for delete tests"
MT_NAME2=$(unique_name "InvCapType2")

xbe_json do material-types create --name "$MT_NAME2"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID2" && "$CREATED_MATERIAL_TYPE_ID2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID2"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a second material type"
        run_tests
    fi
else
    fail "Failed to create secondary material type"
    echo "Cannot continue without a second material type"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create inventory capacity with required relationships"
xbe_json do inventory-capacities create \
    --material-site "$CREATED_MATERIAL_SITE_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --min-capacity-tons 50 \
    --max-capacity-tons 500 \
    --threshold-tons 75

if [[ $status -eq 0 ]]; then
    CREATED_INVENTORY_CAPACITY_ID=$(json_get ".id")
    if [[ -n "$CREATED_INVENTORY_CAPACITY_ID" && "$CREATED_INVENTORY_CAPACITY_ID" != "null" ]]; then
        register_cleanup "inventory-capacities" "$CREATED_INVENTORY_CAPACITY_ID"
        pass
    else
        fail "Created inventory capacity but no ID returned"
    fi
else
    fail "Failed to create inventory capacity"
fi

if [[ -z "$CREATED_INVENTORY_CAPACITY_ID" || "$CREATED_INVENTORY_CAPACITY_ID" == "null" ]]; then
    echo "Cannot continue without a valid inventory capacity ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update inventory capacity max-capacity-tons"
xbe_json do inventory-capacities update "$CREATED_INVENTORY_CAPACITY_ID" --max-capacity-tons 600
assert_success

test_name "Update inventory capacity min-capacity-tons"
xbe_json do inventory-capacities update "$CREATED_INVENTORY_CAPACITY_ID" --min-capacity-tons 40
assert_success

test_name "Update inventory capacity threshold-tons"
xbe_json do inventory-capacities update "$CREATED_INVENTORY_CAPACITY_ID" --threshold-tons 90
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show inventory capacity"
xbe_json view inventory-capacities show "$CREATED_INVENTORY_CAPACITY_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List inventory capacities"
xbe_json view inventory-capacities list
assert_success

test_name "List inventory capacities returns array"
xbe_json view inventory-capacities list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list inventory capacities"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List inventory capacities with --material-site filter"
xbe_json view inventory-capacities list --material-site "$CREATED_MATERIAL_SITE_ID"
assert_success

test_name "List inventory capacities with --material-type filter"
xbe_json view inventory-capacities list --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List inventory capacities with --limit"
xbe_json view inventory-capacities list --limit 5
assert_success

test_name "List inventory capacities with --offset"
xbe_json view inventory-capacities list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create inventory capacity without material type fails"
xbe_json do inventory-capacities create --material-site "$CREATED_MATERIAL_SITE_ID"
assert_failure

test_name "Update inventory capacity without fields fails"
xbe_json do inventory-capacities update "$CREATED_INVENTORY_CAPACITY_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete inventory capacity requires --confirm flag"
xbe_json do inventory-capacities create \
    --material-site "$CREATED_MATERIAL_SITE_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID2" \
    --max-capacity-tons 300

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    register_cleanup "inventory-capacities" "$DEL_ID"
    xbe_json do inventory-capacities delete "$DEL_ID"
    assert_failure
else
    fail "Failed to create inventory capacity for deletion test"
fi

test_name "Delete inventory capacity with --confirm"
if [[ -n "$DEL_ID" && "$DEL_ID" != "null" ]]; then
    xbe_json do inventory-capacities delete "$DEL_ID" --confirm
    assert_success
else
    skip "No inventory capacity ID available for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
