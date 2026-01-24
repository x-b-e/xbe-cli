#!/bin/bash
#
# XBE CLI Integration Tests: Project Margin Matrices
#
# Tests list, show, create, and delete operations for the project-margin-matrices resource.
#
# COVERAGE: List filters + show + create/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_MARGIN_MATRIX_ID=""

SAMPLE_ID=""
SAMPLE_PROJECT_ID=""
LIST_SUPPORTED="false"

describe "Resource: project-margin-matrices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project margin matrices"
xbe_json view project-margin-matrices list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Project margin matrices list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List project margin matrices returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-margin-matrices list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list project margin matrices"
    fi
else
    skip "Project margin matrices list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show)
# ==========================================================================

test_name "Capture sample project margin matrix"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view project-margin-matrices list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PROJECT_ID=$(json_get ".[0].project_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No project margin matrices available for show"
        fi
    else
        skip "Could not list project margin matrices to capture sample"
    fi
else
    skip "Project margin matrices list endpoint not available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project margin matrices with --project filter"
if [[ -n "$SAMPLE_PROJECT_ID" && "$SAMPLE_PROJECT_ID" != "null" ]]; then
    xbe_json view project-margin-matrices list --project "$SAMPLE_PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available for filter test"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project margin matrix"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-margin-matrices show "$SAMPLE_ID"
    assert_success
else
    skip "No project margin matrix ID available"
fi

# ============================================================================
# Prerequisites - Create broker, developer, and project
# ============================================================================

test_name "Create prerequisite broker"
BROKER_NAME=$(unique_name "ProjectMarginMatrixBroker")

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
DEVELOPER_NAME=$(unique_name "ProjectMarginMatrixDeveloper")

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

test_name "Create project for margin matrix"
PROJECT_NAME=$(unique_name "ProjectMarginMatrix")

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

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project margin matrix without required project fails"
xbe_run do project-margin-matrices create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project margin matrix"
xbe_json do project-margin-matrices create \
    --project "$CREATED_PROJECT_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_MARGIN_MATRIX_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_MARGIN_MATRIX_ID" && "$CREATED_PROJECT_MARGIN_MATRIX_ID" != "null" ]]; then
        register_cleanup "project-margin-matrices" "$CREATED_PROJECT_MARGIN_MATRIX_ID"
        assert_json_equals ".project_id" "$CREATED_PROJECT_ID"
    else
        fail "Created project margin matrix but no ID returned"
    fi
else
    fail "Failed to create project margin matrix"
fi

test_name "List project margin matrices with --project filter (created project)"
if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
    xbe_json view project-margin-matrices list --project "$CREATED_PROJECT_ID" --limit 5
    assert_success
else
    skip "No created project ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project margin matrix"
if [[ -n "$CREATED_PROJECT_MARGIN_MATRIX_ID" && "$CREATED_PROJECT_MARGIN_MATRIX_ID" != "null" ]]; then
    xbe_run do project-margin-matrices delete "$CREATED_PROJECT_MARGIN_MATRIX_ID" --confirm
    assert_success
else
    skip "No project margin matrix created for deletion"
fi

run_tests
