#!/bin/bash
#
# XBE CLI Integration Tests: Project Categories
#
# Tests create and list operations for the project_categories resource.
# Project categories classify projects by type.
#
# COVERAGE: Create + list filters (create-only, no update/delete)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_PC_ID=""
CREATED_BROKER_ID=""

describe "Resource: project-categories"

# ============================================================================
# Prerequisites - Create broker for project categories
# ============================================================================

test_name "Create prerequisite broker for project categories tests"
BROKER_NAME=$(unique_name "PCTestBroker")

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

test_name "Create project category with required fields"
TEST_NAME=$(unique_name "ProjCat")
TEST_ABBREV=$(echo "$TEST_NAME" | cut -c1-8)
xbe_json do project-categories create --name "$TEST_NAME" --abbreviation "$TEST_ABBREV" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_PC_ID=$(json_get ".id")
    if [[ -n "$CREATED_PC_ID" && "$CREATED_PC_ID" != "null" ]]; then
        # Note: No delete available for project-categories
        pass
    else
        fail "Created project category but no ID returned"
    fi
else
    fail "Failed to create project category: $output"
fi

test_name "Create project category with abbreviation"
TEST_NAME2=$(unique_name "ProjCat2")
xbe_json do project-categories create \
    --name "$TEST_NAME2" \
    --abbreviation "PC2" \
    --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    pass
else
    fail "Failed to create project category with abbreviation"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project categories"
xbe_json view project-categories list
assert_success

test_name "List project categories returns array"
xbe_json view project-categories list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project categories"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project categories with --name filter"
xbe_json view project-categories list --name "commercial"
assert_success

test_name "List project categories with --q filter"
xbe_json view project-categories list --q "road"
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project categories with --limit"
xbe_json view project-categories list --limit 5
assert_success

test_name "List project categories with --offset"
xbe_json view project-categories list --limit 5 --offset 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project category without name fails"
xbe_json do project-categories create --broker "$CREATED_BROKER_ID"
assert_failure

test_name "Create project category without broker fails"
xbe_json do project-categories create --name "Missing Broker"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
