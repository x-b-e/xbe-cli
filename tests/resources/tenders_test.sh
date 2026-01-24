#!/bin/bash
#
# XBE CLI Integration Tests: Tenders
#
# Tests list filters and show operations for the tenders resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TENDER_ID="${XBE_TEST_TENDER_ID:-}"
JOB_ID=""
JOB_SITE_ID="${XBE_TEST_TENDER_JOB_SITE_ID:-}"
JOB_NUMBER=""
BUYER_ID=""
BUYER_TYPE=""
SELLER_ID=""
SELLER_TYPE=""
BROKER_ID="${XBE_TEST_BROKER_ID:-}"
STATUS_VALUE=""

describe "Resource: tenders"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tenders"
xbe_json view tenders list --limit 5
if [[ $status -eq 0 ]]; then
    if [[ -z "$TENDER_ID" ]]; then
        TENDER_ID=$(json_get ".[0].id")
    fi
    JOB_ID=$(json_get ".[0].job_id")
    JOB_NUMBER=$(json_get ".[0].job_number")
    BUYER_ID=$(json_get ".[0].buyer_id")
    BUYER_TYPE=$(json_get ".[0].buyer_type")
    SELLER_ID=$(json_get ".[0].seller_id")
    SELLER_TYPE=$(json_get ".[0].seller_type")
    STATUS_VALUE=$(json_get ".[0].status")
    if [[ -z "$BROKER_ID" ]]; then
        if [[ "$BUYER_TYPE" == "brokers" ]]; then
            BROKER_ID="$BUYER_ID"
        elif [[ "$SELLER_TYPE" == "brokers" ]]; then
            BROKER_ID="$SELLER_ID"
        fi
    fi
    pass
else
    fail "Failed to list tenders"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
    xbe_json view tenders show "$TENDER_ID"
    assert_success
else
    skip "No tender ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

run_filter() {
    local name="$1"
    shift
    test_name "$name"
    xbe_json view tenders list "$@" --limit 5
    assert_success
}

run_filter "Filter by buyer" --buyer "${BUYER_ID:-1}"
run_filter "Filter by seller" --seller "${SELLER_ID:-1}"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    run_filter "Filter by broker" --broker "$BROKER_ID"
else
    run_filter "Filter by broker" --broker 1
fi
if [[ -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
    run_filter "Filter by job" --job "$JOB_ID"
else
    run_filter "Filter by job" --job 1
fi
if [[ -n "$STATUS_VALUE" && "$STATUS_VALUE" != "null" ]]; then
    run_filter "Filter by status" --status "$STATUS_VALUE"
else
    run_filter "Filter by status" --status editing
fi

if [[ -n "$JOB_SITE_ID" ]]; then
    run_filter "Filter by job site" --job-site "$JOB_SITE_ID"
else
    test_name "Filter by job site"
    skip "No job site ID available"
fi

if [[ -n "$JOB_NUMBER" && "$JOB_NUMBER" != "null" ]]; then
    run_filter "Filter by job number" --job-number "$JOB_NUMBER"
else
    run_filter "Filter by job number" --job-number "JOB"
fi

run_filter "Filter by start-at-min" --start-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by start-at-max" --start-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by end-at-max" --end-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by with-alive-shifts" --with-alive-shifts true
run_filter "Filter by has-flexible-shifts" --has-flexible-shifts false
run_filter "Filter by job production plan name or number" --job-production-plan-name-or-number-like "Plan"
run_filter "Filter by business unit" --business-unit 1
run_filter "Filter by trailer classification or equivalent" --job-production-plan-trailer-classification-or-equivalent 1
run_filter "Filter by job production plan material sites" --job-production-plan-material-sites 1
run_filter "Filter by job production plan material types" --job-production-plan-material-types 1
run_filter "Filter by created-at-min" --created-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by created-at-max" --created-at-max "2030-01-01T00:00:00Z"
run_filter "Filter by updated-at-min" --updated-at-min "2024-01-01T00:00:00Z"
run_filter "Filter by updated-at-max" --updated-at-max "2030-01-01T00:00:00Z"

# ============================================================================
# Summary
# ============================================================================

run_tests
