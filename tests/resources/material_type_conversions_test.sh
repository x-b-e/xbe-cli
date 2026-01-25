#!/bin/bash
#
# XBE CLI Integration Tests: Material Type Conversions
#
# Tests create/update/delete operations and list filters for the
# material_type_conversions resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_LOCAL_SUPPLIER_ID=""
CREATED_FOREIGN_SUPPLIER_ID=""
CREATED_FOREIGN_SUPPLIER_ID_2=""
CREATED_LOCAL_TYPE_ID=""
CREATED_LOCAL_TYPE_ID_2=""
CREATED_FOREIGN_TYPE_ID=""
CREATED_FOREIGN_TYPE_ID_2=""
CREATED_LOCAL_SITE_ID=""
CREATED_LOCAL_SITE_ID_2=""
CREATED_FOREIGN_SITE_ID=""
CREATED_FOREIGN_SITE_ID_2=""
CREATED_CONVERSION_ID=""

describe "Resource: material-type-conversions"

# ==========================================================================
# Prerequisites - Create broker, suppliers, types, and sites
# ==========================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "MTConvBroker")

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
    fail "Failed to create broker"
    echo "Cannot continue without a broker"
    run_tests
fi

test_name "Create local material supplier"
LOCAL_SUPPLIER_NAME=$(unique_name "MTConvLocalSupplier")

xbe_json do material-suppliers create --name "$LOCAL_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_LOCAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCAL_SUPPLIER_ID" && "$CREATED_LOCAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_LOCAL_SUPPLIER_ID"
        pass
    else
        fail "Created material supplier but no ID returned"
        echo "Cannot continue without a local material supplier"
        run_tests
    fi
else
    fail "Failed to create local material supplier"
    echo "Cannot continue without a local material supplier"
    run_tests
fi

test_name "Create foreign material supplier"
FOREIGN_SUPPLIER_NAME=$(unique_name "MTConvForeignSupplier")

xbe_json do material-suppliers create --name "$FOREIGN_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_SUPPLIER_ID" && "$CREATED_FOREIGN_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_FOREIGN_SUPPLIER_ID"
        pass
    else
        fail "Created foreign material supplier but no ID returned"
        echo "Cannot continue without a foreign material supplier"
        run_tests
    fi
else
    fail "Failed to create foreign material supplier"
    echo "Cannot continue without a foreign material supplier"
    run_tests
fi

test_name "Create alternate foreign material supplier"
FOREIGN_SUPPLIER_NAME_2=$(unique_name "MTConvForeignSupplier2")

xbe_json do material-suppliers create --name "$FOREIGN_SUPPLIER_NAME_2" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_SUPPLIER_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_SUPPLIER_ID_2" && "$CREATED_FOREIGN_SUPPLIER_ID_2" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_FOREIGN_SUPPLIER_ID_2"
        pass
    else
        fail "Created alternate foreign material supplier but no ID returned"
        echo "Cannot continue without a foreign material supplier"
        run_tests
    fi
else
    fail "Failed to create alternate foreign material supplier"
    echo "Cannot continue without a foreign material supplier"
    run_tests
fi

test_name "Create local material type"
LOCAL_TYPE_NAME=$(unique_name "MTConvLocalType")

xbe_json do material-types create --name "$LOCAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_LOCAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCAL_TYPE_ID" && "$CREATED_LOCAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_LOCAL_TYPE_ID"
        pass
    else
        fail "Created local material type but no ID returned"
        echo "Cannot continue without a local material type"
        run_tests
    fi
else
    fail "Failed to create local material type"
    echo "Cannot continue without a local material type"
    run_tests
fi

test_name "Create alternate local material type"
LOCAL_TYPE_NAME_2=$(unique_name "MTConvLocalType2")

xbe_json do material-types create --name "$LOCAL_TYPE_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_LOCAL_TYPE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_LOCAL_TYPE_ID_2" && "$CREATED_LOCAL_TYPE_ID_2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_LOCAL_TYPE_ID_2"
        pass
    else
        fail "Created alternate local material type but no ID returned"
        echo "Cannot continue without a local material type"
        run_tests
    fi
