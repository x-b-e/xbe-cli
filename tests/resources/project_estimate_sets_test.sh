#!/bin/bash
#
# XBE CLI Integration Tests: Project Estimate Sets
#
# Tests CRUD operations for the project-estimate-sets resource.
# Project estimate sets require a project relationship.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_ESTIMATE_SET_ID=""
BACKUP_ESTIMATE_SET_ID=""
CURRENT_USER_ID=""

describe "Resource: project-estimate-sets"

# ============================================================================
# Prerequisites - Create resources for project estimate set tests
# ============================================================================

test_name "Create prerequisite broker for project estimate set tests"
BROKER_NAME=$(unique_name "EstimateSetBroker")

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

test_name "Create prerequisite developer for project estimate set tests"
DEV_NAME=$(unique_name "EstimateSetDeveloper")

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

test_name "Create prerequisite project for project estimate set tests"
PROJECT_NAME=$(unique_name "EstimateSetProject")

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

test_name "Resolve current user for created-by tests"
xbe_json auth whoami

if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        fail "No user ID returned from auth whoami"
        echo "Cannot continue without current user ID"
        run_tests
    fi
else
    fail "Failed to resolve current user"
    echo "Cannot continue without current user ID"
    run_tests
fi

test_name "List project estimate sets for backup selection"
xbe_json view project-estimate-sets list --project "$CREATED_PROJECT_ID" --limit 10

if [[ $status -eq 0 ]]; then
    BACKUP_ESTIMATE_SET_ID=$(json_get ".[0].id")
    if [[ -n "$BACKUP_ESTIMATE_SET_ID" && "$BACKUP_ESTIMATE_SET_ID" != "null" ]]; then
        pass
    else
        fail "No project estimate set found for backup"
        echo "Cannot continue without a backup estimate set"
        run_tests
    fi
else
    fail "Failed to list project estimate sets"
    echo "Cannot continue without a backup estimate set"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project estimate set with name and backup"
ESTIMATE_SET_NAME=$(unique_name "EstimateSet")

xbe_json do project-estimate-sets create \
    --project "$CREATED_PROJECT_ID" \
    --name "$ESTIMATE_SET_NAME" \
    --created-by "$CURRENT_USER_ID" \
    --backup-estimate-set "$BACKUP_ESTIMATE_SET_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ESTIMATE_SET_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ESTIMATE_SET_ID" && "$CREATED_PROJECT_ESTIMATE_SET_ID" != "null" ]]; then
        register_cleanup "project-estimate-sets" "$CREATED_PROJECT_ESTIMATE_SET_ID"
        pass
    else
        fail "Created project estimate set but no ID returned"
    fi
else
    fail "Failed to create project estimate set"
fi

if [[ -z "$CREATED_PROJECT_ESTIMATE_SET_ID" || "$CREATED_PROJECT_ESTIMATE_SET_ID" == "null" ]]; then
    echo "Cannot continue without a valid project estimate set ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project estimate set name"
UPDATED_ESTIMATE_SET_NAME=$(unique_name "EstimateSetUpdated")

xbe_json do project-estimate-sets update "$CREATED_PROJECT_ESTIMATE_SET_ID" --name "$UPDATED_ESTIMATE_SET_NAME"
assert_success

test_name "Update project estimate set backup"
xbe_json do project-estimate-sets update "$CREATED_PROJECT_ESTIMATE_SET_ID" --backup-estimate-set "$BACKUP_ESTIMATE_SET_ID"
assert_success

test_name "Update project estimate set created-by"
xbe_json do project-estimate-sets update "$CREATED_PROJECT_ESTIMATE_SET_ID" --created-by "$CURRENT_USER_ID"
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project estimate set"
xbe_json view project-estimate-sets show "$CREATED_PROJECT_ESTIMATE_SET_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project estimate sets"
xbe_json view project-estimate-sets list --limit 5
assert_success

test_name "List project estimate sets returns array"
xbe_json view project-estimate-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project estimate sets"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project estimate sets with --project filter"
xbe_json view project-estimate-sets list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

test_name "List project estimate sets with --created-by filter"
xbe_json view project-estimate-sets list --created-by "$CURRENT_USER_ID" --limit 10
assert_success

test_name "List project estimate sets with --is-bid filter"
xbe_json view project-estimate-sets list --project "$CREATED_PROJECT_ID" --is-bid true --limit 10
assert_success

test_name "List project estimate sets with --is-actual filter"
xbe_json view project-estimate-sets list --project "$CREATED_PROJECT_ID" --is-actual true --limit 10
assert_success

test_name "List project estimate sets with --is-possible filter"
xbe_json view project-estimate-sets list --project "$CREATED_PROJECT_ID" --is-possible true --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination & Sorting
# ============================================================================

test_name "List project estimate sets with --limit"
xbe_json view project-estimate-sets list --limit 3
assert_success

test_name "List project estimate sets with --sort"
xbe_json view project-estimate-sets list --sort name --limit 5
assert_success

run_tests
