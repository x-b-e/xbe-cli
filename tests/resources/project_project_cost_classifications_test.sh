#!/bin/bash
#
# XBE CLI Integration Tests: Project Project Cost Classifications
#
# Tests CRUD operations for the project_project_cost_classifications resource.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_COST_CLASSIFICATION_ID=""
CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID=""

describe "Resource: project_project_cost_classifications"

# ============================================================================
# Prerequisites - Create broker, developer, project, project cost classification
# ============================================================================

test_name "Create prerequisite broker for project project cost classification tests"
BROKER_NAME=$(unique_name "PPCCBroker")

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

test_name "Create developer for project project cost classification tests"
if [[ -n "$XBE_TEST_DEVELOPER_ID" ]]; then
    CREATED_DEVELOPER_ID="$XBE_TEST_DEVELOPER_ID"
    echo "    Using XBE_TEST_DEVELOPER_ID: $CREATED_DEVELOPER_ID"
    pass
else
    DEV_NAME=$(unique_name "PPCCDev")
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

test_name "Create project for project project cost classification tests"
if [[ -n "$XBE_TEST_PROJECT_ID" ]]; then
    CREATED_PROJECT_ID="$XBE_TEST_PROJECT_ID"
    echo "    Using XBE_TEST_PROJECT_ID: $CREATED_PROJECT_ID"
    pass
else
    PROJECT_NAME=$(unique_name "PPCCProject")
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

test_name "Create project cost classification for project project cost classification tests"
CLASSIFICATION_NAME=$(unique_name "PPCCClass")
xbe_json do project-cost-classifications create \
    --name "$CLASSIFICATION_NAME" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_COST_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_COST_CLASSIFICATION_ID" && "$CREATED_PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-cost-classifications" "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
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
# CREATE Tests
# ============================================================================

test_name "Create project project cost classification with required fields"
NAME_OVERRIDE=$(unique_name "PPCCOverride")
xbe_json do project-project-cost-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID" \
    --name-override "$NAME_OVERRIDE"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID" && "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-project-cost-classifications" "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID"
        pass
    else
        fail "Created project project cost classification but no ID returned"
    fi
else
    fail "Failed to create project project cost classification"
fi

if [[ -z "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID" || "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project project cost classification ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project project cost classification"
xbe_json view project-project-cost-classifications show "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project project cost classification name override"
UPDATED_NAME_OVERRIDE=$(unique_name "PPCCUpdated")
xbe_json do project-project-cost-classifications update "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID" --name-override "$UPDATED_NAME_OVERRIDE"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project project cost classifications"
xbe_json view project-project-cost-classifications list --limit 5
assert_success

test_name "List project project cost classifications returns array"
xbe_json view project-project-cost-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project project cost classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project project cost classifications with --project filter"
xbe_json view project-project-cost-classifications list --project "$CREATED_PROJECT_ID" --limit 5
assert_success

test_name "List project project cost classifications with --project-cost-classification filter"
xbe_json view project-project-cost-classifications list --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID" --limit 5
assert_success

test_name "List project project cost classifications with --created-at-min filter"
xbe_json view project-project-cost-classifications list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project project cost classifications with --created-at-max filter"
xbe_json view project-project-cost-classifications list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project project cost classifications with --updated-at-min filter"
xbe_json view project-project-cost-classifications list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List project project cost classifications with --updated-at-max filter"
xbe_json view project-project-cost-classifications list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project project cost classification requires --confirm flag"
xbe_run do project-project-cost-classifications delete "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID"
assert_failure

test_name "Delete project project cost classification with --confirm"
NAME_OVERRIDE_DELETE=$(unique_name "PPCCDelete")
xbe_json do project-project-cost-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID" \
    --name-override "$NAME_OVERRIDE_DELETE"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-project-cost-classifications delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project project cost classification for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project project cost classification without project fails"
xbe_json do project-project-cost-classifications create --project-cost-classification "$CREATED_PROJECT_COST_CLASSIFICATION_ID"
assert_failure

test_name "Create project project cost classification without project cost classification fails"
xbe_json do project-project-cost-classifications create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-project-cost-classifications update "$CREATED_PROJECT_PROJECT_COST_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
