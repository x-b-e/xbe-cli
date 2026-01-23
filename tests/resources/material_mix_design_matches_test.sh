#!/bin/bash
#
# XBE CLI Integration Tests: Material Mix Design Matches
#
# Tests create operations for the material-mix-design-matches resource.
# Requires a material type with a material supplier relationship.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_TYPE_ID=""
PARENT_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_SITE_ID=""
CREATED_MATERIAL_SITE_ID_2=""

AS_OF_TIMESTAMP="2026-01-23T00:00:00Z"

describe "Resource: material-mix-design-matches"

# ============================================================================
# Prerequisites - Create broker, material supplier, material type, material sites
# ============================================================================

test_name "Create prerequisite broker for material mix design match tests"
BROKER_NAME=$(unique_name "MixDesignMatchBroker")

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
SUPPLIER_NAME=$(unique_name "MixDesignMatchSupplier")

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

test_name "Create prerequisite base material type"
BASE_MATERIAL_TYPE_NAME=$(unique_name "MixDesignMatchBaseMaterialType")

xbe_json do material-types create --name "$BASE_MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    PARENT_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$PARENT_MATERIAL_TYPE_ID" && "$PARENT_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$PARENT_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created base material type but no ID returned"
        echo "Cannot continue without a base material type"
        run_tests
    fi
else
    fail "Failed to create base material type"
    echo "Cannot continue without a base material type"
    run_tests
fi

test_name "Create prerequisite material type with supplier"
MATERIAL_TYPE_NAME=$(unique_name "MixDesignMatchMaterialType")

xbe_json do material-types create \
    --name "$MATERIAL_TYPE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --parent-material-type "$PARENT_MATERIAL_TYPE_ID"

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

test_name "Create prerequisite material sites"
SITE_NAME_1=$(unique_name "MixDesignMatchSiteA")
SITE_NAME_2=$(unique_name "MixDesignMatchSiteB")

xbe_json do material-sites create \
    --name "$SITE_NAME_1" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "123 Mix Design Way, Springfield, IL"

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

xbe_json do material-sites create \
    --name "$SITE_NAME_2" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --address "456 Mix Design Way, Springfield, IL"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SITE_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SITE_ID_2" && "$CREATED_MATERIAL_SITE_ID_2" != "null" ]]; then
        register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID_2"
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material mix design match with required fields"
xbe_json do material-mix-design-matches create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --as-of "$AS_OF_TIMESTAMP"

if [[ $status -eq 0 ]]; then
    MATCH_ID=$(json_get ".id")
    if [[ -n "$MATCH_ID" && "$MATCH_ID" != "null" ]]; then
        pass
    else
        fail "Created material mix design match but no ID returned"
    fi
else
    fail "Failed to create material mix design match"
fi

test_name "Create material mix design match with material sites"
xbe_json do material-mix-design-matches create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --as-of "$AS_OF_TIMESTAMP" \
    --material-sites "$CREATED_MATERIAL_SITE_ID,$CREATED_MATERIAL_SITE_ID_2"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create material mix design match with material sites"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material mix design match without material type fails"
xbe_json do material-mix-design-matches create --as-of "$AS_OF_TIMESTAMP"
assert_failure

test_name "Create material mix design match without as-of fails"
xbe_json do material-mix-design-matches create --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
