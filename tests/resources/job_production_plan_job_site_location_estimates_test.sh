#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Job Site Location Estimates
#
# Tests view operations for the job production plan job site location estimates resource.
#
# COVERAGE: List + filters + show (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""

describe "Resource: job-production-plan-job-site-location-estimates (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job site location estimates"
xbe_json view job-production-plan-job-site-location-estimates list --limit 5
assert_success

test_name "List job site location estimates returns array"
xbe_json view job-production-plan-job-site-location-estimates list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list job site location estimates"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample job site location estimate"
xbe_json view job-production-plan-job-site-location-estimates list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No job site location estimates available for follow-on tests"
    fi
else
    skip "Could not list job site location estimates to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List job site location estimates with --job-production-plan filter"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view job-production-plan-job-site-location-estimates list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No sample job production plan ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job site location estimate"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-job-site-location-estimates show "$SAMPLE_ID"
    assert_success
else
    skip "No job site location estimate ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
