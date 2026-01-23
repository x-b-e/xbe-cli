#!/bin/bash
#
# XBE CLI Integration Tests: Project Revenue Classifications
#
# Tests CRUD operations for the project_revenue_classifications resource.
# Project revenue classifications define revenue categories for projects.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_CLASSIFICATION_ID=""
CREATED_BROKER_ID=""

describe "Resource: project_revenue_classifications"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for project revenue classification tests"
BROKER_NAME=$(unique_name "PRCTestBroker")

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
# CREATE Tests
# ============================================================================

test_name "Create project revenue classification with required fields"
TEST_NAME=$(unique_name "ProjRevClass")

xbe_json do project-revenue-classifications create \
    --name "$TEST_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-revenue-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created project revenue classification but no ID returned"
    fi
else
    fail "Failed to create project revenue classification"
fi

# Only continue if we successfully created a classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project revenue classification ID"
    run_tests
fi

test_name "Create project revenue classification with parent"
TEST_NAME2=$(unique_name "ProjRevClass2")
xbe_json do project-revenue-classifications create \
    --name "$TEST_NAME2" \
    --broker "$CREATED_BROKER_ID" \
    --parent "$CREATED_CLASSIFICATION_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-revenue-classifications" "$id"
    pass
else
    fail "Failed to create project revenue classification with parent"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project revenue classification name"
UPDATED_NAME=$(unique_name "UpdatedPRC")
xbe_json do project-revenue-classifications update "$CREATED_CLASSIFICATION_ID" --name "$UPDATED_NAME"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project revenue classifications"
xbe_json view project-revenue-classifications list --limit 5
assert_success

test_name "List project revenue classifications returns array"
xbe_json view project-revenue-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project revenue classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project revenue classifications with --broker filter"
xbe_json view project-revenue-classifications list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project revenue classifications with --limit"
xbe_json view project-revenue-classifications list --limit 3
assert_success

test_name "List project revenue classifications with --offset"
xbe_json view project-revenue-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project revenue classification requires --confirm flag"
xbe_run do project-revenue-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Delete project revenue classification with --confirm"
# Create a classification specifically for deletion
TEST_DEL_NAME=$(unique_name "DeletePRC")
xbe_json do project-revenue-classifications create \
    --name "$TEST_DEL_NAME" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-revenue-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project revenue classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project revenue classification without name fails"
xbe_json do project-revenue-classifications create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project revenue classification without broker fails"
xbe_json do project-revenue-classifications create --name "NoBroker"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-revenue-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
