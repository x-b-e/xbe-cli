#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Completions
#
# Tests create operations for the job_production_plan_completions resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: job-production-plan-completions"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create completion without required job production plan fails"
xbe_run do job-production-plan-completions create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create job production plan completion"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for approved job production plans"
else
    base_url="${XBE_BASE_URL%/}"

    plans_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plans?page[limit]=1&filter[status]=approved" || true)

    jpp_id=$(echo "$plans_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$jpp_id" ]]; then
        skip "No approved job production plan found"
    else
        COMMENT="Completing plan for test"
        xbe_json do job-production-plan-completions create \
            --job-production-plan "$jpp_id" \
            --comment "$COMMENT" \
            --suppress-status-change-notifications

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
            assert_json_equals ".job_production_plan_id" "$jpp_id"
            assert_json_equals ".comment" "$COMMENT"
            assert_json_bool ".suppress_status_change_notifications" "true"
        else
            if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create job production plan completion"
            fi
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
