#!/bin/bash
#
# XBE CLI Integration Tests: Driver Movement Segment Sets
#
# Tests list and show operations for the driver_movement_segment_sets resource.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_DRIVER_DAY_ID=""
SAMPLE_DRIVER_ID=""

describe "Resource: driver-movement-segment-sets"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver movement segment sets"
xbe_json view driver-movement-segment-sets list --limit 5
assert_success

test_name "List driver movement segment sets returns array"
xbe_json view driver-movement-segment-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver movement segment sets"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample driver movement segment set"
xbe_json view driver-movement-segment-sets list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No driver movement segment sets available for follow-on tests"
    fi
else
    skip "Could not list driver movement segment sets to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List driver movement segment sets with --driver-day filter"
if [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    xbe_json view driver-movement-segment-sets list --driver-day "$SAMPLE_DRIVER_DAY_ID" --limit 5
    assert_success
else
    skip "No driver day ID available"
fi

test_name "List driver movement segment sets with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view driver-movement-segment-sets list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver movement segment set"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-movement-segment-sets show "$SAMPLE_ID"
    assert_success
else
    skip "No driver movement segment set ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