else
    fail "Failed to create alternate local material type"
    echo "Cannot continue without a local material type"
    run_tests
fi

test_name "Create foreign material type"
FOREIGN_TYPE_NAME=$(unique_name "MTConvForeignType")

xbe_json do material-types create --name "$FOREIGN_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_TYPE_ID" && "$CREATED_FOREIGN_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_FOREIGN_TYPE_ID"
        pass
    else
        fail "Created foreign material type but no ID returned"
        echo "Cannot continue without a foreign material type"
        run_tests
    fi
else
    fail "Failed to create foreign material type"
    echo "Cannot continue without a foreign material type"
    run_tests
fi

test_name "Create alternate foreign material type"
FOREIGN_TYPE_NAME_2=$(unique_name "MTConvForeignType2")

xbe_json do material-types create --name "$FOREIGN_TYPE_NAME_2"

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_TYPE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_TYPE_ID_2" && "$CREATED_FOREIGN_TYPE_ID_2" != "null" ]]; then
        register_cleanup "material-types" "$CREATED_FOREIGN_TYPE_ID_2"
        pass
    else
        fail "Created alternate foreign material type but no ID returned"
        echo "Cannot continue without a foreign material type"
        run_tests
    fi
else
    fail "Failed to create alternate foreign material type"
    echo "Cannot continue without a foreign material type"
    run_tests
fi

test_name "Create local material site"
LOCAL_SITE_NAME=$(unique_name "MTConvLocalSite")
LOCAL_SITE_ADDRESS="100 Test Quarry Rd, Chicago, IL 60601"

xbe_json do material-sites create --name "$LOCAL_SITE_NAME" --material-supplier "$CREATED_LOCAL_SUPPLIER_ID" --address "$LOCAL_SITE_ADDRESS" --skip-geocoding

if [[ $status -eq 0 ]]; then
    CREATED_LOCAL_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_LOCAL_SITE_ID" && "$CREATED_LOCAL_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_LOCAL_SITE_ID"
        pass
    else
        fail "Created local material site but no ID returned"
        echo "Cannot continue without a local material site"
        run_tests
    fi
else
    fail "Failed to create local material site"
    echo "Cannot continue without a local material site"
    run_tests
fi

test_name "Create alternate local material site"
LOCAL_SITE_NAME_2=$(unique_name "MTConvLocalSite2")
LOCAL_SITE_ADDRESS_2="101 Test Quarry Rd, Chicago, IL 60601"

xbe_json do material-sites create --name "$LOCAL_SITE_NAME_2" --material-supplier "$CREATED_LOCAL_SUPPLIER_ID" --address "$LOCAL_SITE_ADDRESS_2" --skip-geocoding

if [[ $status -eq 0 ]]; then
    CREATED_LOCAL_SITE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_LOCAL_SITE_ID_2" && "$CREATED_LOCAL_SITE_ID_2" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_LOCAL_SITE_ID_2"
        pass
    else
        fail "Created alternate local material site but no ID returned"
        echo "Cannot continue without a local material site"
        run_tests
    fi
else
    fail "Failed to create alternate local material site"
    echo "Cannot continue without a local material site"
    run_tests
fi

test_name "Create foreign material site"
FOREIGN_SITE_NAME=$(unique_name "MTConvForeignSite")
FOREIGN_SITE_ADDRESS="102 Test Quarry Rd, Chicago, IL 60601"

xbe_json do material-sites create --name "$FOREIGN_SITE_NAME" --material-supplier "$CREATED_FOREIGN_SUPPLIER_ID" --address "$FOREIGN_SITE_ADDRESS" --skip-geocoding

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_SITE_ID=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_SITE_ID" && "$CREATED_FOREIGN_SITE_ID" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_FOREIGN_SITE_ID"
        pass
    else
        fail "Created foreign material site but no ID returned"
        echo "Cannot continue without a foreign material site"
        run_tests
    fi
else
    fail "Failed to create foreign material site"
    echo "Cannot continue without a foreign material site"
    run_tests
fi

