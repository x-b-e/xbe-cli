#!/bin/bash
#
# XBE CLI Integration Tests: Project Transport Plan Strategy Steps
#
# Tests list and show operations for the project_transport_plan_strategy_steps resource.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_POSITION=""
SAMPLE_STRATEGY_ID=""
SAMPLE_EVENT_TYPE_ID=""

describe "Resource: project-transport-plan-strategy-steps"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List project transport plan strategy steps"
xbe_json view project-transport-plan-strategy-steps list --limit 5
assert_success

test_name "List project transport plan strategy steps returns array"
xbe_json view project-transport-plan-strategy-steps list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list project transport plan strategy steps"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample project transport plan strategy step"
xbe_json view project-transport-plan-strategy-steps list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_POSITION=$(json_get ".[0].position")
    SAMPLE_STRATEGY_ID=$(json_get ".[0].strategy_id")
    SAMPLE_EVENT_TYPE_ID=$(json_get ".[0].event_type_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No project transport plan strategy steps available for follow-on tests"
    fi
else
    skip "Could not list project transport plan strategy steps to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List project transport plan strategy steps with --position filter"
if [[ -n "$SAMPLE_POSITION" && "$SAMPLE_POSITION" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-steps list --position "$SAMPLE_POSITION" --limit 5
    assert_success
else
    skip "No position available"
fi

test_name "List project transport plan strategy steps with --strategy filter"
if [[ -n "$SAMPLE_STRATEGY_ID" && "$SAMPLE_STRATEGY_ID" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-steps list --strategy "$SAMPLE_STRATEGY_ID" --limit 5
    assert_success
else
    skip "No strategy ID available"
fi

test_name "List project transport plan strategy steps with --event-type filter"
if [[ -n "$SAMPLE_EVENT_TYPE_ID" && "$SAMPLE_EVENT_TYPE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-steps list --event-type "$SAMPLE_EVENT_TYPE_ID" --limit 5
    assert_success
else
    skip "No event type ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show project transport plan strategy step"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view project-transport-plan-strategy-steps show "$SAMPLE_ID"
    assert_success
else
    skip "No project transport plan strategy step ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
