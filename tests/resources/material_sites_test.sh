#!/bin/bash
#
# XBE CLI Integration Tests: Material Sites
#
# Tests CRUD operations for the material_sites resource.
# Material sites are locations like plants, quarries, stockpiles.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_BROKER_ID=""

describe "Resource: material_sites"

# ============================================================================
# Prerequisites - Create broker and material supplier for tests
# ============================================================================

test_name "Create prerequisite broker for material site tests"
BROKER_NAME=$(unique_name "MSTestBroker")

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

test_name "Create prerequisite material supplier for tests"
SUPPLIER_NAME=$(unique_name "MSTestSupplier")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material site with required fields"
TEST_NAME=$(unique_name "MaterialSite")

xbe_json do material-sites create \
    --name "$TEST_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Test Quarry Rd, Chicago, IL 60601"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
        pass
    else
        fail "Created material site but no ID returned"
    fi
else
    fail "Failed to create material site"
fi

# Only continue if we successfully created a material site
if [[ -z "$CREATED_MATERIAL_SITE_ID" || "$CREATED_MATERIAL_SITE_ID" == "null" ]]; then
    echo "Cannot continue without a valid material site ID"
    run_tests
fi

test_name "Create material site with phone-number"
TEST_NAME2=$(unique_name "MaterialSite2")
TEST_PHONE=$(unique_mobile)
xbe_json do material-sites create \
    --name "$TEST_NAME2" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "101 Test Quarry Rd, Chicago, IL 60601" \
    --phone-number "$TEST_PHONE"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with phone-number"
fi

test_name "Create material site with cb-channel"
TEST_NAME3=$(unique_name "MaterialSite3")
xbe_json do material-sites create \
    --name "$TEST_NAME3" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "102 Test Quarry Rd, Chicago, IL 60601" \
    --cb-channel "Channel 9"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with cb-channel"
fi

test_name "Create material site with hours-description"
TEST_NAME4=$(unique_name "MaterialSite4")
xbe_json do material-sites create \
    --name "$TEST_NAME4" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "103 Test Quarry Rd, Chicago, IL 60601" \
    --hours-description "Mon-Fri 6am-6pm"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with hours-description"
fi

test_name "Create material site with notes"
TEST_NAME5=$(unique_name "MaterialSite5")
xbe_json do material-sites create \
    --name "$TEST_NAME5" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "104 Test Quarry Rd, Chicago, IL 60601" \
    --notes "Test notes for material site"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with notes"
fi

test_name "Create material site with color-hex"
TEST_NAME6=$(unique_name "MaterialSite6")
xbe_json do material-sites create \
    --name "$TEST_NAME6" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "105 Test Quarry Rd, Chicago, IL 60601" \
    --color-hex "#FF5500"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with color-hex"
fi

test_name "Create material site with operating-status inactive"
TEST_NAME7=$(unique_name "MaterialSite7")
xbe_json do material-sites create \
    --name "$TEST_NAME7" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "106 Test Quarry Rd, Chicago, IL 60601" \
    --operating-status "inactive"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with operating-status"
fi

test_name "Create material site with is-ticket-maker"
TEST_NAME8=$(unique_name "MaterialSite8")
xbe_json do material-sites create \
    --name "$TEST_NAME8" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "107 Test Quarry Rd, Chicago, IL 60601" \
    --is-ticket-maker
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with is-ticket-maker"
fi

test_name "Create material site with has-scale"
TEST_NAME9=$(unique_name "MaterialSite9")
xbe_json do material-sites create \
    --name "$TEST_NAME9" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "108 Test Quarry Rd, Chicago, IL 60601" \
    --has-scale
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with has-scale"
fi

test_name "Create material site with address"
TEST_NAME10=$(unique_name "MaterialSite10")
xbe_json do material-sites create \
    --name "$TEST_NAME10" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "100 Quarry Road, Springfield, IL 62701"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with address"
fi

