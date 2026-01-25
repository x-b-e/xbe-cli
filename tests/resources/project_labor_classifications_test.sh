#!/bin/bash
#
# XBE CLI Integration Tests: Project Labor Classifications
#
# Tests CRUD operations for the project_labor_classifications resource.
# Project labor classifications link projects to labor classifications and store wage rates.
#
# NOTE: This test requires creating prerequisite resources: broker, developer, project, labor classification
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_LABOR_CLASSIFICATION_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_LABOR_CLASSIFICATION_ID=""

DESCRIBE_NAME="Resource: project_labor_classifications"

describe "$DESCRIBE_NAME"

# ============================================================================
# Prerequisites - Create broker, developer, project, labor classification
# ============================================================================

test_name "Create prerequisite broker for project labor classification tests"
BROKER_NAME=$(unique_name "PLCTestBroker")

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

test_name "Create prerequisite developer for project"
DEVELOPER_NAME=$(unique_name "PLCTestDev")

xbe_json do developers create \
    --name "$DEVELOPER_NAME" \
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

test_name "Create prerequisite project"
PROJECT_NAME=$(unique_name "PLCTestProject")

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

test_name "Create prerequisite labor classification"
LABOR_NAME=$(unique_name "PLCTestLabor")

xbe_json do labor-classifications create --name "$LABOR_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_LABOR_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_LABOR_CLASSIFICATION_ID" && "$CREATED_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "labor-classifications" "$CREATED_LABOR_CLASSIFICATION_ID"
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

test_name "Create project labor classification with required fields"

xbe_json do project-labor-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_LABOR_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" && "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-labor-classifications" "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID"
        pass
    else
        fail "Created project labor classification but no ID returned"
    fi
else
    fail "Failed to create project labor classification"
fi

# Only continue if we successfully created a project labor classification
if [[ -z "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" || "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project labor classification ID"
    run_tests
fi

test_name "Create project labor classification with rates"

RATE_CLASS_NAME=$(unique_name "PLCWithRates")

xbe_json do project-labor-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" \
    --basic-hourly-rate "45" \
    --fringe-hourly-rate "12"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-labor-classifications" "$id"
    pass
else
    fail "Failed to create project labor classification with rates"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project labor classification basic hourly rate"

xbe_json do project-labor-classifications update "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" --basic-hourly-rate "50"
assert_success

test_name "Update project labor classification fringe hourly rate"

xbe_json do project-labor-classifications update "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID" --fringe-hourly-rate "15"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project labor classification"

xbe_json view project-labor-classifications show "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to show project labor classification"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project labor classifications"

xbe_json view project-labor-classifications list --limit 5
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project labor classifications with --project filter"

xbe_json view project-labor-classifications list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

test_name "List project labor classifications with --labor-classification filter"

xbe_json view project-labor-classifications list --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project labor classifications with --limit"

xbe_json view project-labor-classifications list --limit 3
assert_success

test_name "List project labor classifications with --offset"

xbe_json view project-labor-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project labor classification requires --confirm flag"

xbe_run do project-labor-classifications delete "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID"
assert_failure

test_name "Delete project labor classification with --confirm"

TEST_DEL_NAME=$(unique_name "PLCDelete")

xbe_json do project-labor-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-labor-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project labor classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project labor classification without project fails"

xbe_json do project-labor-classifications create --labor-classification "$CREATED_LABOR_CLASSIFICATION_ID"
assert_failure

test_name "Create project labor classification without labor classification fails"

xbe_json do project-labor-classifications create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without any fields fails"

xbe_json do project-labor-classifications update "$CREATED_PROJECT_LABOR_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
