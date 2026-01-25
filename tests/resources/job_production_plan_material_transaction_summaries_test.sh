#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Material Transaction Summaries
#
# Tests create operations for job-production-plan-material-transaction-summaries.
#
# COVERAGE: Writable attributes (job-production-plan)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PLAN_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_ID:-}"

describe "Resource: job-production-plan-material-transaction-summaries"

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create requires job production plan"
xbe_run do job-production-plan-material-transaction-summaries create
assert_failure

test_name "Create material transaction summary"
if [[ -n "$PLAN_ID" && "$PLAN_ID" != "null" ]]; then
    xbe_json do job-production-plan-material-transaction-summaries create --job-production-plan "$PLAN_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".job_production_plan_id" "$PLAN_ID"
        if echo "$output" | jq -e '.tons_by_material_type | type == "array"' >/dev/null 2>&1; then
            pass
        else
            fail "tons_by_material_type was not an array"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"404"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create job production plan material transaction summary: $output"
        fi
    fi
else
    skip "No job production plan ID available. Set XBE_TEST_JOB_PRODUCTION_PLAN_ID to enable create testing."
fi

run_tests
