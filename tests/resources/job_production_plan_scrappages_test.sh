#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Scrappages
#
# Tests create operations for
# job_production_plan_scrappages.
#
# COVERAGE: Writable attributes (comment, suppress-status-change-notifications,
#           job-production-plan-cancellation-reason-type)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SCRAPPABLE_JPP_ID=""
CANCELLATION_REASON_TYPE_ID=""

# Optional: provide a known-safe approved plan ID for scrappage
SCRAPPABLE_JPP_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_SCRAPPAGE_ID:-}"
if [[ -z "$SCRAPPABLE_JPP_ID" && -n "$XBE_TEST_JOB_PRODUCTION_PLAN_ID" ]]; then
    SCRAPPABLE_JPP_ID="$XBE_TEST_JOB_PRODUCTION_PLAN_ID"
fi

describe "Resource: job-production-plan-scrappages"

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create scrappage requires job production plan"
xbe_run do job-production-plan-scrappages create --comment "missing plan"
assert_failure

# Validate scrappable plan status (approved) if provided
if [[ -n "$SCRAPPABLE_JPP_ID" && "$SCRAPPABLE_JPP_ID" != "null" ]]; then
    xbe_json view job-production-plans show "$SCRAPPABLE_JPP_ID"
    if [[ $status -eq 0 ]]; then
        PLAN_STATUS=$(json_get ".status")
        if [[ "$PLAN_STATUS" != "approved" ]]; then
            echo "    Provided job production plan is not approved (status: $PLAN_STATUS)."
            SCRAPPABLE_JPP_ID=""
        fi
    else
        SCRAPPABLE_JPP_ID=""
    fi
fi

# Load a cancellation reason type if available
xbe_json view job-production-plan-cancellation-reason-types list --limit 1
if [[ $status -eq 0 ]]; then
    CANCELLATION_REASON_TYPE_ID=$(echo "$output" | jq -r '.[0].id // empty')
fi

test_name "Create job production plan scrappage"
if [[ -n "$SCRAPPABLE_JPP_ID" && "$SCRAPPABLE_JPP_ID" != "null" ]]; then
    COMMENT=$(unique_name "JPPScrappage")
    CREATE_ARGS=(do job-production-plan-scrappages create
        --job-production-plan "$SCRAPPABLE_JPP_ID"
        --comment "$COMMENT"
        --suppress-status-change-notifications)

    if [[ -n "$CANCELLATION_REASON_TYPE_ID" && "$CANCELLATION_REASON_TYPE_ID" != "null" ]]; then
        CREATE_ARGS+=(--job-production-plan-cancellation-reason-type "$CANCELLATION_REASON_TYPE_ID")
    else
        echo "    (No cancellation reason type available; creating without it)"
    fi

    xbe_json "${CREATE_ARGS[@]}"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create job production plan scrappage"
    fi
else
    skip "No approved job production plan ID available for scrappage. Set XBE_TEST_JOB_PRODUCTION_PLAN_SCRAPPAGE_ID to enable create testing."
fi

run_tests