test_name "Create material site with coordinates and skip-geocoding"
TEST_NAME11=$(unique_name "MaterialSite11")
xbe_json do material-sites create \
    --name "$TEST_NAME11" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "200 Manual St" \
    --address-latitude "41.8781" \
    --address-longitude "-87.6298" \
    --skip-geocoding
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "material-sites" "$id"
    pass
else
    fail "Failed to create material site with coordinates"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material site name"
UPDATED_NAME=$(unique_name "UpdatedMS")
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update material site phone-number"
UPDATED_PHONE=$(unique_mobile)
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --phone-number "$UPDATED_PHONE"
assert_success

test_name "Update material site cb-channel"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --cb-channel "Channel 19"
assert_success

test_name "Update material site hours-description"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --hours-description "24/7 Operations"
assert_success

test_name "Update material site notes"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --notes "Updated notes"
assert_success

test_name "Update material site color-hex"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --color-hex "#00FF00"
assert_success

test_name "Update material site operating-status"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --operating-status "inactive"
assert_success

test_name "Update material site is-ticket-maker to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --is-ticket-maker
assert_success

test_name "Update material site is-ticket-maker to false"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --no-is-ticket-maker
assert_success

test_name "Update material site has-scale to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --has-scale
assert_success

test_name "Update material site has-scale to false"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --no-has-scale
assert_success

test_name "Update material site can-be-job-material-site to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --can-be-job-material-site
assert_success

test_name "Update material site can-be-start-site to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --can-be-start-site
assert_success

test_name "Update material site is-portable to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --is-portable
assert_success

test_name "Update material site is-productive to true"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --is-productive
assert_success

test_name "Update material site address"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID" --address "300 Updated Rd, Chicago, IL 60601"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material sites"
xbe_json view material-sites list --limit 5
assert_success

test_name "List material sites returns array"
xbe_json view material-sites list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material sites"
fi

# ============================================================================
# LIST Tests - All Filters
# ============================================================================

test_name "List material sites with --name filter"
xbe_json view material-sites list --name "$UPDATED_NAME" --limit 10
assert_success

test_name "List material sites with --active filter"
xbe_json view material-sites list --active --limit 10
assert_success

test_name "List material sites with --broker filter"
xbe_json view material-sites list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List material sites with --q filter"
xbe_json view material-sites list --q "MaterialSite" --limit 10
assert_success

test_name "List material sites with --material-supplier filter"
xbe_json view material-sites list --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --limit 10
assert_success

test_name "List material sites with --operating-status filter"
xbe_json view material-sites list --operating-status "active" --limit 10
assert_success

test_name "List material sites with --is-ticket-maker filter"
xbe_json view material-sites list --is-ticket-maker true --limit 10
assert_success

test_name "List material sites with --is-broker-active filter"
xbe_json view material-sites list --is-broker-active true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List material sites with --limit"
xbe_json view material-sites list --limit 3
assert_success

test_name "List material sites with --offset"
xbe_json view material-sites list --limit 3 --offset 3
assert_success

test_name "List material sites with pagination (limit + offset)"
xbe_json view material-sites list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete material site requires --confirm flag"
xbe_run do material-sites delete "$CREATED_MATERIAL_SITE_ID"
assert_failure

test_name "Delete material site with --confirm"
# Create a material site specifically for deletion
TEST_DEL_NAME=$(unique_name "DeleteMeMS")
xbe_json do material-sites create \
    --name "$TEST_DEL_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "999 Delete Rd, Chicago, IL 60601"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do material-sites delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create material site for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material site without name fails"
xbe_json do material-sites create --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" --address "123 Test Rd"
assert_failure

test_name "Create material site without material-supplier fails"
xbe_json do material-sites create --name "Test Material Site" --address "123 Test Rd"
assert_failure

test_name "Create material site without address fails"
xbe_json do material-sites create --name "Test Material Site" --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do material-sites update "$CREATED_MATERIAL_SITE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
