#!/bin/bash
#
# XBE CLI Integration Tests: Project Revenue Items
#
# Tests list, show, create, update, delete operations for the
# project-revenue-items resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_REVENUE_CLASSIFICATION_ID=""
CREATED_REVENUE_CLASSIFICATION_ID_2=""
UNIT_OF_MEASURE_ID=""
UNIT_OF_MEASURE_ID_2=""
CREATED_ITEM_ID=""

describe "Resource: project_revenue_items"

# ============================================================================
# Prerequisites - Create broker, developer, project, revenue classification
# ============================================================================

test_name "Create prerequisite broker for project revenue item tests"
BROKER_NAME=$(unique_name "PRIBroker")

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

test_name "Create developer for project revenue item tests"
if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
    CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
    echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
    pass
else
    DEV_NAME=$(unique_name "PRIDev")
    xbe_json do developers create \
        --name "$DEV_NAME" \
        --broker "$CREATED_BROKER_ID"

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
fi

test_name "Create project for project revenue item tests"
if [[ -n "$XBE_TEST_PROJECT_ID" ]]; then
    CREATED_PROJECT_ID="$XBE_TEST_PROJECT_ID"
    echo "    Using XBE_TEST_PROJECT_ID: $CREATED_PROJECT_ID"
    pass
else
    PROJECT_NAME=$(unique_name "PRIProject")
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
fi

test_name "Create revenue classification for project revenue item tests"
CLASSIFICATION_NAME=$(unique_name "PRIClass")
xbe_json do project-revenue-classifications create \
    --name "$CLASSIFICATION_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_REVENUE_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_REVENUE_CLASSIFICATION_ID" && "$CREATED_REVENUE_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-revenue-classifications" "$CREATED_REVENUE_CLASSIFICATION_ID"
        pass
    else
        fail "Created revenue classification but no ID returned"
        echo "Cannot continue without a revenue classification"
        run_tests
    fi
else
    fail "Failed to create revenue classification"
    echo "Cannot continue without a revenue classification"
    run_tests
fi

test_name "Find unit of measure for project revenue item tests"
if [[ -n "$XBE_TEST_UNIT_OF_MEASURE_ID" ]]; then
    UNIT_OF_MEASURE_ID="$XBE_TEST_UNIT_OF_MEASURE_ID"
    echo "    Using XBE_TEST_UNIT_OF_MEASURE_ID: $UNIT_OF_MEASURE_ID"
    pass
else
    xbe_json view unit-of-measures list --limit 5
    if [[ $status -eq 0 ]]; then
        UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
        UNIT_OF_MEASURE_ID_2=$(echo "$output" | jq -r '.[].id' | grep -v "$UNIT_OF_MEASURE_ID" | head -n 1)
        if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
            pass
        else
            fail "Could not find unit of measure ID"
            echo "Cannot continue without a unit of measure"
            run_tests
        fi
    else
        fail "Failed to list unit of measures"
        echo "Cannot continue without a unit of measure"
        run_tests
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project revenue item with required fields"
ITEM_DESCRIPTION=$(unique_name "PRI")
EXTERNAL_ID="EXT-$(unique_suffix)"
DEV_QTY_ESTIMATE="1250.5"

xbe_json do project-revenue-items create \
    --project "$CREATED_PROJECT_ID" \
    --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --description "$ITEM_DESCRIPTION" \
    --external-developer-revenue-item-id "$EXTERNAL_ID" \
    --developer-quantity-estimate "$DEV_QTY_ESTIMATE"

if [[ $status -eq 0 ]]; then
    CREATED_ITEM_ID=$(json_get ".id")
    if [[ -n "$CREATED_ITEM_ID" && "$CREATED_ITEM_ID" != "null" ]]; then
        register_cleanup "project-revenue-items" "$CREATED_ITEM_ID"
        pass
    else
        fail "Created project revenue item but no ID returned"
    fi
else
    fail "Failed to create project revenue item"
fi

if [[ -z "$CREATED_ITEM_ID" || "$CREATED_ITEM_ID" == "null" ]]; then
    echo "Cannot continue without a valid project revenue item ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project revenue item"
xbe_json view project-revenue-items show "$CREATED_ITEM_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project revenue item description"
UPDATED_DESC=$(unique_name "UpdatedPRI")
xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --description "$UPDATED_DESC"
assert_success

test_name "Update project revenue item external developer ID"
UPDATED_EXTERNAL_ID="EXT-UPDATED-$(unique_suffix)"
xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --external-developer-revenue-item-id "$UPDATED_EXTERNAL_ID"
assert_success

