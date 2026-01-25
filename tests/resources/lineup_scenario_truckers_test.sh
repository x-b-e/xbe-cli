#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Truckers
#
# Tests list, show, create, update, and delete operations for lineup_scenario_truckers.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_LINEUP_SCENARIO_TRUCKER_ID=""
LINEUP_SCENARIO_ID="${XBE_TEST_LINEUP_SCENARIO_ID:-}"
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
CREATED_LINEUP_SCENARIO_TRUCKER_ID=""

describe "Resource: lineup-scenario-truckers"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List lineup scenario truckers"
xbe_json view lineup-scenario-truckers list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_LINEUP_SCENARIO_TRUCKER_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$LINEUP_SCENARIO_ID" || "$LINEUP_SCENARIO_ID" == "null" ]]; then
            LINEUP_SCENARIO_ID=$(echo "$output" | jq -r '.[0].lineup_scenario_id')
        fi
        if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
            TRUCKER_ID=$(echo "$output" | jq -r '.[0].trucker_id')
        fi
    fi
else
    fail "Failed to list lineup scenario truckers"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario trucker"
if [[ -n "$SEED_LINEUP_SCENARIO_TRUCKER_ID" && "$SEED_LINEUP_SCENARIO_TRUCKER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-truckers show "$SEED_LINEUP_SCENARIO_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        if [[ -z "$LINEUP_SCENARIO_ID" || "$LINEUP_SCENARIO_ID" == "null" ]]; then
            LINEUP_SCENARIO_ID=$(json_get ".lineup_scenario_id")
        fi
        if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
            TRUCKER_ID=$(json_get ".trucker_id")
        fi
        pass
    else
        fail "Failed to show lineup scenario trucker"
    fi
else
    skip "No lineup scenario trucker available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create lineup scenario trucker"
if [[ -n "$LINEUP_SCENARIO_ID" && "$LINEUP_SCENARIO_ID" != "null" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json do lineup-scenario-truckers create --lineup-scenario "$LINEUP_SCENARIO_ID" --trucker "$TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_LINEUP_SCENARIO_TRUCKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" && "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-truckers" "$CREATED_LINEUP_SCENARIO_TRUCKER_ID"
            pass
        else
            fail "Created lineup scenario trucker but no ID returned"
        fi
    else
        fail "Failed to create lineup scenario trucker"
    fi
else
    skip "No lineup scenario or trucker ID available for creation (set XBE_TEST_LINEUP_SCENARIO_ID and XBE_TEST_TRUCKER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update lineup scenario trucker attributes"
if [[ -n "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" && "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" != "null" ]]; then
    xbe_json do lineup-scenario-truckers update "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" \
        --minimum-assignment-count 1 \
        --maximum-assignment-count 2 \
        --maximum-minutes-to-start-site 45 \
        --material-type-constraints '[]' \
        --trailer-classification-constraints '[]'
    assert_success
else
    skip "No created lineup scenario trucker to update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete lineup scenario trucker"
if [[ -n "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" && "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" != "null" ]]; then
    xbe_run do lineup-scenario-truckers delete "$CREATED_LINEUP_SCENARIO_TRUCKER_ID" --confirm
    assert_success
else
    skip "No created lineup scenario trucker to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by lineup scenario"
if [[ -n "$LINEUP_SCENARIO_ID" && "$LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json view lineup-scenario-truckers list --lineup-scenario "$LINEUP_SCENARIO_ID" --limit 5
    assert_success
else
    skip "No lineup scenario ID available for filter"
fi

test_name "Filter by trucker"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-truckers list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available for filter"
fi

run_tests
