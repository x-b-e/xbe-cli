#!/bin/bash
#
# XBE CLI Integration Tests: Tender Job Schedule Shift Cancellations
#
# Tests list/show/create operations for tender-job-schedule-shift-cancellations.
#
# COVERAGE: List + writable attributes (status-change-comment, status-changed-by,
# is-returned, job-production-plan-cancellation-comment, skip-trucker-notifications)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
STATUS_CHANGED_BY_ID="${XBE_TEST_USER_ID:-}"
SAMPLE_ID=""

describe "Resource: tender-job-schedule-shift-cancellations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List tender job schedule shift cancellations"
xbe_json view tender-job-schedule-shift-cancellations list --limit 5
assert_success

test_name "List tender job schedule shift cancellations returns array"
xbe_json view tender-job-schedule-shift-cancellations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tender job schedule shift cancellations"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample cancellation"
xbe_json view tender-job-schedule-shift-cancellations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No cancellations available for show test"
    fi
else
    skip "Could not list cancellations to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show tender job schedule shift cancellation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view tender-job-schedule-shift-cancellations show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show cancellation: $output"
        fi
    fi
else
    skip "No cancellation ID available for show"
fi

# ============================================================================
# Resolve current user (optional)
# ============================================================================

if [[ -z "$STATUS_CHANGED_BY_ID" || "$STATUS_CHANGED_BY_ID" == "null" ]]; then
    test_name "Resolve current user for status-changed-by"
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        STATUS_CHANGED_BY_ID=$(json_get ".id")
        if [[ -n "$STATUS_CHANGED_BY_ID" && "$STATUS_CHANGED_BY_ID" != "null" ]]; then
            pass
        else
            skip "No user ID returned from auth whoami"
        fi
    else
        skip "Failed to resolve current user"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cancellation requires tender job schedule shift"
xbe_run do tender-job-schedule-shift-cancellations create --status-change-comment "missing shift"
assert_failure

test_name "Create tender job schedule shift cancellation"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    COMMENT=$(unique_name "ShiftCancellation")
    JPP_COMMENT=$(unique_name "JppCancellation")
    create_args=(do tender-job-schedule-shift-cancellations create
        --tender-job-schedule-shift "$SHIFT_ID"
        --status-change-comment "$COMMENT"
        --job-production-plan-cancellation-comment "$JPP_COMMENT"
        --is-returned false
        --skip-trucker-notifications true)
    if [[ -n "$STATUS_CHANGED_BY_ID" && "$STATUS_CHANGED_BY_ID" != "null" ]]; then
        create_args+=(--status-changed-by "$STATUS_CHANGED_BY_ID")
    fi

    xbe_json "${create_args[@]}"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".tender_job_schedule_shift_id" "$SHIFT_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"already cancelled"* ]] || [[ "$output" == *"already canceled"* ]] || [[ "$output" == *"cannot cancel"* ]] || [[ "$output" == *"time card"* ]] || [[ "$output" == *"material transactions"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"not valid"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create cancellation: $output"
        fi
    fi
else
    skip "No tender job schedule shift ID available. Set XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID to enable create testing."
fi

run_tests