test_name "Update project revenue item developer quantity estimate"
UPDATED_DEV_QTY="1500"
xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --developer-quantity-estimate "$UPDATED_DEV_QTY"
assert_success

test_name "Update project revenue item revenue classification"
CLASSIFICATION_NAME_2=$(unique_name "PRIClass2")
xbe_json do project-revenue-classifications create \
    --name "$CLASSIFICATION_NAME_2" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    CREATED_REVENUE_CLASSIFICATION_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_REVENUE_CLASSIFICATION_ID_2" && "$CREATED_REVENUE_CLASSIFICATION_ID_2" != "null" ]]; then
        register_cleanup "project-revenue-classifications" "$CREATED_REVENUE_CLASSIFICATION_ID_2"
        xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID_2"
        assert_success
    else
        skip "Created revenue classification but no ID returned"
    fi
else
    skip "Could not create revenue classification for update"
fi

test_name "Update project revenue item unit of measure"
if [[ -n "$UNIT_OF_MEASURE_ID_2" && "$UNIT_OF_MEASURE_ID_2" != "null" ]]; then
    xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --unit-of-measure "$UNIT_OF_MEASURE_ID_2"
    assert_success
else
    skip "No alternate unit of measure available"
fi

test_name "Update project revenue item quantity estimate"
xbe_json view project-revenue-items show "$CREATED_ITEM_ID"
if [[ $status -eq 0 ]]; then
    QUANTITY_ESTIMATE_ID=$(json_get ".quantity_estimate_id")
    if [[ -n "$QUANTITY_ESTIMATE_ID" && "$QUANTITY_ESTIMATE_ID" != "null" ]]; then
        xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --quantity-estimate "$QUANTITY_ESTIMATE_ID"
        assert_success
    else
        skip "No quantity estimate available to update"
    fi
else
    skip "Could not fetch project revenue item to update quantity estimate"
fi

test_name "Update project revenue item price estimate"
xbe_json view project-revenue-items show "$CREATED_ITEM_ID"
if [[ $status -eq 0 ]]; then
    PRICE_ESTIMATE_ID=$(json_get ".price_estimate_id")
    if [[ -n "$PRICE_ESTIMATE_ID" && "$PRICE_ESTIMATE_ID" != "null" ]]; then
        xbe_json do project-revenue-items update "$CREATED_ITEM_ID" --price-estimate "$PRICE_ESTIMATE_ID"
        assert_success
    else
        skip "No price estimate available to update"
    fi
else
    skip "Could not fetch project revenue item to update price estimate"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project revenue items"
xbe_json view project-revenue-items list --limit 5
assert_success

test_name "List project revenue items returns array"
xbe_json view project-revenue-items list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project revenue items"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project revenue items with --project filter"
xbe_json view project-revenue-items list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

test_name "List project revenue items with --revenue-classification filter"
xbe_json view project-revenue-items list --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID" --limit 5
assert_success

test_name "List project revenue items with --unit-of-measure filter"
xbe_json view project-revenue-items list --unit-of-measure "$UNIT_OF_MEASURE_ID" --limit 5
assert_success

test_name "List project revenue items with --developer-quantity-estimate filter"
xbe_json view project-revenue-items list --developer-quantity-estimate "$UPDATED_DEV_QTY" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project revenue items with --limit"
xbe_json view project-revenue-items list --limit 3
assert_success

test_name "List project revenue items with --offset"
xbe_json view project-revenue-items list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project revenue item requires --confirm flag"
xbe_run do project-revenue-items delete "$CREATED_ITEM_ID"
assert_failure

test_name "Delete project revenue item with --confirm"
ITEM_DESCRIPTION_DEL=$(unique_name "PRIDelete")
xbe_json do project-revenue-items create \
    --project "$CREATED_PROJECT_ID" \
    --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --description "$ITEM_DESCRIPTION_DEL"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-revenue-items delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project revenue item for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project revenue item without project fails"
xbe_json do project-revenue-items create \
    --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID"
assert_failure

test_name "Create project revenue item without revenue classification fails"
xbe_json do project-revenue-items create \
    --project "$CREATED_PROJECT_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID"
assert_failure

test_name "Create project revenue item without unit of measure fails"
xbe_json do project-revenue-items create \
    --project "$CREATED_PROJECT_ID" \
    --revenue-classification "$CREATED_REVENUE_CLASSIFICATION_ID"
assert_failure

test_name "Update project revenue item without fields fails"
xbe_json do project-revenue-items update "$CREATED_ITEM_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
