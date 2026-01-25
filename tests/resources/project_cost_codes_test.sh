#!/bin/bash
#
# XBE CLI Integration Tests: Project Cost Codes
#
# Tests operations for the project-cost-codes resource.
# Project cost codes require a project-customer and cost-code relationship which
# is complex to set up. This test focuses on list operations and error cases.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""

describe "Resource: project-cost-codes"

# ============================================================================
# Prerequisites - Create a broker for filtering tests
# ============================================================================

test_name "Create prerequisite broker for project cost code tests"
BROKER_NAME=$(unique_name "PCCTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project cost codes"
xbe_json view project-cost-codes list --limit 5
assert_success

test_name "List project cost codes returns array"
xbe_json view project-cost-codes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project cost codes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project cost codes with --query filter"
xbe_json view project-cost-codes list --query "test" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project cost codes with --limit"
xbe_json view project-cost-codes list --limit 3
assert_success

test_name "List project cost codes with --offset"
xbe_json view project-cost-codes list --limit 3 --offset 1
assert_success

test_name "List project cost codes with pagination (limit + offset)"
xbe_json view project-cost-codes list --limit 5 --offset 10
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create project cost code without project-customer fails"
xbe_json do project-cost-codes create --cost-code "1"
assert_failure

test_name "Create project cost code without cost-code fails"
xbe_json do project-cost-codes create --project-customer "1"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
