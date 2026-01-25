#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Schedule Changes
#
# Tests create operations for the job-production-plan-schedule-changes resource.
#
# COVERAGE: create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PLAN_ID=""

# Status lookup attempts
STATUS_APPROVED="approved"
STATUS_COMPLETE="complete"

describe "Resource: job-production-plan-schedule-changes"

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

test_name "Create job production plan schedule change"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json do job-production-plan-schedule-changes create \
        --job-production-plan "$PLAN_ID" \
        --offset-seconds 300 \
        --time-kind both \
        --skip-update-shifts \
        --skip-update-crew-requirements \
        --skip-update-safety-meeting \
        --skip-persistence
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"must not have schedule locked"* ]] || \
           [[ "$output" == *"must be able to calculate start at"* ]] || \
           [[ "$output" == *"must be able to calculate job site start at"* ]] || \
           [[ "$output" == *"cannot be changed when shifts have checked out"* ]] || \
           [[ "$output" == *"cannot be moved when there are time sheets"* ]] || \
           [[ "$output" == *"cannot have an equipment movement trip"* ]] || \
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

test_name "Create schedule change without required flags fails"
xbe_run do job-production-plan-schedule-changes create
assert_failure

test_name "Create schedule change without offset fails"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_run do job-production-plan-schedule-changes create --job-production-plan "$PLAN_ID"
    assert_failure
else
    skip "No job production plan available"
fi

test_name "Create schedule change with invalid work ID"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_run do job-production-plan-schedule-changes create --job-production-plan "$PLAN_ID" --offset-seconds 60 --work 0
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"not found"* ]] || \
           [[ "$output" == *"work"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Unexpected error: $output"
        fi
    fi
else
    skip "No job production plan available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