test_name "Create alternate foreign material site"
FOREIGN_SITE_NAME_2=$(unique_name "MTConvForeignSite2")
FOREIGN_SITE_ADDRESS_2="103 Test Quarry Rd, Chicago, IL 60601"

xbe_json do material-sites create --name "$FOREIGN_SITE_NAME_2" --material-supplier "$CREATED_FOREIGN_SUPPLIER_ID_2" --address "$FOREIGN_SITE_ADDRESS_2" --skip-geocoding

if [[ $status -eq 0 ]]; then
    CREATED_FOREIGN_SITE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_FOREIGN_SITE_ID_2" && "$CREATED_FOREIGN_SITE_ID_2" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_FOREIGN_SITE_ID_2"
        pass
    else
        fail "Created alternate foreign material site but no ID returned"
        echo "Cannot continue without a foreign material site"
        run_tests
    fi
else
    fail "Failed to create alternate foreign material site"
    echo "Cannot continue without a foreign material site"
    run_tests
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create material type conversion"
xbe_json do material-type-conversions create \
    --material-supplier "$CREATED_LOCAL_SUPPLIER_ID" \
    --material-site "$CREATED_LOCAL_SITE_ID" \
    --material-type "$CREATED_LOCAL_TYPE_ID" \
    --foreign-material-supplier "$CREATED_FOREIGN_SUPPLIER_ID" \
    --foreign-material-site "$CREATED_FOREIGN_SITE_ID" \
    --foreign-material-type "$CREATED_FOREIGN_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CONVERSION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CONVERSION_ID" && "$CREATED_CONVERSION_ID" != "null" ]]; then
        register_cleanup "material-type-conversions" "$CREATED_CONVERSION_ID"
        pass
    else
        fail "Created material type conversion but no ID returned"
        echo "Cannot continue without a material type conversion"
        run_tests
    fi
else
    fail "Failed to create material type conversion"
    echo "Cannot continue without a material type conversion"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show material type conversion"
xbe_json view material-type-conversions show "$CREATED_CONVERSION_ID"
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update material type conversion relationships"
xbe_json do material-type-conversions update "$CREATED_CONVERSION_ID" \
    --material-type "$CREATED_LOCAL_TYPE_ID_2" \
    --material-site "$CREATED_LOCAL_SITE_ID_2" \
    --foreign-material-supplier "$CREATED_FOREIGN_SUPPLIER_ID_2" \
    --foreign-material-site "$CREATED_FOREIGN_SITE_ID_2" \
    --foreign-material-type "$CREATED_FOREIGN_TYPE_ID_2"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List material type conversions"
xbe_json view material-type-conversions list --limit 10
assert_success

test_name "List material type conversions returns array"
xbe_json view material-type-conversions list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material type conversions"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List material type conversions with --material-supplier filter"
xbe_json view material-type-conversions list --material-supplier "$CREATED_LOCAL_SUPPLIER_ID" --limit 10
assert_success

test_name "List material type conversions with --material-site filter"
xbe_json view material-type-conversions list --material-site "$CREATED_LOCAL_SITE_ID_2" --limit 10
assert_success

test_name "List material type conversions with --material-type filter"
xbe_json view material-type-conversions list --material-type "$CREATED_LOCAL_TYPE_ID_2" --limit 10
assert_success

test_name "List material type conversions with --foreign-material-supplier filter"
xbe_json view material-type-conversions list --foreign-material-supplier "$CREATED_FOREIGN_SUPPLIER_ID_2" --limit 10
assert_success

test_name "List material type conversions with --foreign-material-site filter"
xbe_json view material-type-conversions list --foreign-material-site "$CREATED_FOREIGN_SITE_ID_2" --limit 10
assert_success

test_name "List material type conversions with --foreign-material-type filter"
xbe_json view material-type-conversions list --foreign-material-type "$CREATED_FOREIGN_TYPE_ID_2" --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete material type conversion requires --confirm flag"
xbe_json do material-type-conversions delete "$CREATED_CONVERSION_ID"
assert_failure

test_name "Delete material type conversion with --confirm"
xbe_json do material-type-conversions delete "$CREATED_CONVERSION_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
