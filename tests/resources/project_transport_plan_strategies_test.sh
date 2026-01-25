#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Strategies
#
# Tests view behavior for project-transport-plan-strategies.
#
# COVERAGE: List + list filters + show
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: project-transport-plan-strategies"

SAMPLE_ID=""
SAMPLE_NAME=""
SAMPLE_STEP_PATTERN=""

NAME_FILTER="${XBE_TEST_PROJECT_TRANSPORT_PLAN_STRATEGY_NAME:-}"
STEP_PATTERN_FILTER="${XBE_TEST_PROJECT_TRANSPORT_PLAN_STRATEGY_STEP_PATTERN:-}"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan strategies"
xbe_json view project-transport-plan-strategies list --limit 5
assert_success

test_name "List project transport plan strategies returns array"
xbe_json view project-transport-plan-strategies list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan strategies"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample strategy"
xbe_json view project-transport-plan-strategies list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_NAME=$(json_get ".[0].name")
    SAMPLE_STEP_PATTERN=$(json_get ".[0].step_pattern")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No strategies available for follow-on tests"
    fi
else
    skip "Could not list strategies to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List strategies with --name filter"
FILTER_NAME="$SAMPLE_NAME"
if [[ -z "$FILTER_NAME" || "$FILTER_NAME" == "null" ]]; then
    FILTER_NAME="$NAME_FILTER"
fi
if [[ -n "$FILTER_NAME" && "$FILTER_NAME" != "null" ]]; then
    xbe_json view project-transport-plan-strategies list --name "$FILTER_NAME" --limit 5
    assert_success
else
    skip "No strategy name available"
fi

test_name "List strategies with --step-pattern filter"
FILTER_STEP_PATTERN="$SAMPLE_STEP_PATTERN"
if [[ -z "$FILTER_STEP_PATTERN" || "$FILTER_STEP_PATTERN" == "null" ]]; then
    FILTER_STEP_PATTERN="$STEP_PATTERN_FILTER"
fi
if [[ -n "$FILTER_STEP_PATTERN" && "$FILTER_STEP_PATTERN" != "null" ]]; then
    xbe_json view project-transport-plan-strategies list --step-pattern "$FILTER_STEP_PATTERN" --limit 5
    assert_success
else
    skip "No step pattern available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan strategy details"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-strategies show "$SAMPLE_ID"
    assert_success
else
    skip "No strategy ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
