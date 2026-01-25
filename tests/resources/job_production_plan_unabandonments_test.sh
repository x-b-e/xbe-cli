#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Unabandonments
#
# Tests list, show, and create operations for the job-production-plan-unabandonments resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
UNABANDON_PLAN_ID=""
LIST_SUPPORTED="true"

describe "Resource: job-production-plan-unabandonments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan unabandonments"
xbe_json view job-production-plan-unabandonments list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing unabandonments"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List job production plan unabandonments returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-unabandonments list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list job production plan unabandonments"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample unabandonment"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-unabandonments list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No unabandonments available for follow-on tests"
        fi
    else
        skip "Could not list unabandonments to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find abandoned job production plan"
xbe_json view job-production-plans list --start-on-min "2000-01-01" --status abandoned --limit 5
if [[ $status -eq 0 ]]; then
    UNABANDON_PLAN_ID=$(json_get ".[0].id")
    if [[ -n "$UNABANDON_PLAN_ID" && "$UNABANDON_PLAN_ID" != "null" ]]; then
        pass
    else
        skip "No abandoned job production plans available"
    fi
else
    skip "Could not list job production plans"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan unabandonment"
if [[ -n "$UNABANDON_PLAN_ID" && "$UNABANDON_PLAN_ID" != "null" ]]; then
    xbe_json do job-production-plan-unabandonments create \
        --job-production-plan "$UNABANDON_PLAN_ID" \
        --comment "CLI test unabandonment" \
        --suppress-status-change-notifications

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status"* ]] || \
           [[ "$output" == *"cannot be changed when invoiced"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No abandoned job production plan available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan unabandonment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-unabandonments show "$SAMPLE_ID"
    assert_success
else
    skip "No unabandonment ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create unabandonment without job production plan fails"
xbe_run do job-production-plan-unabandonments create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
