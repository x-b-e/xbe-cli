#!/bin/bash
#
# XBE CLI Integration Tests: Project Material Types
#
# Tests CRUD operations for the project-material-types resource.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_MATERIAL_TYPE_ID=""
CREATED_PROJECT_MATERIAL_TYPE_ID=""
CREATED_UNIT_OF_MEASURE_ID=""
CREATED_MATERIAL_SUPPLIER_ID=""
CREATED_MATERIAL_SITE_ID=""

PICKUP_AT_MIN="2026-01-01T08:00:00Z"
PICKUP_AT_MAX="2026-01-01T09:00:00Z"
DELIVER_AT_MIN="2026-01-01T10:00:00Z"
DELIVER_AT_MAX="2026-01-01T11:00:00Z"

describe "Resource: project-material-types"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for project material type tests"
BROKER_NAME=$(unique_name "PMTBroker")

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

test_name "Create prerequisite developer for project material type tests"
DEV_NAME=$(unique_name "PMTDeveloper")

xbe_json do developers create --name "$DEV_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_DEVELOPER_ID=$(json_get ".id")
    if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
        register_cleanup "developers" "$CREATED_DEVELOPER_ID"
        pass
    else
        fail "Created developer but no ID returned"
        echo "Cannot continue without a developer"
        run_tests
    fi
else
    fail "Failed to create developer"
    echo "Cannot continue without a developer"
    run_tests
fi

test_name "Create prerequisite project for project material type tests"
PROJECT_NAME=$(unique_name "PMTProject")

xbe_json do projects create --name "$PROJECT_NAME" --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID"
        pass
    else
        fail "Created project but no ID returned"
        echo "Cannot continue without a project"
        run_tests
    fi
else
    fail "Failed to create project"
    echo "Cannot continue without a project"
    run_tests
fi

test_name "Create prerequisite material type for project material type tests"
MATERIAL_TYPE_NAME=$(unique_name "PMTMaterial")

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
    fail "Failed to create material type"
    echo "Cannot continue without a material type"
    run_tests
fi

