#!/bin/bash
#
# XBE CLI Integration Tests: Material Type Material Site Inventory Locations
#
# Tests CRUD operations for the material_type_material_site_inventory_locations resource.
#
# COVERAGE: Full CRUD + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
CREATED_BROKER_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""
PARENT_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_TYPE_ID=""
UPDATED_MATERIAL_TYPE_ID=""
CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID=""
UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID=""

describe "Resource: material-type-material-site-inventory-locations"

cleanup_material_site_inventory_locations() {
    if [[ -z "$XBE_TOKEN" ]]; then
        return
    fi

    for location_id in "$UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID" "$CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID"; do
        if [[ -n "$location_id" && "$location_id" != "null" ]]; then
            curl -sS -X DELETE \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                "$XBE_BASE_URL/v1/material-site-inventory-locations/$location_id" \
                >/dev/null 2>&1 || true
        fi
    done
}

trap 'run_cleanup; cleanup_material_site_inventory_locations' EXIT

resolve_token_from_store() {
    local base_url="$1"
    local tmp_dir
    local tmp_file
    local token

    tmp_dir=$(mktemp -d "${PROJECT_ROOT}/xbe_token_tmp.XXXXXX")
    tmp_file="${tmp_dir}/token.go"
    cat >"$tmp_file" <<'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/xbe-inc/xbe-cli/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	token, _, err := auth.ResolveToken(os.Args[1], "")
	if err != nil || token == "" {
		os.Exit(1)
	}
	fmt.Print(token)
}
EOF

    token=$(cd "$PROJECT_ROOT" && go run "$tmp_file" "$base_url" 2>/dev/null)
    local status=$?
    rm -rf "$tmp_dir"

    if [[ $status -ne 0 ]]; then
        return 1
    fi

    printf "%s" "$token"
}

