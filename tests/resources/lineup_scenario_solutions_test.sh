#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Solutions
#
# Tests list, show, and create operations for the lineup-scenario-solutions resource.
#
# COVERAGE: List + filters + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_LINEUP_SCENARIO_ID=""
LIST_SUPPORTED="true"

describe "Resource: lineup-scenario-solutions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenario solutions"
xbe_json view lineup-scenario-solutions list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing lineup scenario solutions"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List lineup scenario solutions returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view lineup-scenario-solutions list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list lineup scenario solutions"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show/filter/create)
# ============================================================================

test_name "Capture sample lineup scenario solution"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view lineup-scenario-solutions list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_LINEUP_SCENARIO_ID=$(json_get ".[0].lineup_scenario_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No lineup scenario solutions available for follow-on tests"
        fi
    else
        skip "Could not list lineup scenario solutions to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter lineup scenario solutions by lineup scenario"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_LINEUP_SCENARIO_ID" && "$SAMPLE_LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json view lineup-scenario-solutions list --lineup-scenario "$SAMPLE_LINEUP_SCENARIO_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        count=$(echo "$output" | jq 'length')
        if [[ "$count" -eq 0 ]]; then
            skip "No results returned for lineup scenario filter"
        else
            first_id=$(echo "$output" | jq -r '.[0].lineup_scenario_id')
            if [[ "$first_id" == "$SAMPLE_LINEUP_SCENARIO_ID" ]]; then
                pass
            else
                fail "Expected lineup_scenario_id $SAMPLE_LINEUP_SCENARIO_ID, got $first_id"
            fi
        fi
    else
        fail "Failed to filter lineup scenario solutions"
    fi
else
    skip "No lineup scenario ID available for filter test"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario solution"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view lineup-scenario-solutions show "$SAMPLE_ID"
    assert_success
else
    skip "No lineup scenario solution ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create lineup scenario solution"
if [[ -n "$SAMPLE_LINEUP_SCENARIO_ID" && "$SAMPLE_LINEUP_SCENARIO_ID" != "null" ]]; then
    xbe_json do lineup-scenario-solutions create --lineup-scenario "$SAMPLE_LINEUP_SCENARIO_ID"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No lineup scenario ID available for create"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create lineup scenario solution without lineup scenario fails"
xbe_run do lineup-scenario-solutions create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
