#!/bin/bash
#
# XBE CLI Integration Tests: Project Trailer Classifications
#
# Tests CRUD operations for the project_trailer_classifications resource.
# Project trailer classifications associate trailer classifications with projects.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_PROJECT_ID_2=""
TRAILER_CLASSIFICATION_ID=""
CREATED_CLASSIFICATION_ID=""

describe "Resource: project_trailer_classifications"

# ============================================================================
# Prerequisites - Create broker, developer, project, trailer classification
# ============================================================================

test_name "Create prerequisite broker for project trailer classification tests"
BROKER_NAME=$(unique_name "PTCTestBroker")

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

test_name "Create prerequisite developer for project trailer classification tests"
DEV_NAME=$(unique_name "PTCTestDeveloper")

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

test_name "Create prerequisite project for project trailer classification tests"
PROJECT_NAME=$(unique_name "PTCTestProject")

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

test_name "Fetch trailer classification"
xbe_json view trailer-classifications list --limit 1

if [[ $status -eq 0 ]]; then
    TRAILER_CLASSIFICATION_ID=$(json_get ".[0].id")
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        pass
    else
        fail "No trailer classification ID returned"
        echo "Cannot continue without a trailer classification"
        run_tests
    fi
else
    fail "Failed to list trailer classifications"
    echo "Cannot continue without a trailer classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create project trailer classification with required fields"
xbe_json do project-trailer-classifications create \
    --project "$CREATED_PROJECT_ID" \
    --trailer-classification "$TRAILER_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_CLASSIFICATION_ID" && "$CREATED_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "project-trailer-classifications" "$CREATED_CLASSIFICATION_ID"
        pass
    else
        fail "Created project trailer classification but no ID returned"
    fi
else
    fail "Failed to create project trailer classification"
fi

# Only continue if we successfully created a project trailer classification
if [[ -z "$CREATED_CLASSIFICATION_ID" || "$CREATED_CLASSIFICATION_ID" == "null" ]]; then
    echo "Cannot continue without a valid project trailer classification ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project trailer classification project labor classification (clear)"
xbe_json do project-trailer-classifications update "$CREATED_CLASSIFICATION_ID" --project-labor-classification ""
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project trailer classification"
xbe_json view project-trailer-classifications show "$CREATED_CLASSIFICATION_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project trailer classifications"
xbe_json view project-trailer-classifications list --limit 5
assert_success

test_name "List project trailer classifications returns array"
xbe_json view project-trailer-classifications list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project trailer classifications"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project trailer classifications with --project filter"
xbe_json view project-trailer-classifications list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

test_name "List project trailer classifications with --trailer-classification filter"
xbe_json view project-trailer-classifications list --trailer-classification "$TRAILER_CLASSIFICATION_ID" --limit 10
assert_success

test_name "List project trailer classifications with --project-labor-classification filter"
xbe_json view project-trailer-classifications list --project-labor-classification null --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project trailer classifications with --limit"
xbe_json view project-trailer-classifications list --limit 3
assert_success

test_name "List project trailer classifications with --offset"
xbe_json view project-trailer-classifications list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project trailer classification requires --confirm flag"
xbe_run do project-trailer-classifications delete "$CREATED_CLASSIFICATION_ID"
assert_failure

test_name "Create second project for delete test"
PROJECT_NAME_2=$(unique_name "PTCTestProjectDelete")

xbe_json do projects create \
    --name "$PROJECT_NAME_2" \
    --developer "$CREATED_DEVELOPER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_ID_2=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_ID_2" && "$CREATED_PROJECT_ID_2" != "null" ]]; then
        register_cleanup "projects" "$CREATED_PROJECT_ID_2"
        pass
    else
        fail "Created project but no ID returned"
        echo "Skipping delete test"
    fi
else
    fail "Failed to create second project"
    echo "Skipping delete test"
fi

test_name "Delete project trailer classification with --confirm"
if [[ -n "$CREATED_PROJECT_ID_2" && "$CREATED_PROJECT_ID_2" != "null" ]]; then
    xbe_json do project-trailer-classifications create \
        --project "$CREATED_PROJECT_ID_2" \
        --trailer-classification "$TRAILER_CLASSIFICATION_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_run do project-trailer-classifications delete "$DEL_ID" --confirm
        assert_success
    else
        skip "Could not create project trailer classification for deletion test"
    fi
else
    skip "Missing second project ID; skipping deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project trailer classification without project fails"
xbe_json do project-trailer-classifications create --trailer-classification "$TRAILER_CLASSIFICATION_ID"
assert_failure

test_name "Create project trailer classification without trailer classification fails"
xbe_json do project-trailer-classifications create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-trailer-classifications update "$CREATED_CLASSIFICATION_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
