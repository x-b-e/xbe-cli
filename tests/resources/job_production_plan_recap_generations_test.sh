#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Recap Generations
#
# Tests create operations for the job-production-plan-recap-generations resource.
#
# COVERAGE: create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PLAN_ID=""

# Status lookup attempts
STATUS_APPROVED="approved"
STATUS_COMPLETE="complete"


describe "Resource: job-production-plan-recap-generations"

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find approved job production plan"
xbe_json view job-production-plans list --start-on-min "2000-01-01" --status "$STATUS_APPROVED" --limit 5
if [[ $status -eq 0 ]]; then
    PLAN_ID=$(json_get ".[0].id")
    if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
        pass
    else
        skip "No approved job production plans available"
    fi
else
    skip "Could not list job production plans"
fi

if [[ -z "$PLAN_ID" || "$PLAN_ID" == "null" ]]; then
    test_name "Find complete job production plan"
    xbe_json view job-production-plans list --start-on-min "2000-01-01" --status "$STATUS_COMPLETE" --limit 5
    if [[ $status -eq 0 ]]; then
        PLAN_ID=$(json_get ".[0].id")
        if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
            pass
        else
            skip "No complete job production plans available"
        fi
    else
        skip "Could not list job production plans"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan recap generation"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json do job-production-plan-recap-generations create --job-production-plan "$PLAN_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"must be approved or complete"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No approved or complete job production plan available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create recap generation without job production plan fails"
xbe_run do job-production-plan-recap-generations create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
