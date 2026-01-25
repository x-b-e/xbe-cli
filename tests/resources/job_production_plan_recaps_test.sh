#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Recaps
#
# Tests list and show operations for the job-production-plan-recaps resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PLAN_ID=""

describe "Resource: job-production-plan-recaps"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan recaps"
xbe_json view job-production-plan-recaps list --limit 5
assert_success

test_name "List job production plan recaps returns array"
xbe_json view job-production-plan-recaps list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job production plan recaps"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample recap"
xbe_json view job-production-plan-recaps list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PLAN_ID=$(json_get ".[0].plan_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No recaps available for follow-on tests"
    fi
else
    skip "Could not list recaps to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List recaps with --plan filter"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json view job-production-plan-recaps list --plan "$SAMPLE_PLAN_ID" --limit 5
    assert_success
else
    skip "No plan ID available"
fi

test_name "List recaps with --created-at-min filter"
xbe_json view job-production-plan-recaps list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recaps with --created-at-max filter"
xbe_json view job-production-plan-recaps list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recaps with --updated-at-min filter"
xbe_json view job-production-plan-recaps list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List recaps with --updated-at-max filter"
xbe_json view job-production-plan-recaps list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan recap"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-recaps show "$SAMPLE_ID"
    assert_success
else
    skip "No recap ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
