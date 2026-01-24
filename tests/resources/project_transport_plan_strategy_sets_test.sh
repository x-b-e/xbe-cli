#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Strategy Sets
#
# Tests view operations for the project_transport_plan_strategy_sets resource.
#
# COVERAGE: List + filters + show (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_STRATEGY_SET_ID=""
STRATEGY_PATTERN=""

describe "Resource: project-transport-plan-strategy-sets (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan strategy sets"
xbe_json view project-transport-plan-strategy-sets list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_STRATEGY_SET_ID=$(echo "$output" | jq -r '.[0].id')
        STRATEGY_PATTERN=$(echo "$output" | jq -r '.[0].strategy_pattern')
    fi
else
    fail "Failed to list project transport plan strategy sets"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan strategy set"
if [[ -n "$SEED_STRATEGY_SET_ID" && "$SEED_STRATEGY_SET_ID" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-sets show "$SEED_STRATEGY_SET_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show project transport plan strategy set"
    fi
else
    skip "No project transport plan strategy set available to show"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List project transport plan strategy sets with --limit"
xbe_json view project-transport-plan-strategy-sets list --limit 5
assert_success

test_name "List project transport plan strategy sets with --offset"
xbe_json view project-transport-plan-strategy-sets list --limit 5 --offset 1
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by strategy pattern"
if [[ -n "$STRATEGY_PATTERN" && "$STRATEGY_PATTERN" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-sets list --strategy-pattern "$STRATEGY_PATTERN" --limit 5
    assert_success
else
    skip "No strategy pattern available for filter"
fi

run_tests