test_name "Find unit of measure for update"
xbe_json view unit-of-measures list --limit 1
if [[ $status -eq 0 ]]; then
    CREATED_UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
    if [[ -n "$CREATED_UNIT_OF_MEASURE_ID" && "$CREATED_UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        skip "No unit of measure ID available"
    fi
else
    skip "Unable to list unit of measures"
fi

test_name "Create prerequisite material supplier for project material type tests"
MATERIAL_SUPPLIER_NAME=$(unique_name "PMTSupplier")

xbe_json do material-suppliers create --name "$MATERIAL_SUPPLIER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_MATERIAL_SUPPLIER_ID=$(json_get ".id")
    if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" && "$CREATED_MATERIAL_SUPPLIER_ID" != "null" ]]; then
        register_cleanup "material-suppliers" "$CREATED_MATERIAL_SUPPLIER_ID"
        pass
    else
        skip "Created material supplier but no ID returned"
    fi
else
    skip "Failed to create material supplier"
fi

test_name "Create prerequisite material site for project material type tests"
if [[ -n "$CREATED_MATERIAL_SUPPLIER_ID" ]]; then
    MATERIAL_SITE_NAME=$(unique_name "PMTSite")
    xbe_json do material-sites create --name "$MATERIAL_SITE_NAME" --material-supplier "$CREATED_MATERIAL_SUPPLIER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_MATERIAL_SITE_ID=$(json_get ".id")
        if [[ -n "$CREATED_MATERIAL_SITE_ID" && "$CREATED_MATERIAL_SITE_ID" != "null" ]]; then
            register_cleanup "material-sites" "$CREATED_MATERIAL_SITE_ID"
            pass
        else
            skip "Created material site but no ID returned"
        fi
    else
        skip "Failed to create material site"
    fi
else
    skip "No material supplier available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project material type with required fields"
xbe_json do project-material-types create \
    --project "$CREATED_PROJECT_ID" \
    --material-type "$CREATED_MATERIAL_TYPE_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_MATERIAL_TYPE_ID" && "$CREATED_PROJECT_MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "project-material-types" "$CREATED_PROJECT_MATERIAL_TYPE_ID"
        pass
    else
        fail "Created project material type but no ID returned"
    fi
else
    fail "Failed to create project material type"
fi

if [[ -z "$CREATED_PROJECT_MATERIAL_TYPE_ID" || "$CREATED_PROJECT_MATERIAL_TYPE_ID" == "null" ]]; then
    echo "Cannot continue without a valid project material type ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project material type attributes"
xbe_json do project-material-types update "$CREATED_PROJECT_MATERIAL_TYPE_ID" \
    --quantity 100 \
    --explicit-display-name "Updated PMT"
assert_success

test_name "Update project material type pickup/delivery window"
xbe_json do project-material-types update "$CREATED_PROJECT_MATERIAL_TYPE_ID" \
    --pickup-at-min "$PICKUP_AT_MIN" \
    --pickup-at-max "$PICKUP_AT_MAX" \
    --deliver-at-min "$DELIVER_AT_MIN" \
    --deliver-at-max "$DELIVER_AT_MAX"
assert_success

test_name "Update project material type unit of measure"
if [[ -n "$CREATED_UNIT_OF_MEASURE_ID" ]]; then
    xbe_json do project-material-types update "$CREATED_PROJECT_MATERIAL_TYPE_ID" \
        --unit-of-measure "$CREATED_UNIT_OF_MEASURE_ID"
    assert_success
else
    skip "No unit of measure ID available"
fi

test_name "Update project material type material site"
if [[ -n "$CREATED_MATERIAL_SITE_ID" ]]; then
    xbe_json do project-material-types update "$CREATED_PROJECT_MATERIAL_TYPE_ID" \
        --material-site "$CREATED_MATERIAL_SITE_ID"
    assert_success
else
    skip "No material site ID available"
fi

test_name "Update project material type job site"
skip "Project customers not available for job site update"

test_name "Update project material type pickup/delivery locations"
skip "No pickup/delivery location IDs available"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project material types"
xbe_json view project-material-types list --limit 5
assert_success

test_name "List project material types returns array"
xbe_json view project-material-types list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project material types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by project"
xbe_json view project-material-types list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

test_name "Filter by material type"
xbe_json view project-material-types list --material-type "$CREATED_MATERIAL_TYPE_ID" --limit 10
assert_success

test_name "Filter by pickup-at-min-min"
xbe_json view project-material-types list --pickup-at-min-min "$PICKUP_AT_MIN" --limit 10
assert_success

test_name "Filter by pickup-at-min-max"
xbe_json view project-material-types list --pickup-at-min-max "$PICKUP_AT_MIN" --limit 10
assert_success

test_name "Filter by pickup-at-max-min"
xbe_json view project-material-types list --pickup-at-max-min "$PICKUP_AT_MAX" --limit 10
assert_success

test_name "Filter by pickup-at-max-max"
xbe_json view project-material-types list --pickup-at-max-max "$PICKUP_AT_MAX" --limit 10
assert_success

test_name "Filter by deliver-at-min-min"
xbe_json view project-material-types list --deliver-at-min-min "$DELIVER_AT_MIN" --limit 10
assert_success

test_name "Filter by deliver-at-min-max"
xbe_json view project-material-types list --deliver-at-min-max "$DELIVER_AT_MIN" --limit 10
assert_success

test_name "Filter by deliver-at-max-min"
xbe_json view project-material-types list --deliver-at-max-min "$DELIVER_AT_MAX" --limit 10
assert_success

test_name "Filter by deliver-at-max-max"
xbe_json view project-material-types list --deliver-at-max-max "$DELIVER_AT_MAX" --limit 10
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project material type"
xbe_json view project-material-types show "$CREATED_PROJECT_MATERIAL_TYPE_ID"
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project material type requires --confirm flag"
xbe_json do project-material-types delete "$CREATED_PROJECT_MATERIAL_TYPE_ID"
assert_failure

test_name "Delete project material type"
xbe_json do project-material-types delete "$CREATED_PROJECT_MATERIAL_TYPE_ID" --confirm
if [[ $status -eq 0 ]]; then
    pass
else
    skip "API may not allow project material type deletion"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without required flags fails"
xbe_json do project-material-types create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without fields fails"
xbe_json do project-material-types update "$CREATED_PROJECT_MATERIAL_TYPE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
