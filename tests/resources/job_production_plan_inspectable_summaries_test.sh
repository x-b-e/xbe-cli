#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Inspectable Summaries
#
# Tests list and show operations for the job-production-plan-inspectable-summaries resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUMMARY_ID=""
DEVELOPER_ID=""
PROJECT_ID=""
JOB_NUMBER=""
START_ON=""
SKIP_ID_FILTERS=0

describe "Resource: job-production-plan-inspectable-summaries"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List inspectable summaries"
xbe_json view job-production-plan-inspectable-summaries list --limit 5
assert_success

test_name "List inspectable summaries returns array"
xbe_json view job-production-plan-inspectable-summaries list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list inspectable summaries"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample inspectable summary"
xbe_json view job-production-plan-inspectable-summaries list --limit 1
if [[ $status -eq 0 ]]; then
    SUMMARY_ID=$(json_get ".[0].id")
    if [[ -n "$SUMMARY_ID" && "$SUMMARY_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No inspectable summaries available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list inspectable summaries"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show inspectable summary"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SUMMARY_ID" && "$SUMMARY_ID" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries show "$SUMMARY_ID"
    assert_success
    if [[ $status -eq 0 ]]; then
        DEVELOPER_ID=$(json_get ".developer_id")
        PROJECT_ID=$(json_get ".project_id")
        JOB_NUMBER=$(json_get ".job_number")
        START_ON=$(json_get ".start_on")
    fi
else
    skip "No inspectable summary ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List inspectable summaries with --developer filter"
if [[ -n "$DEVELOPER_ID" && "$DEVELOPER_ID" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --developer "$DEVELOPER_ID" --limit 5
    assert_success
else
    skip "No developer ID available"
fi

test_name "List inspectable summaries with --project filter"
if [[ -n "$PROJECT_ID" && "$PROJECT_ID" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

test_name "List inspectable summaries with --job-number filter"
if [[ -n "$JOB_NUMBER" && "$JOB_NUMBER" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --job-number "$JOB_NUMBER" --limit 5
    assert_success
else
    skip "No job number available"
fi

test_name "List inspectable summaries with --start-on filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --start-on "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

test_name "List inspectable summaries with --start-on-min filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --start-on-min "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

test_name "List inspectable summaries with --start-on-max filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view job-production-plan-inspectable-summaries list --start-on-max "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
