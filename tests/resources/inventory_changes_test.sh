#!/bin/bash
#
# XBE CLI Integration Tests: Inventory Changes
#
# Tests list, show, create, and delete operations for the inventory-changes resource.
#
# COVERAGE: List filters + show + create/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
MATERIAL_SITE_ID=""
MATERIAL_SUPPLIER_ID=""
MATERIAL_TYPE_ID=""
FORECAST_START_AT=""
SAMPLE_ID=""

describe "Resource: inventory-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List inventory changes"
xbe_json view inventory-changes list --limit 5
assert_success

test_name "List inventory changes returns array"
xbe_json view inventory-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list inventory changes"
fi

# ============================================================================
# Sample Record (used for show fallback)
# ============================================================================

test_name "Capture sample inventory change"
xbe_json view inventory-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No inventory changes available for sample"
    fi
else
    skip "Could not list inventory changes to capture sample"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find material site with supplier"
xbe_json view material-sites list --limit 50
if [[ $status -eq 0 ]]; then
    MATERIAL_SITE_ID=$(json_get "map(select(.supplier_id != null and .supplier_id != \"\")) | .[0].id")
    MATERIAL_SUPPLIER_ID=$(json_get "map(select(.supplier_id != null and .supplier_id != \"\")) | .[0].supplier_id")
    if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_SUPPLIER_ID" && "$MATERIAL_SUPPLIER_ID" != "null" ]]; then
        pass
    else
        skip "No material site with supplier available"
    fi
else
    skip "Could not list material sites"
fi

test_name "Find material type for supplier"
if [[ -n "$MATERIAL_SUPPLIER_ID" && "$MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view material-types list --material-supplier "$MATERIAL_SUPPLIER_ID" --limit 50
    if [[ $status -eq 0 ]]; then
        MATERIAL_TYPE_ID=$(json_get ".[0].id")
        if [[ -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
            pass
        else
            skip "No material types available for supplier"
        fi
    else
        skip "Could not list material types for supplier"
    fi
else
    skip "No material supplier available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create inventory change with required fields"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    ESTIMATE_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    FORECAST_START_AT="2000-01-01T00:00:00Z"
    xbe_json do inventory-changes create \
        --material-site "$MATERIAL_SITE_ID" \
        --material-type "$MATERIAL_TYPE_ID" \
        --estimate-at "$ESTIMATE_AT" \
        --forecast-start-at "$FORECAST_START_AT"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "inventory-changes" "$CREATED_ID"
            FORECAST_START_AT=$(json_get ".forecast_start_at")
            pass
        else
            fail "Created inventory change but no ID returned"
        fi
    else
        fail "Failed to create inventory change: $output"
    fi
else
    skip "Missing material site or material type for create"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show inventory change"
SHOW_ID="$CREATED_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$SAMPLE_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view inventory-changes show "$SHOW_ID"
    assert_success
else
    skip "No inventory change ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List inventory changes with --material-site filter"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" ]]; then
    xbe_json view inventory-changes list --material-site "$MATERIAL_SITE_ID" --limit 5
    assert_success
else
    skip "No material site ID available"
fi

test_name "List inventory changes with --material-type filter"
if [[ -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view inventory-changes list --material-type "$MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "List inventory changes with --material-supplier filter"
if [[ -n "$MATERIAL_SUPPLIER_ID" && "$MATERIAL_SUPPLIER_ID" != "null" ]]; then
    xbe_json view inventory-changes list --material-supplier "$MATERIAL_SUPPLIER_ID" --limit 5
    assert_success
else
    skip "No material supplier ID available"
fi

test_name "List inventory changes with --forecast-start-at filter"
if [[ -n "$FORECAST_START_AT" && "$FORECAST_START_AT" != "null" ]]; then
    xbe_json view inventory-changes list --forecast-start-at "$FORECAST_START_AT" --limit 5
    assert_success
else
    skip "No forecast-start-at value available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete inventory change requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do inventory-changes delete "$CREATED_ID"
    assert_failure
else
    skip "No created inventory change ID available"
fi

test_name "Delete inventory change with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do inventory-changes delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created inventory change ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create inventory change without estimate-at fails"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json do inventory-changes create --material-site "$MATERIAL_SITE_ID" --material-type "$MATERIAL_TYPE_ID"
    assert_failure
else
    skip "Missing material site or material type for error test"
fi

test_name "Create inventory change with forecast-start-at after estimate-at fails"
if [[ -n "$MATERIAL_SITE_ID" && "$MATERIAL_SITE_ID" != "null" && -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
    ESTIMATE_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    LATE_FORECAST="2999-01-01T00:00:00Z"
    xbe_json do inventory-changes create \
        --material-site "$MATERIAL_SITE_ID" \
        --material-type "$MATERIAL_TYPE_ID" \
        --estimate-at "$ESTIMATE_AT" \
        --forecast-start-at "$LATE_FORECAST"
    assert_failure
else
    skip "Missing material site or material type for error test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
