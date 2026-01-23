#!/bin/bash
#
# XBE CLI Integration Tests: Inventory Estimates
#
# Tests CRUD operations for the inventory-estimates resource.
# Inventory estimates track estimated inventory amounts for material sites and types.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID_1=""
CREATED_MATERIAL_SITE_ID_2=""
CREATED_MATERIAL_TYPE_ID_1=""
CREATED_MATERIAL_TYPE_ID_2=""
CREATED_ESTIMATE_ID_1=""
CREATED_ESTIMATE_ID_2=""
CURRENT_USER_ID=""
HAS_USER="false"

ESTIMATED_AT_1="2025-01-05T08:00:00Z"
ESTIMATED_AT_2="2025-01-06T08:00:00Z"
ESTIMATED_AT_UPDATED="2025-01-07T08:00:00Z"
AMOUNT_TONS_1="100"
AMOUNT_TONS_2="150"
AMOUNT_TONS_UPDATED="110"

DESCRIPTION_1="Initial estimate"
DESCRIPTION_2="Second estimate"
DESCRIPTION_UPDATED="Adjusted estimate"

describe "Resource: inventory-estimates"

# ============================================================================
# Prerequisites - Create broker, material supplier, material sites, material types
# ============================================================================

test_name "Create prerequisite broker for inventory estimate tests"
BROKER_NAME=$(unique_name "InvEstBroker")

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

test_name "Create prerequisite material supplier"
SUPPLIER_NAME=$(unique_name "InvEstSupplier")

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

test_name "Create prerequisite material site 1"
SITE_NAME_1=$(unique_name "InvEstSite1")

xbe_json do material-sites create \
    --name "$SITE_NAME_1" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Inventory Way, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID_1=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID_1" && "$CREATED_MATERIAL_SITE_ID_1" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID_1"
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

test_name "Create prerequisite material site 2"
SITE_NAME_2=$(unique_name "InvEstSite2")

xbe_json do material-sites create \
    --name "$SITE_NAME_2" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "101 Inventory Way, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID_2" && "$CREATED_MATERIAL_SITE_ID_2" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID_2"
        pass
    else
        fail "Created material site 2 but no ID returned"
        echo "Cannot continue without a material site"
        run_tests
    fi
else
    fail "Failed to create material site 2"
    echo "Cannot continue without a material site"
    run_tests
fi

test_name "Create prerequisite material type 1"
MT_NAME_1=$(unique_name "InvEstType1")

xbe_json do material-types create --name "$MT_NAME_1"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_1=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_1" && "$CREATED_MATERIAL_TYPE_ID_1" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_1"
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

test_name "Create prerequisite material type 2"
MT_NAME_2=$(unique_name "InvEstType2")

xbe_json do material-types create --name "$MT_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_TYPE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_TYPE_ID_2" && "$CREATED_MATERIAL_TYPE_ID_2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_MATERIAL_TYPE_ID_2"
        pass
    else
        fail "Created material type 2 but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    fail "Failed to create material type 2"
    echo "Cannot continue without a material type"
    run_tests
fi

test_name "Fetch current user id"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        HAS_USER="true"
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to fetch current user"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create inventory estimate with required fields"
xbe_json do inventory-estimates create \
    --material-site "$CREATED_MATERIAL_SITE_ID_1" \
    --material-type "$CREATED_MATERIAL_TYPE_ID_1" \
    --estimated-at "$ESTIMATED_AT_1" \
    --amount-tons "$AMOUNT_TONS_1" \
    --description "$DESCRIPTION_1"

if [[ $status -eq 0 ]]; then
    CREATED_ESTIMATE_ID_1=$(json_get ".id")
    if [[ -n "$CREATED_ESTIMATE_ID_1" && "$CREATED_ESTIMATE_ID_1" != "null" ]]; then
        register_cleanup "inventory-estimates" "$CREATED_ESTIMATE_ID_1"
        pass
    else
        fail "Created inventory estimate but no ID returned"
    fi
else
    fail "Failed to create inventory estimate"
fi

if [[ -z "$CREATED_ESTIMATE_ID_1" || "$CREATED_ESTIMATE_ID_1" == "null" ]]; then
    echo "Cannot continue without a valid inventory estimate ID"
    run_tests
fi

