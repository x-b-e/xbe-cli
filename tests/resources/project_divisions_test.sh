#!/bin/bash
#
# XBE CLI Integration Tests: Project Divisions
#
# Tests create and list operations for the project_divisions resource.
# Project divisions are organizational units for projects.
#
# COVERAGE: Create + list filters (create-only, no update/delete)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PD_ID=""
CREATED_BROKER_ID=""

describe "Resource: project-divisions"

# ============================================================================
# Prerequisites - Create broker for project divisions
# ============================================================================

test_name "Create prerequisite broker for project divisions tests"
BROKER_NAME=$(unique_name "PDTestBroker")

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

test_name "Create project division with required fields"
TEST_NAME=$(unique_name "ProjDiv")
TEST_ABBREV=$(echo "$TEST_NAME" | cut -c1-8)
xbe_json do project-divisions create --name "$TEST_NAME" --abbreviation "$TEST_ABBREV" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PD_ID=$(json_get ".id")
    if [[ -n "$CREATED_PD_ID" && "$CREATED_PD_ID" != "null" ]]; then
        # Note: No delete available for project-divisions
        pass
    else
        fail "Created project division but no ID returned"
    fi
else
    fail "Failed to create project division: $output"
fi

test_name "Create project division with abbreviation"
TEST_NAME2=$(unique_name "ProjDiv2")
xbe_json do project-divisions create \
    --name "$TEST_NAME2" \
    --abbreviation "PD2" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create project division with abbreviation"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project divisions"
xbe_json view project-divisions list
assert_success

test_name "List project divisions returns array"
xbe_json view project-divisions list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project divisions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project divisions with --name filter"
xbe_json view project-divisions list --name "north"
assert_success

test_name "List project divisions with --q filter"
xbe_json view project-divisions list --q "region"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project divisions with --limit"
xbe_json view project-divisions list --limit 5
assert_success

test_name "List project divisions with --offset"
xbe_json view project-divisions list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project division without name fails"
xbe_json do project-divisions create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project division without broker fails"
xbe_json do project-divisions create --name "Missing Broker"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
