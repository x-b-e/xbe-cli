#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Lineups
#
# Tests list, show, create, and delete operations for lineup_scenario_lineups.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_LINEUP_SCENARIO_LINEUP_ID=""
LINEUP_SCENARIO_ID="${XBE_TEST_LINEUP_SCENARIO_ID:-}"
LINEUP_ID="${XBE_TEST_LINEUP_ID:-}"
CREATED_LINEUP_SCENARIO_LINEUP_ID=""

describe "Resource: lineup-scenario-lineups"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List lineup scenario lineups"
xbe_json view lineup-scenario-lineups list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_LINEUP_SCENARIO_LINEUP_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$LINEUP_SCENARIO_ID" || "$LINEUP_SCENARIO_ID" == "null" ]]; then
            LINEUP_SCENARIO_ID=$(echo "$output" | jq -r '.[0].lineup_scenario_id')
        fi
        if [[ -z "$LINEUP_ID" || "$LINEUP_ID" == "null" ]]; then
            LINEUP_ID=$(echo "$output" | jq -r '.[0].lineup_id')
        fi
    fi
else
    fail "Failed to list lineup scenario lineups"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario lineup"
if [[ -n "$SEED_LINEUP_SCENARIO_LINEUP_ID" && "$SEED_LINEUP_SCENARIO_LINEUP_ID" != "null" ]]; then
    xbe_json view lineup-scenario-lineups show "$SEED_LINEUP_SCENARIO_LINEUP_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        if [[ -z "$LINEUP_SCENARIO_ID" || "$LINEUP_SCENARIO_ID" == "null" ]]; then
            LINEUP_SCENARIO_ID=$(json_get ".lineup_scenario_id")
        fi
        if [[ -z "$LINEUP_ID" || "$LINEUP_ID" == "null" ]]; then
            LINEUP_ID=$(json_get ".lineup_id")
        fi
        pass
    else
        fail "Failed to show lineup scenario lineup"
    fi
else
    skip "No lineup scenario lineup available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create lineup scenario lineup"
if [[ -n "$LINEUP_SCENARIO_ID" && "$LINEUP_SCENARIO_ID" != "null" && -n "$LINEUP_ID" && "$LINEUP_ID" != "null" ]]; then
    xbe_json do lineup-scenario-lineups create --lineup-scenario "$LINEUP_SCENARIO_ID" --lineup "$LINEUP_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_LINEUP_SCENARIO_LINEUP_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINEUP_SCENARIO_LINEUP_ID" && "$CREATED_LINEUP_SCENARIO_LINEUP_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-lineups" "$CREATED_LINEUP_SCENARIO_LINEUP_ID"
            pass
        else
            fail "Created lineup scenario lineup but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario lineup"
    fi
else
    skip "No lineup scenario or lineup ID available for creation (set XBE_TEST_LINEUP_SCENARIO_ID and XBE_TEST_LINEUP_ID)"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete lineup scenario lineup"
if [[ -n "$CREATED_LINEUP_SCENARIO_LINEUP_ID" && "$CREATED_LINEUP_SCENARIO_LINEUP_ID" != "null" ]]; then
    xbe_run do lineup-scenario-lineups delete "$CREATED_LINEUP_SCENARIO_LINEUP_ID" --confirm
    assert_success
else
    skip "No created lineup scenario lineup to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by lineup scenario"
if [[ -n "$LINEUP_SCENARIO_ID" && "$LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json view lineup-scenario-lineups list --lineup-scenario "$LINEUP_SCENARIO_ID" --limit 5
    assert_success
else
    skip "No lineup scenario ID available for filter"
fi

test_name "Filter by lineup"
if [[ -n "$LINEUP_ID" && "$LINEUP_ID" != "null" ]]; then
    xbe_json view lineup-scenario-lineups list --lineup "$LINEUP_ID" --limit 5
    assert_success
else
    skip "No lineup ID available for filter"
fi

run_tests
