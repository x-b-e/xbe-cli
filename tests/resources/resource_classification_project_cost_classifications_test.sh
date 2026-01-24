#!/bin/bash
#
# XBE CLI Integration Tests: Resource Classification Project Cost Classifications
#
# Tests CRUD operations for the resource_classification_project_cost_classifications resource.
# Links resource classifications (labor/equipment) to project cost classifications.
#
# COVERAGE: All create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ASSOCIATION_ID=""
CREATED_BROKER_ID=""
CREATED_PROJECT_COST_CLASS_ID=""
CREATED_LABOR_CLASS_ID=""

describe "Resource: resource_classification_project_cost_classifications"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "RCPCCTestBroker")

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

# ============================================================================
# Prerequisites - Create project cost classification
# ============================================================================

test_name "Create project cost classification"
PCC_NAME=$(unique_name "RCPCCTestPCC")

xbe_json do project-cost-classifications create --name "$PCC_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_COST_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_COST_CLASS_ID" && "$CREATED_PROJECT_COST_CLASS_ID" != "null" ]]; then
        register_cleanup "project-cost-classifications" "$CREATED_PROJECT_COST_CLASS_ID"
        pass
    else
        fail "Created project cost classification but no ID returned"
        echo "Cannot continue without a project cost classification"
        run_tests
    fi
else
    fail "Failed to create project cost classification"
    echo "Cannot continue without a project cost classification"
    run_tests
fi

# ============================================================================
# Prerequisites - Create labor classification
# ============================================================================

test_name "Create labor classification"
LABOR_NAME=$(unique_name "RCPCCTestLabor")

xbe_json do labor-classifications create --name "$LABOR_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASS_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASS_ID" && "$CREATED_LABOR_CLASS_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASS_ID"
        pass
    else
        fail "Created labor classification but no ID returned"
        echo "Cannot continue without a labor classification"
        run_tests
    fi
else
    fail "Failed to create labor classification"
    echo "Cannot continue without a labor classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create resource classification project cost classification"
xbe_json do resource-classification-project-cost-classifications create \
    --resource-classification-type LaborClassification \
    --resource-classification-id "$CREATED_LABOR_CLASS_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASS_ID" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_ASSOCIATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_ASSOCIATION_ID" && "$CREATED_ASSOCIATION_ID" != "null" ]]; then
        register_cleanup "resource-classification-project-cost-classifications" "$CREATED_ASSOCIATION_ID"
        pass
    else
        fail "Created association but no ID returned"
    fi
else
    fail "Failed to create resource classification project cost classification"
fi

# Only continue if we successfully created an association
if [[ -z "$CREATED_ASSOCIATION_ID" || "$CREATED_ASSOCIATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid association ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show resource classification project cost classification"
xbe_json view resource-classification-project-cost-classifications show "$CREATED_ASSOCIATION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List resource classification project cost classifications"
xbe_json view resource-classification-project-cost-classifications list --limit 5
assert_success

test_name "List resource classification project cost classifications returns array"
xbe_json view resource-classification-project-cost-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list resource classification project cost classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List with --resource-classification-type filter"
xbe_json view resource-classification-project-cost-classifications list \
    --resource-classification-type LaborClassification --limit 10
assert_success

test_name "List with resource classification type and id filter"
xbe_json view resource-classification-project-cost-classifications list \
    --resource-classification-type LaborClassification \
    --resource-classification-id "$CREATED_LABOR_CLASS_ID" \
    --limit 10
assert_success

test_name "List with --project-cost-classification filter"
xbe_json view resource-classification-project-cost-classifications list \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASS_ID" --limit 10
assert_success

test_name "List with --broker filter"
xbe_json view resource-classification-project-cost-classifications list \
    --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List with created-at range filters"
xbe_json view resource-classification-project-cost-classifications list \
    --created-at-min 2000-01-01T00:00:00Z \
    --created-at-max 2100-01-01T00:00:00Z \
    --limit 10
assert_success

test_name "List with updated-at range filters"
xbe_json view resource-classification-project-cost-classifications list \
    --updated-at-min 2000-01-01T00:00:00Z \
    --updated-at-max 2100-01-01T00:00:00Z \
    --limit 10
assert_success

test_name "List with --is-created-at filter"
xbe_json view resource-classification-project-cost-classifications list --is-created-at true --limit 10
assert_success

test_name "List with --is-updated-at filter"
xbe_json view resource-classification-project-cost-classifications list --is-updated-at true --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete requires --confirm flag"
xbe_run do resource-classification-project-cost-classifications delete "$CREATED_ASSOCIATION_ID"
assert_failure

# Create a second project cost classification for deletion test
TEST_DELETE_PCC_NAME=$(unique_name "RCPCCTestDeletePCC")
xbe_json do project-cost-classifications create --name "$TEST_DELETE_PCC_NAME" --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DELETE_PCC_ID=$(json_get ".id")
    register_cleanup "project-cost-classifications" "$DELETE_PCC_ID"
else
    skip "Could not create project cost classification for deletion test"
    run_tests
fi

# Create association for deletion
xbe_json do resource-classification-project-cost-classifications create \
    --resource-classification-type LaborClassification \
    --resource-classification-id "$CREATED_LABOR_CLASS_ID" \
    --project-cost-classification "$DELETE_PCC_ID" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    DELETE_ASSOC_ID=$(json_get ".id")
    xbe_run do resource-classification-project-cost-classifications delete "$DELETE_ASSOC_ID" --confirm
    assert_success
else
    skip "Could not create association for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create without resource classification type fails"
xbe_json do resource-classification-project-cost-classifications create \
    --resource-classification-id "$CREATED_LABOR_CLASS_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASS_ID" \
    --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create without project cost classification fails"
xbe_json do resource-classification-project-cost-classifications create \
    --resource-classification-type LaborClassification \
    --resource-classification-id "$CREATED_LABOR_CLASS_ID" \
    --broker "$CREATED_BROKER_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
