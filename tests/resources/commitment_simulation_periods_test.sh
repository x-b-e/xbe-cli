#!/bin/bash
#
# XBE CLI Integration Tests: Commitment Simulation Periods
#
# Tests view operations for the commitment-simulation-periods resource.
#
# COVERAGE: List + filters + pagination + show (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_COMMITMENT_SIMULATION_ID=""
COMMITMENT_SIMULATION_ID="${XBE_TEST_COMMITMENT_SIMULATION_ID:-}"

describe "Resource: commitment-simulation-periods (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List commitment simulation periods"
xbe_json view commitment-simulation-periods list --limit 5
assert_success

test_name "List commitment simulation periods returns array"
xbe_json view commitment-simulation-periods list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list commitment simulation periods"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample commitment simulation period"
xbe_json view commitment-simulation-periods list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_COMMITMENT_SIMULATION_ID=$(json_get ".[0].commitment_simulation_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No commitment simulation periods available for follow-on tests"
    fi
else
    skip "Could not list commitment simulation periods to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List commitment simulation periods with --commitment-simulation filter"
FILTER_ID="$SAMPLE_COMMITMENT_SIMULATION_ID"
if [[ -z "$FILTER_ID" || "$FILTER_ID" == "null" ]]; then
    FILTER_ID="$COMMITMENT_SIMULATION_ID"
fi
if [[ -n "$FILTER_ID" && "$FILTER_ID" != "null" ]]; then
    xbe_json view commitment-simulation-periods list --commitment-simulation "$FILTER_ID" --limit 5
    assert_success
else
    skip "No commitment simulation ID available (set XBE_TEST_COMMITMENT_SIMULATION_ID to enable)"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List commitment simulation periods with --limit"
xbe_json view commitment-simulation-periods list --limit 2
assert_success

test_name "List commitment simulation periods with --offset"
xbe_json view commitment-simulation-periods list --limit 2 --offset 2
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show commitment simulation period"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view commitment-simulation-periods show "$SAMPLE_ID"
    assert_success
else
    skip "No commitment simulation period ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
