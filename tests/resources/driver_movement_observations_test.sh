#!/bin/bash
#
# XBE CLI Integration Tests: Driver Movement Observations
#
# Tests list and show operations for the driver-movement-observations resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""

describe "Resource: driver-movement-observations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver movement observations"
xbe_json view driver-movement-observations list --limit 5
assert_success

test_name "List driver movement observations returns array"
xbe_json view driver-movement-observations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver movement observations"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample observation"
xbe_json view driver-movement-observations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PLAN_ID=$(json_get ".[0].plan_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No observations available for follow-on tests"
    fi
else
    skip "Could not list observations to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List observations with --plan filter"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view driver-movement-observations list --plan "$SAMPLE_PLAN_ID" --limit 5
    assert_success
else
    skip "No plan ID available"
fi

test_name "List observations with --is-current filter"
xbe_json view driver-movement-observations list --is-current --limit 5
assert_success

test_name "List observations with --created-at-min filter"
xbe_json view driver-movement-observations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List observations with --created-at-max filter"
xbe_json view driver-movement-observations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List observations with --updated-at-min filter"
xbe_json view driver-movement-observations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List observations with --updated-at-max filter"
xbe_json view driver-movement-observations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver movement observation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-movement-observations show "$SAMPLE_ID"
    assert_success
else
    skip "No observation ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
