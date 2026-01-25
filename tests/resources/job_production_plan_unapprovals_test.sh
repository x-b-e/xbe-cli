#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Unapprovals
#
# Tests create operations for
# job_production_plan_unapprovals.
#
# COVERAGE: Writable attributes (comment, suppress-status-change-notifications)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

UNAPPROVABLE_JPP_ID=""

# Optional: provide a known-safe approved or scrapped plan ID for unapproval
UNAPPROVABLE_JPP_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_UNAPPROVAL_ID:-}"
if [[ -z "$UNAPPROVABLE_JPP_ID" && -n "$XBE_TEST_JOB_PRODUCTION_PLAN_ID" ]]; then
    UNAPPROVABLE_JPP_ID="$XBE_TEST_JOB_PRODUCTION_PLAN_ID"
fi

describe "Resource: job-production-plan-unapprovals"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create unapproval requires job production plan"
xbe_run do job-production-plan-unapprovals create --comment "missing plan"
assert_failure

# Validate unapprovable plan status (approved or scrapped) if provided
if [[ -n "$UNAPPROVABLE_JPP_ID" && "$UNAPPROVABLE_JPP_ID" != "null" ]]; then
    xbe_json view job-production-plans show "$UNAPPROVABLE_JPP_ID"
    if [[ $status -eq 0 ]]; then
        PLAN_STATUS=$(json_get ".status")
        if [[ "$PLAN_STATUS" != "approved" && "$PLAN_STATUS" != "scrapped" ]]; then
            echo "    Provided job production plan is not approved or scrapped (status: $PLAN_STATUS)."
            UNAPPROVABLE_JPP_ID=""
        fi
    else
        UNAPPROVABLE_JPP_ID=""
    fi
fi

test_name "Create job production plan unapproval"
if [[ -n "$UNAPPROVABLE_JPP_ID" && "$UNAPPROVABLE_JPP_ID" != "null" ]]; then
    COMMENT=$(unique_name "JPPUnapproval")
    xbe_json do job-production-plan-unapprovals create \
        --job-production-plan "$UNAPPROVABLE_JPP_ID" \
        --comment "$COMMENT" \
        --suppress-status-change-notifications
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create job production plan unapproval"
    fi
else
    skip "No approved or scrapped job production plan ID available for unapproval. Set XBE_TEST_JOB_PRODUCTION_PLAN_UNAPPROVAL_ID to enable create testing."
fi

run_tests
