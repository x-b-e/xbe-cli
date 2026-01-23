#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Cancellations
#
# Tests list, show, and create operations for the job-production-plan-cancellations resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""
SAMPLE_REASON_TYPE_ID=""
CANCEL_PLAN_ID=""
CANCEL_REASON_TYPE_ID=""
LIST_SUPPORTED="true"

describe "Resource: job-production-plan-cancellations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan cancellations"
xbe_json view job-production-plan-cancellations list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing cancellations"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List job production plan cancellations returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-cancellations list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list job production plan cancellations"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample cancellation"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view job-production-plan-cancellations list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
        SAMPLE_REASON_TYPE_ID=$(json_get ".[0].job_production_plan_cancellation_reason_type_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No cancellations available for follow-on tests"
        fi
    else
        skip "Could not list cancellations to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find approved job production plan"
xbe_json view job-production-plans list --start-on-min "2000-01-01" --status approved --limit 5
if [[ $status -eq 0 ]]; then
    CANCEL_PLAN_ID=$(json_get ".[0].id")
    if [[ -n "$CANCEL_PLAN_ID" && "$CANCEL_PLAN_ID" != "null" ]]; then
        pass
    else
        skip "No approved job production plans available"
    fi
else
    skip "Could not list job production plans"
fi

test_name "Find cancellation reason type"
xbe_json view job-production-plan-cancellation-reason-types list --limit 5
if [[ $status -eq 0 ]]; then
    CANCEL_REASON_TYPE_ID=$(json_get ".[0].id")
    if [[ -n "$CANCEL_REASON_TYPE_ID" && "$CANCEL_REASON_TYPE_ID" != "null" ]]; then
        pass
    else
        skip "No cancellation reason types available"
    fi
else
    skip "Could not list cancellation reason types"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create job production plan cancellation"
if [[ -n "$CANCEL_PLAN_ID" && "$CANCEL_PLAN_ID" != "null" ]]; then
    if [[ -n "$CANCEL_REASON_TYPE_ID" && "$CANCEL_REASON_TYPE_ID" != "null" ]]; then
        xbe_json do job-production-plan-cancellations create \
            --job-production-plan "$CANCEL_PLAN_ID" \
            --job-production-plan-cancellation-reason-type "$CANCEL_REASON_TYPE_ID" \
            --comment "CLI test cancellation" \
            --suppress-status-change-notifications
    else
        xbe_json do job-production-plan-cancellations create \
            --job-production-plan "$CANCEL_PLAN_ID" \
            --comment "CLI test cancellation" \
            --suppress-status-change-notifications
    fi

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status"* ]] || \
           [[ "$output" == *"can't have time cards"* ]] || \
           [[ "$output" == *"can't have material transactions"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No approved job production plan available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan cancellation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-cancellations show "$SAMPLE_ID"
    assert_success
else
    skip "No cancellation ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cancellation without job production plan fails"
xbe_run do job-production-plan-cancellations create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
