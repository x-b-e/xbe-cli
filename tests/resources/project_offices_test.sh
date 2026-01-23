#!/bin/bash
#
# XBE CLI Integration Tests: Project Offices
#
# Tests CRUD operations for the project_offices resource.
# Project offices define branch offices or regional divisions.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PROJECT_OFFICE_ID=""
CREATED_BROKER_ID=""

describe "Resource: project_offices"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for project office tests"
BROKER_NAME=$(unique_name "POTestBroker")

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

test_name "Create project office with required fields"
TEST_NAME=$(unique_name "ProjOffice")

xbe_json do project-offices create \
    --name "$TEST_NAME" \
    --abbreviation "PO1" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PROJECT_OFFICE_ID=$(json_get ".id")
    if [[ -n "$CREATED_PROJECT_OFFICE_ID" && "$CREATED_PROJECT_OFFICE_ID" != "null" ]]; then
        register_cleanup "project-offices" "$CREATED_PROJECT_OFFICE_ID"
        pass
    else
        fail "Created project office but no ID returned"
    fi
else
    fail "Failed to create project office"
fi

# Only continue if we successfully created a project office
if [[ -z "$CREATED_PROJECT_OFFICE_ID" || "$CREATED_PROJECT_OFFICE_ID" == "null" ]]; then
    echo "Cannot continue without a valid project office ID"
    run_tests
fi

test_name "Create project office with abbreviation"
TEST_NAME2=$(unique_name "ProjOffice2")
xbe_json do project-offices create \
    --name "$TEST_NAME2" \
    --abbreviation "PO2" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-offices" "$id"
    pass
else
    fail "Failed to create project office with abbreviation"
fi

test_name "Create project office with is-active=false"
TEST_NAME3=$(unique_name "ProjOffice3")
xbe_json do project-offices create \
    --name "$TEST_NAME3" \
    --abbreviation "PO3" \
    --broker "$CREATED_BROKER_ID" \
    --is-active=false
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-offices" "$id"
    pass
else
    fail "Failed to create project office with is-active=false"
fi

test_name "Create project office with all optional fields"
TEST_NAME4=$(unique_name "ProjOffice4")
xbe_json do project-offices create \
    --name "$TEST_NAME4" \
    --abbreviation "PO4" \
    --broker "$CREATED_BROKER_ID" \
    --is-active=true
if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "project-offices" "$id"
    pass
else
    fail "Failed to create project office with all optional fields"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update project office name"
UPDATED_NAME=$(unique_name "UpdatedPO")
xbe_json do project-offices update "$CREATED_PROJECT_OFFICE_ID" --name "$UPDATED_NAME"
assert_success

test_name "Update project office abbreviation"
xbe_json do project-offices update "$CREATED_PROJECT_OFFICE_ID" --abbreviation "UPDPO"
assert_success

test_name "Update project office is-active"
xbe_json do project-offices update "$CREATED_PROJECT_OFFICE_ID" --is-active=false
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project offices"
xbe_json view project-offices list --limit 5
assert_success

test_name "List project offices returns array"
xbe_json view project-offices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project offices"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project offices with --broker filter"
xbe_json view project-offices list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project offices with --limit"
xbe_json view project-offices list --limit 3
assert_success

test_name "List project offices with --offset"
xbe_json view project-offices list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete project office requires --confirm flag"
xbe_run do project-offices delete "$CREATED_PROJECT_OFFICE_ID"
assert_failure

test_name "Delete project office with --confirm"
# Create a project office specifically for deletion
TEST_DEL_NAME=$(unique_name "DeletePO")
xbe_json do project-offices create \
    --name "$TEST_DEL_NAME" \
    --abbreviation "DPO" \
    --broker "$CREATED_BROKER_ID"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do project-offices delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create project office for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project office without name fails"
xbe_json do project-offices create --abbreviation "TST" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project office without abbreviation fails"
xbe_json do project-offices create --name "NoAbbrev" --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project office without broker fails"
xbe_json do project-offices create --name "NoBroker" --abbreviation "NB"
assert_failure

test_name "Update without any fields fails"
xbe_json do project-offices update "$CREATED_PROJECT_OFFICE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