create_inventory_location() {
    local qualified_name="$1"
    local material_site_id="$2"
    local payload
    payload=$(jq -n \
        --arg qualified_name "$qualified_name" \
        --arg material_site_id "$material_site_id" \
        '{
            data: {
                type: "material-site-inventory-locations",
                attributes: {
                    "qualified-name": $qualified_name
                },
                relationships: {
                    "material-site": {
                        data: {
                            type: "material-sites",
                            id: $material_site_id
                        }
                    }
                }
            }
        }')

    local response
    if ! response=$(curl -sS -w "\n%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -H "Content-Type: application/vnd.api+json" \
        -d "$payload" \
        "$XBE_BASE_URL/v1/material-site-inventory-locations"); then
        return 1
    fi

    local body
    local status_code
    body=$(printf "%s" "$response" | sed '$d')
    status_code=$(printf "%s" "$response" | tail -n1)

    if [[ "$status_code" == 2* ]]; then
        printf "%s" "$body" | jq -r '.data.id'
        return 0
    fi

    echo "$body" >&2
    return 1
}

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for inventory location tests"
BROKER_NAME=$(unique_name "MTMSILBroker")

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
SUPPLIER_NAME=$(unique_name "MTMSILSupplier")

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

test_name "Create prerequisite material site"
MATERIAL_SITE_NAME=$(unique_name "MTMSILMaterialSite")

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

test_name "Create prerequisite parent material type"
PARENT_MATERIAL_TYPE_NAME=$(unique_name "MTMSILParentMaterialType")

xbe_json do material-types create --name "$PARENT_MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    PARENT_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$PARENT_MATERIAL_TYPE_ID" && "$PARENT_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$PARENT_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created parent material type but no ID returned"
        echo "Cannot continue without a parent material type"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_TYPE_ID" ]]; then
        PARENT_MATERIAL_TYPE_ID="$XBE_TEST_MATERIAL_TYPE_ID"
        echo "    Using XBE_TEST_MATERIAL_TYPE_ID: $PARENT_MATERIAL_TYPE_ID"
        pass
    else
        fail "Failed to create parent material type and XBE_TEST_MATERIAL_TYPE_ID not set"
        echo "Cannot continue without a parent material type"
        run_tests
    fi
fi

test_name "Create prerequisite supplier-specific material type"
MATERIAL_TYPE_NAME=$(unique_name "MTMSILMaterialType")

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

test_name "Create second material type for update"
UPDATED_MATERIAL_TYPE_NAME=$(unique_name "MTMSILMaterialTypeUpdated")

xbe_json do material-types create \
    --name "$UPDATED_MATERIAL_TYPE_NAME" \
    --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID" \
    --parent-material-type "$PARENT_MATERIAL_TYPE_ID"

if [[ $status -eq 0 ]]; then
    UPDATED_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$UPDATED_MATERIAL_TYPE_ID" && "$UPDATED_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$UPDATED_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a second material type"
        run_tests
    fi
else
    fail "Failed to create second material type"
    run_tests
fi

if [[ -z "$XBE_TOKEN" && -n "$XBE_USE_STORED_AUTH" ]]; then
    XBE_TOKEN=$(resolve_token_from_store "$XBE_BASE_URL") || true
fi

if [[ -z "$XBE_TOKEN" ]]; then
    fail "XBE_TOKEN is required to create material site inventory locations"
    run_tests
fi

test_name "Create material site inventory location"
QUALIFIED_NAME=$(unique_name "MTMSILInventoryLocation")
CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID=$(create_inventory_location "$QUALIFIED_NAME" "$CREATED_MATERIAL_SITE_ID")
if [[ -n "$CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID" && "$CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID" != "null" ]]; then
    pass
else
    fail "Failed to create material site inventory location"
    run_tests
fi

test_name "Create second material site inventory location for update"
QUALIFIED_NAME_UPDATED=$(unique_name "MTMSILInventoryLocationUpdated")
UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID=$(create_inventory_location "$QUALIFIED_NAME_UPDATED" "$CREATED_MATERIAL_SITE_ID")
if [[ -n "$UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID" && "$UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID" != "null" ]]; then
    pass
else
    fail "Failed to create second material site inventory location"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material type material site inventory location"
xbe_json do material-type-material-site-inventory-locations create \
    --material-type "$CREATED_MATERIAL_TYPE_ID" \
    --material-site-inventory-location "$CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "material-type-material-site-inventory-locations" "$CREATED_ID"
        pass
    else
        fail "Created mapping but no ID returned"
        run_tests
    fi
else
    fail "Failed to create mapping: $output"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update material type relationship"
xbe_json do material-type-material-site-inventory-locations update "$CREATED_ID" \
    --material-type "$UPDATED_MATERIAL_TYPE_ID"
assert_success

test_name "Update inventory location relationship"
xbe_json do material-type-material-site-inventory-locations update "$CREATED_ID" \
    --material-site-inventory-location "$UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material type material site inventory location"
xbe_json view material-type-material-site-inventory-locations show "$CREATED_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show material type material site inventory location"
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List material type material site inventory locations"
xbe_json view material-type-material-site-inventory-locations list
assert_success

test_name "List material type material site inventory locations with --material-type filter"
xbe_json view material-type-material-site-inventory-locations list --material-type "$UPDATED_MATERIAL_TYPE_ID"
assert_success

test_name "List material type material site inventory locations with --material-site-inventory-location filter"
xbe_json view material-type-material-site-inventory-locations list --material-site-inventory-location "$UPDATED_MATERIAL_SITE_INVENTORY_LOCATION_ID"
assert_success

test_name "List material type material site inventory locations with --limit"
xbe_json view material-type-material-site-inventory-locations list --limit 5
assert_success

test_name "List material type material site inventory locations with --offset"
xbe_json view material-type-material-site-inventory-locations list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create requires --material-type"
xbe_run do material-type-material-site-inventory-locations create \
    --material-site-inventory-location "$CREATED_MATERIAL_SITE_INVENTORY_LOCATION_ID"
assert_failure

test_name "Create requires --material-site-inventory-location"
xbe_run do material-type-material-site-inventory-locations create \
    --material-type "$CREATED_MATERIAL_TYPE_ID"
assert_failure

test_name "Update without fields fails"
xbe_run do material-type-material-site-inventory-locations update "$CREATED_ID"
assert_failure

test_name "Update with empty material type fails"
xbe_run do material-type-material-site-inventory-locations update "$CREATED_ID" --material-type ""
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
xbe_run do material-type-material-site-inventory-locations delete "$CREATED_ID"
assert_failure

test_name "Delete material type material site inventory location with --confirm"
xbe_json do material-type-material-site-inventory-locations delete "$CREATED_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
