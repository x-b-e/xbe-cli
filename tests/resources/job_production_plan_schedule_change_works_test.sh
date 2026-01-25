#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Schedule Change Works
#
# Tests list and show operations for the job-production-plan-schedule-change-works resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

WORK_ID=""
JPP_ID=""
CREATED_BY_ID=""
TIME_KIND=""
BROKER_ID=""
CUSTOMER_ID=""
PROJECT_ID=""
SKIP_ID_FILTERS=0

describe "Resource: job-production-plan-schedule-change-works"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List schedule change works"
xbe_json view job-production-plan-schedule-change-works list --limit 5
assert_success

test_name "List schedule change works returns array"
xbe_json view job-production-plan-schedule-change-works list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list schedule change works"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample schedule change work"
xbe_json view job-production-plan-schedule-change-works list --limit 1
if [[ $status -eq 0 ]]; then
    WORK_ID=$(json_get ".[0].id")
    JPP_ID=$(json_get ".[0].job_production_plan_id")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    TIME_KIND=$(json_get ".[0].time_kind")
    if [[ -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No schedule change works available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list schedule change works"
fi

# ============================================================================
# Relationship Lookup via API
# ============================================================================

test_name "Lookup job production plan relationships via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
elif [[ -n "$JPP_ID" && "$JPP_ID" != "null" ]]; then
    base_url="${XBE_BASE_URL%/}"

    plan_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plans/$JPP_ID?fields[job-production-plans]=broker,customer,project" || true)

    BROKER_ID=$(echo "$plan_json" | jq -r '.data.relationships.broker.data.id // empty' 2>/dev/null || true)
    CUSTOMER_ID=$(echo "$plan_json" | jq -r '.data.relationships.customer.data.id // empty' 2>/dev/null || true)
    PROJECT_ID=$(echo "$plan_json" | jq -r '.data.relationships.project.data.id // empty' 2>/dev/null || true)

    if [[ -z "$BROKER_ID" && -z "$CUSTOMER_ID" && -z "$PROJECT_ID" ]]; then
        skip "No broker/customer/project relationships found"
    else
        pass
    fi
else
    skip "No job production plan ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List schedule change works with --job-production-plan filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$JPP_ID" && "$JPP_ID" != "null" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --job-production-plan "$JPP_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List schedule change works with --created-by filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List schedule change works with --time-kind filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TIME_KIND" && "$TIME_KIND" != "null" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --time-kind "$TIME_KIND" --limit 5
    assert_success
else
    skip "No time kind available"
fi

test_name "List schedule change works with --broker filter"
if [[ -n "$BROKER_ID" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List schedule change works with --customer filter"
if [[ -n "$CUSTOMER_ID" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --customer "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "List schedule change works with --project filter"
if [[ -n "$PROJECT_ID" ]]; then
    xbe_json view job-production-plan-schedule-change-works list --project "$PROJECT_ID" --limit 5
    assert_success
else
    skip "No project ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show schedule change work"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$WORK_ID" && "$WORK_ID" != "null" ]]; then
    xbe_json view job-production-plan-schedule-change-works show "$WORK_ID"
    assert_success
else
    skip "No schedule change work ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
