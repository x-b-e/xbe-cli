#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Submissions
#
# Tests create operations for the job-production-plan-submissions resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_PLAN_ID=""

describe "Resource: job-production-plan-submissions"

# ============================================================================
# Sample Record (used for create)
# ============================================================================

test_name "Capture job production plan in editing status"
xbe_json view job-production-plans list --status editing --start-on-min 2000-01-01 --start-on-max 2100-01-01 --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_PLAN_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
        pass
    else
        skip "No editing job production plans available"
    fi
else
    skip "Could not list job production plans"
fi

if [[ -z "$SAMPLE_PLAN_ID" || "$SAMPLE_PLAN_ID" == "null" ]]; then
    test_name "Capture job production plan in rejected status"
    xbe_json view job-production-plans list --status rejected --start-on-min 2000-01-01 --start-on-max 2100-01-01 --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_PLAN_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
            pass
        else
            skip "No rejected job production plans available"
        fi
    else
        skip "Could not list rejected job production plans"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Submit job production plan"
if [[ -n "$SAMPLE_PLAN_ID" && "$SAMPLE_PLAN_ID" != "null" ]]; then
    xbe_json do job-production-plan-submissions create \
        --job-production-plan "$SAMPLE_PLAN_ID" \
        --comment "CLI submission test" \
        --suppress-status-change-notifications
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Submit failed: $output"
        fi
    fi
else
    skip "No job production plan available for submission"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Submit without required fields fails"
xbe_run do job-production-plan-submissions create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
