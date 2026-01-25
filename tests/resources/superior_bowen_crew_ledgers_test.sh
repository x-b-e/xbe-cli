#!/bin/bash
#
# XBE CLI Integration Tests: Superior Bowen Crew Ledgers
#
# Tests create operations for superior-bowen-crew-ledgers.
#
# COVERAGE: job-production-plan relationship + missing job production plan + invalid ID failure
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

JOB_PRODUCTION_PLAN_ID="${XBE_TEST_SUPERIOR_BOWEN_CREW_LEDGER_JOB_PRODUCTION_PLAN_ID:-}"
CREATED_LEDGER_ID=""

INVALID_JOB_PRODUCTION_PLAN_ID="invalid-jpp-$(date +%s)"

describe "Resource: superior-bowen-crew-ledgers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create Superior Bowen crew ledger requires --job-production-plan"
xbe_run do superior-bowen-crew-ledgers create
assert_failure

test_name "Create Superior Bowen crew ledger rejects invalid job production plan ID"
xbe_run do superior-bowen-crew-ledgers create --job-production-plan "$INVALID_JOB_PRODUCTION_PLAN_ID"
if [[ $status -eq 0 ]]; then
    fail "Expected failure for invalid job production plan ID"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Not authorized to create Superior Bowen crew ledgers"
    else
        pass
    fi
fi

test_name "Create Superior Bowen crew ledger"
if [[ -n "$JOB_PRODUCTION_PLAN_ID" && "$JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json do superior-bowen-crew-ledgers create --job-production-plan "$JOB_PRODUCTION_PLAN_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_LEDGER_ID=$(json_get ".id")
        if [[ -n "$CREATED_LEDGER_ID" && "$CREATED_LEDGER_ID" != "null" ]]; then
            returned_job_production_plan_id=$(json_get ".job_production_plan_id")
            if [[ "$returned_job_production_plan_id" == "$JOB_PRODUCTION_PLAN_ID" ]]; then
                pass
            else
                fail "Job production plan ID mismatch"
            fi
        else
            fail "Created Superior Bowen crew ledger but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
            skip "Not authorized to create Superior Bowen crew ledgers"
        elif [[ "$output" == *"unprocessable"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Job production plan missing required Superior Bowen fields"
        elif [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]] || [[ "$output" == *"404"* ]]; then
            skip "Job production plan not found"
        else
            fail "Failed to create Superior Bowen crew ledger: $output"
        fi
    fi
else
    skip "Missing job production plan ID (set XBE_TEST_SUPERIOR_BOWEN_CREW_LEDGER_JOB_PRODUCTION_PLAN_ID)"
fi

run_tests
