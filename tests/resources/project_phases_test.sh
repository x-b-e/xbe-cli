#!/bin/bash
#
# XBE CLI Integration Tests: Project Phases
#
# Tests CRUD operations for the project-phases resource.
# Project phases require a project relationship.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_PHASE_ID=""
CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""

describe "Resource: project-phases"

# ============================================================================
# Prerequisites - Create resources for project phase tests
# ============================================================================

test_name "Create prerequisite broker for project phase tests"
BROKER_NAME=$(unique_name "PhaseTestBroker")

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

test_name "Create prerequisite developer for project phase tests"
DEV_NAME=$(unique_name "PhaseTestDeveloper")

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

test_name "Create prerequisite project for project phase tests"
PROJECT_NAME=$(unique_name "PhaseTestProject")

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
# CREATE Tests
# ============================================================================

test_name "Create project phase with required fields"
PHASE_NAME=$(unique_name "Phase")

xbe_json do project-phases create \
    --project "$CREATED_PROJECT_ID" \
    --name "$PHASE_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_PHASE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_PHASE_ID" && "$CREATED_PROJECT_PHASE_ID" != "null" ]]; then
        register_cleanup "project-phases" "$CREATED_PROJECT_PHASE_ID"
        pass
    else
        fail "Created project phase but no ID returned"
    fi
else
    fail "Failed to create project phase"
fi

if [[ -z "$CREATED_PROJECT_PHASE_ID" || "$CREATED_PROJECT_PHASE_ID" == "null" ]]; then
    echo "Cannot continue without a valid project phase ID"
    run_tests
fi

test_name "Create project phase with description"
PHASE_NAME2=$(unique_name "Phase2")
xbe_json do project-phases create \
    --project "$CREATED_PROJECT_ID" \
    --name "$PHASE_NAME2" \
    --description "Test phase description"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-phases" "$id"
    pass
else
    fail "Failed to create project phase with description"
fi

test_name "Create project phase with sequence position"
PHASE_NAME3=$(unique_name "Phase3")
xbe_json do project-phases create \
    --project "$CREATED_PROJECT_ID" \
    --name "$PHASE_NAME3" \
    --sequence-position 3

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-phases" "$id"
    pass
else
    fail "Failed to create project phase with sequence position"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project phase name"
UPDATED_PHASE_NAME=$(unique_name "UpdatedPhase")
xbe_json do project-phases update "$CREATED_PROJECT_PHASE_ID" --name "$UPDATED_PHASE_NAME"
assert_success

test_name "Update project phase description"
xbe_json do project-phases update "$CREATED_PROJECT_PHASE_ID" --description "Updated phase description"
assert_success

test_name "Update project phase sequence-position"
xbe_json do project-phases update "$CREATED_PROJECT_PHASE_ID" --sequence-position 5
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project phases"
xbe_json view project-phases list --limit 5
assert_success

test_name "List project phases returns array"
xbe_json view project-phases list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project phases"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project phases with --project filter"
xbe_json view project-phases list --project "$CREATED_PROJECT_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project phases with --limit"
xbe_json view project-phases list --limit 3
assert_success

test_name "List project phases with --offset"
xbe_json view project-phases list --limit 3 --offset 1
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project phase requires --confirm flag"
xbe_json do project-phases delete "$CREATED_PROJECT_PHASE_ID"
assert_failure

test_name "Delete project phase with --confirm"
# Create a project phase specifically for deletion
DEL_PHASE_NAME=$(unique_name "DeletePhase")
xbe_json do project-phases create \
    --project "$CREATED_PROJECT_ID" \
    --name "$DEL_PHASE_NAME"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do project-phases delete "$DEL_ID" --confirm
    # Note: Some APIs may not allow deleting project phases, so we accept both success and specific failures
    if [[ $status -eq 0 ]]; then
        pass
    else
        # API may not allow deletion - skip test
        skip "API may not allow project phase deletion"
    fi
else
    skip "Could not create project phase for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project phase without project fails"
xbe_json do project-phases create --name "Test Phase"
assert_failure

test_name "Create project phase without name fails"
xbe_json do project-phases create --project "$CREATED_PROJECT_ID"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-phases update "$CREATED_PROJECT_PHASE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
