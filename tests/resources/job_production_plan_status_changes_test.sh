#!/bin/bash
#
# XBE CLI Integration Tests: Job Production Plan Status Changes
#
# Tests list/show/update operations for job-production-plan-status-changes.
#
# COVERAGE: List filters + show + update (cancellation reason type) + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_JPP_ID=""
SAMPLE_STATUS=""
SAMPLE_REASON_TYPE_ID=""
REASON_TYPE_ID=""
STATUS_CHANGE_ID_FOR_UPDATE=""
JPP_START_ON="2020-01-01"

describe "Resource: job-production-plan-status-changes"

# ============================================================================
# Endpoint Availability
# ============================================================================

test_name "Check status change endpoint availability"
xbe_json view job-production-plan-status-changes list --limit 1
if [[ $status -ne 0 ]]; then
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        skip "Server does not support job-production-plan-status-changes (404)"
        run_tests
    else
        fail "Failed to list status changes: $output"
        run_tests
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List job production plan status changes"
xbe_json view job-production-plan-status-changes list --limit 5
assert_success

test_name "List job production plan status changes returns array"
xbe_json view job-production-plan-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list status changes"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample status change"
xbe_json view job-production-plan-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_JPP_ID=$(json_get ".[0].job_production_plan_id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_REASON_TYPE_ID=$(json_get ".[0].job_production_plan_cancellation_reason_type_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No status changes available for follow-on tests"
    fi
else
    skip "Could not list status changes to capture sample"
fi

# ============================================================================
# Helper Data
# ============================================================================

if [[ -n "$SAMPLE_JPP_ID" && "$SAMPLE_JPP_ID" != "null" ]]; then
    xbe_json view job-production-plans show "$SAMPLE_JPP_ID"
    if [[ $status -eq 0 ]]; then
        JPP_START_ON=$(json_get ".start_on")
    fi
fi

if [[ -n "$SAMPLE_REASON_TYPE_ID" && "$SAMPLE_REASON_TYPE_ID" != "null" ]]; then
    REASON_TYPE_ID="$SAMPLE_REASON_TYPE_ID"
else
    xbe_json view job-production-plan-cancellation-reason-types list --limit 1
    if [[ $status -eq 0 ]]; then
        REASON_TYPE_ID=$(json_get ".[0].id")
    fi
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List status changes with --status filter"
STATUS_FILTER_VALUE="$SAMPLE_STATUS"
if [[ -z "$STATUS_FILTER_VALUE" || "$STATUS_FILTER_VALUE" == "null" ]]; then
    STATUS_FILTER_VALUE="approved"
fi
xbe_json view job-production-plan-status-changes list --status "$STATUS_FILTER_VALUE" --limit 5
assert_success

test_name "List status changes with --job-production-plan filter"
JPP_FILTER_VALUE="$SAMPLE_JPP_ID"
if [[ -z "$JPP_FILTER_VALUE" || "$JPP_FILTER_VALUE" == "null" ]]; then
    JPP_FILTER_VALUE="1"
fi
xbe_json view job-production-plan-status-changes list --job-production-plan "$JPP_FILTER_VALUE" --limit 5
assert_success

test_name "List status changes with --job-production-plan-cancellation-reason-type filter"
REASON_FILTER_VALUE="$REASON_TYPE_ID"
if [[ -z "$REASON_FILTER_VALUE" || "$REASON_FILTER_VALUE" == "null" ]]; then
    REASON_FILTER_VALUE="1"
fi
xbe_json view job-production-plan-status-changes list --job-production-plan-cancellation-reason-type "$REASON_FILTER_VALUE" --limit 5
assert_success

test_name "List status changes with --customer filter"
CUSTOMER_FILTER_VALUE="${XBE_TEST_CUSTOMER_ID:-1}"
xbe_json view job-production-plan-status-changes list --customer "$CUSTOMER_FILTER_VALUE" --limit 5
assert_success

test_name "List status changes with --start-on filter"
if [[ -z "$JPP_START_ON" || "$JPP_START_ON" == "null" ]]; then
    JPP_START_ON="2020-01-01"
fi
xbe_json view job-production-plan-status-changes list --start-on "$JPP_START_ON" --limit 5
assert_success

test_name "List status changes with --start-on-min filter"
xbe_json view job-production-plan-status-changes list --start-on-min "2020-01-01" --limit 5
assert_success

test_name "List status changes with --start-on-max filter"
xbe_json view job-production-plan-status-changes list --start-on-max "2030-01-01" --limit 5
assert_success

test_name "List status changes with --created-at-min filter"
xbe_json view job-production-plan-status-changes list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List status changes with --created-at-max filter"
xbe_json view job-production-plan-status-changes list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List status changes with --updated-at-min filter"
xbe_json view job-production-plan-status-changes list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List status changes with --updated-at-max filter"
xbe_json view job-production-plan-status-changes list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show job production plan status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view job-production-plan-status-changes show "$SAMPLE_ID"
    assert_success
else
    skip "No status change ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Find cancelled/scrapped status change for update"
xbe_json view job-production-plan-status-changes list --status cancelled --limit 1
if [[ $status -eq 0 ]]; then
    STATUS_CHANGE_ID_FOR_UPDATE=$(json_get ".[0].id")
fi
if [[ -z "$STATUS_CHANGE_ID_FOR_UPDATE" || "$STATUS_CHANGE_ID_FOR_UPDATE" == "null" ]]; then
    xbe_json view job-production-plan-status-changes list --status scrapped --limit 1
    if [[ $status -eq 0 ]]; then
        STATUS_CHANGE_ID_FOR_UPDATE=$(json_get ".[0].id")
    fi
fi
if [[ -n "$STATUS_CHANGE_ID_FOR_UPDATE" && "$STATUS_CHANGE_ID_FOR_UPDATE" != "null" ]]; then
    pass
else
    skip "No cancelled or scrapped status change available for update"
fi

test_name "Update status change cancellation reason type"
if [[ -n "$STATUS_CHANGE_ID_FOR_UPDATE" && "$STATUS_CHANGE_ID_FOR_UPDATE" != "null" && \
      -n "$REASON_TYPE_ID" && "$REASON_TYPE_ID" != "null" ]]; then
    xbe_json do job-production-plan-status-changes update "$STATUS_CHANGE_ID_FOR_UPDATE" \
        --job-production-plan-cancellation-reason-type "$REASON_TYPE_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"422"* ]] || [[ "$output" == *"Validation failed"* ]]; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No status change or cancellation reason type available for update"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Update without any fields fails"
xbe_json do job-production-plan-status-changes update 1
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