test_name "Create second inventory estimate"
xbe_json do inventory-estimates create \
    --material-site "$CREATED_MATERIAL_SITE_ID_2" \
    --material-type "$CREATED_MATERIAL_TYPE_ID_2" \
    --estimated-at "$ESTIMATED_AT_2" \
    --amount-tons "$AMOUNT_TONS_2" \
    --description "$DESCRIPTION_2"

if [[ $status -eq 0 ]]; then
    CREATED_ESTIMATE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_ESTIMATE_ID_2" && "$CREATED_ESTIMATE_ID_2" != "null" ]]; then
        register_cleanup "inventory-estimates" "$CREATED_ESTIMATE_ID_2"
        pass
    else
        fail "Created inventory estimate 2 but no ID returned"
    fi
else
    fail "Failed to create inventory estimate 2"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show inventory estimate"
xbe_json view inventory-estimates show "$CREATED_ESTIMATE_ID_1"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List inventory estimates"
xbe_json view inventory-estimates list --limit 5
assert_success

test_name "List inventory estimates returns array"
xbe_json view inventory-estimates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list inventory estimates"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List inventory estimates with --estimated-at"
xbe_json view inventory-estimates list --estimated-at "$ESTIMATED_AT_1"
assert_success

test_name "List inventory estimates with --amount-tons"
xbe_json view inventory-estimates list --amount-tons "$AMOUNT_TONS_1"
assert_success

test_name "List inventory estimates with --material-site"
xbe_json view inventory-estimates list --material-site "$CREATED_MATERIAL_SITE_ID_1"
assert_success

test_name "List inventory estimates with --material-type"
xbe_json view inventory-estimates list --material-type "$CREATED_MATERIAL_TYPE_ID_1"
assert_success

test_name "List inventory estimates with --material-supplier-id"
xbe_json view inventory-estimates list --material-supplier-id "$CREATED_MATERIAL_SUPPLIER_ID"
assert_success

test_name "List inventory estimates with --material-supplier"
xbe_json view inventory-estimates list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
assert_success

test_name "List inventory estimates with --broker-id"
xbe_json view inventory-estimates list --broker-id "$CREATED_BROKER_ID"
assert_success

test_name "List inventory estimates with --broker"
xbe_json view inventory-estimates list --broker "$CREATED_BROKER_ID"
assert_success

if [[ "$HAS_USER" == "true" ]]; then
    test_name "List inventory estimates with --created-by"
    xbe_json view inventory-estimates list --created-by "$CURRENT_USER_ID"
    assert_success
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update inventory estimate amount and description"
xbe_json do inventory-estimates update "$CREATED_ESTIMATE_ID_1" \
    --amount-tons "$AMOUNT_TONS_UPDATED" \
    --description "$DESCRIPTION_UPDATED"
assert_success

test_name "Update inventory estimate estimated-at"
xbe_json do inventory-estimates update "$CREATED_ESTIMATE_ID_1" \
    --estimated-at "$ESTIMATED_AT_UPDATED"
assert_success

test_name "Update inventory estimate material site and type"
xbe_json do inventory-estimates update "$CREATED_ESTIMATE_ID_1" \
    --material-site "$CREATED_MATERIAL_SITE_ID_2" \
    --material-type "$CREATED_MATERIAL_TYPE_ID_2"
assert_success

test_name "Update inventory estimate without fields fails"
xbe_json do inventory-estimates update "$CREATED_ESTIMATE_ID_1"
assert_failure

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete inventory estimate requires --confirm flag"
xbe_run do inventory-estimates delete "$CREATED_ESTIMATE_ID_2"
assert_failure

test_name "Delete inventory estimate with --confirm"
EST_DEL_EST_AT="2025-01-08T08:00:00Z"
EST_DEL_AMT="90"
xbe_json do inventory-estimates create \
    --material-site "$CREATED_MATERIAL_SITE_ID_1" \
    --material-type "$CREATED_MATERIAL_TYPE_ID_1" \
    --estimated-at "$EST_DEL_EST_AT" \
    --amount-tons "$EST_DEL_AMT"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do inventory-estimates delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create inventory estimate for deletion test"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create inventory estimate missing required flags fails"
xbe_run do inventory-estimates create --material-site "$CREATED_MATERIAL_SITE_ID_1"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
