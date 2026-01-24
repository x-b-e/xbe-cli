#!/bin/bash
#
# XBE CLI Integration Tests: Project Revenue Item Price Estimates
#
# Tests create/update/delete operations and list filters for project-revenue-item-price-estimates.
# Requires creating supporting project revenue items and estimate sets.
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_REVENUE_CLASSIFICATION_ID=""
CREATED_PROJECT_REVENUE_ITEM_ID=""
CREATED_PROJECT_ESTIMATE_SET_ID=""
UNIT_OF_MEASURE_ID=""
CREATED_BY_ID=""

SKIP_MUTATION=0
if [[ -z "$XBE_TOKEN" ]]; then
    SKIP_MUTATION=1
fi

describe "Resource: project-revenue-item-price-estimates"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "PripBroker")

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

test_name "Create prerequisite developer"
DEV_NAME=$(unique_name "PripDeveloper")

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

test_name "Create prerequisite project"
PROJECT_NAME=$(unique_name "PripProject")

xbe_json do projects create \
    --name "$PROJECT_NAME" \
    --developer "$CREATED_DEVELOPER_ID"

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

test_name "Resolve unit of measure"
xbe_json view unit-of-measures list --limit 1
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
        pass
    else
        fail "No unit of measure found"
        echo "Cannot continue without a unit of measure"
        run_tests
    fi
else
    fail "Failed to list unit of measures"
    echo "Cannot continue without a unit of measure"
    run_tests
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Create project revenue classification"
    REV_CLASS_NAME=$(unique_name "PripRevClass")
    xbe_json do project-revenue-classifications create --name "$REV_CLASS_NAME" --broker "$CREATED_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_REVENUE_CLASSIFICATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" && "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID" != "null" ]]; then
            register_cleanup "project-revenue-classifications" "$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID"
            pass
        else
            fail "Created project revenue classification but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create project revenue classification"
        run_tests
    fi
fi

cleanup_project_revenue_item_price_estimates() {
    if [[ -n "$XBE_TOKEN" ]]; then
        if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" && "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-revenue-item-price-estimates/$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_ESTIMATE_SET_ID" && "$CREATED_PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-estimate-sets/$CREATED_PROJECT_ESTIMATE_SET_ID" \
                >/dev/null 2>&1 || true
        fi

        if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_ID" && "$CREATED_PROJECT_REVENUE_ITEM_ID" != "null" ]]; then
            curl -s -f \
                -H "Authorization: Bearer $XBE_TOKEN" \
                -H "Accept: application/vnd.api+json" \
                -X DELETE "$XBE_BASE_URL/v1/project-revenue-items/$CREATED_PROJECT_REVENUE_ITEM_ID" \
                >/dev/null 2>&1 || true
        fi
    fi

    run_cleanup
}
trap cleanup_project_revenue_item_price_estimates EXIT

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project revenue item price estimate without required fields fails"
xbe_json do project-revenue-item-price-estimates create
assert_failure

if [[ $SKIP_MUTATION -eq 1 ]]; then
    skip "Skipping create/update/delete/filter tests without XBE_TOKEN"
    test_name "List project revenue item price estimates"
    xbe_json view project-revenue-item-price-estimates list --limit 5
    assert_success
    run_tests
fi

# Resolve current user for created-by filter tests
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CREATED_BY_ID=$(json_get ".id")
fi

# Create project revenue item (direct API)
PROJECT_REVENUE_ITEM_DESC=$(unique_name "PripRevItem")
revenue_item_payload=$(cat <<JSON
{"data":{"type":"project-revenue-items","attributes":{"description":"$PROJECT_REVENUE_ITEM_DESC"},"relationships":{"project":{"data":{"type":"projects","id":"$CREATED_PROJECT_ID"}},"revenue-classification":{"data":{"type":"project-revenue-classifications","id":"$CREATED_PROJECT_REVENUE_CLASSIFICATION_ID"}},"unit-of-measure":{"data":{"type":"unit-of-measures","id":"$UNIT_OF_MEASURE_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-revenue-items" \
    -d "$revenue_item_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_REVENUE_ITEM_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_ID" && "$CREATED_PROJECT_REVENUE_ITEM_ID" != "null" ]]; then
        pass
    else
        fail "Created project revenue item but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project revenue item"
    run_tests
fi

# Create project estimate set (direct API)
ESTIMATE_SET_NAME=$(unique_name "PripEstimateSet")
estimate_set_payload=$(cat <<JSON
{"data":{"type":"project-estimate-sets","attributes":{"name":"$ESTIMATE_SET_NAME"},"relationships":{"project":{"data":{"type":"projects","id":"$CREATED_PROJECT_ID"}}}}}
JSON
)

run curl -s -f \
    -H "Authorization: Bearer $XBE_TOKEN" \
    -H "Accept: application/vnd.api+json" \
    -H "Content-Type: application/vnd.api+json" \
    -X POST "$XBE_BASE_URL/v1/project-estimate-sets" \
    -d "$estimate_set_payload"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ESTIMATE_SET_ID=$(echo "$output" | jq -r '.data.id // empty')
    if [[ -n "$CREATED_PROJECT_ESTIMATE_SET_ID" && "$CREATED_PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
        pass
    else
        fail "Created project estimate set but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project estimate set"
    run_tests
fi

# Create project revenue item price estimate
PRICE_ESTIMATE_VALUE="55.25"
test_name "Create project revenue item price estimate"
xbe_json do project-revenue-item-price-estimates create \
    --project-revenue-item "$CREATED_PROJECT_REVENUE_ITEM_ID" \
    --project-estimate-set "$CREATED_PROJECT_ESTIMATE_SET_ID" \
    --kind explicit \
    --price-per-unit-explicit "$PRICE_ESTIMATE_VALUE"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" && "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" != "null" ]]; then
        register_cleanup "project-revenue-item-price-estimates" "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID"
        pass
    else
        fail "Created project revenue item price estimate but no ID returned"
        run_tests
    fi
else
    fail "Failed to create project revenue item price estimate"
    run_tests
fi

# ==========================================================================
# SHOW/LIST Tests
# ==========================================================================

test_name "Show project revenue item price estimate"
xbe_json view project-revenue-item-price-estimates show "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID"
assert_success

# List filters

test_name "List project revenue item price estimates filtered by project revenue item"
xbe_json view project-revenue-item-price-estimates list --project-revenue-item "$CREATED_PROJECT_REVENUE_ITEM_ID"
assert_success

if [[ -n "$CREATED_PROJECT_ESTIMATE_SET_ID" ]]; then
    test_name "List project revenue item price estimates filtered by project estimate set"
    xbe_json view project-revenue-item-price-estimates list --project-estimate-set "$CREATED_PROJECT_ESTIMATE_SET_ID"
    assert_success
fi

if [[ -n "$CREATED_BY_ID" ]]; then
    test_name "List project revenue item price estimates filtered by created-by"
    xbe_json view project-revenue-item-price-estimates list --created-by "$CREATED_BY_ID"
    assert_success
fi

# ==========================================================================
# UPDATE/DELETE Tests
# ==========================================================================

test_name "Update project revenue item price estimate"
xbe_json do project-revenue-item-price-estimates update "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" \
    --kind cost_multiplier \
    --cost-multiplier 1.1
assert_success

test_name "Delete project revenue item price estimate"
xbe_json do project-revenue-item-price-estimates delete "$CREATED_PROJECT_REVENUE_ITEM_PRICE_ESTIMATE_ID" --confirm
assert_success

run_tests
