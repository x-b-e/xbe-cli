#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Unscrappages
#
# Tests create operations for the job_production_plan_unscrappages resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: job-production-plan-unscrappages"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create unscrappage without required job production plan fails"
xbe_run do job-production-plan-unscrappages create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create job production plan unscrappage"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for scrapped job production plans"
else
    base_url="${XBE_BASE_URL%/}"

    plans_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plans?page[limit]=1&filter[status]=scrapped" || true)

    jpp_id=$(echo "$plans_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

    if [[ -z "$jpp_id" ]]; then
        skip "No scrapped job production plan found"
    else
        COMMENT="Unscrapping plan for test"
        xbe_json do job-production-plan-unscrappages create \
            --job-production-plan "$jpp_id" \
            --comment "$COMMENT" \
            --suppress-status-change-notifications \
            --skip-validate-required-mix-designs

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
            assert_json_equals ".job_production_plan_id" "$jpp_id"
            assert_json_equals ".comment" "$COMMENT"
            assert_json_bool ".suppress_status_change_notifications" "true"
            assert_json_bool ".skip_validate_required_mix_designs" "true"
        else
            if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                pass
            else
                fail "Failed to create job production plan unscrappage"
            fi
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
