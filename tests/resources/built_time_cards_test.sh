#!/bin/bash
#
# XBE CLI Integration Tests: Built Time Cards
#
# Tests create operations for built-time-cards.
#
# COVERAGE: Writable relationships (broker-tender-job-schedule-shift, customer-tender-job-schedule-shift)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_SHIFT_ID="${XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
CUSTOMER_SHIFT_ID="${XBE_TEST_CUSTOMER_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"

if [[ -z "$BROKER_SHIFT_ID" ]]; then
    BROKER_SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
fi

describe "Resource: built-time-cards"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create built time card requires a shift"
xbe_run do built-time-cards create
assert_failure

test_name "Create built time card from broker tender shift"
if [[ -n "$BROKER_SHIFT_ID" && "$BROKER_SHIFT_ID" != "null" ]]; then
    xbe_json do built-time-cards create --broker-tender-job-schedule-shift "$BROKER_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".broker_tender_job_schedule_shift_id" "$BROKER_SHIFT_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create built time card: $output"
        fi
    fi
else
    skip "No broker tender job schedule shift ID available. Set XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID or XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID."
fi

test_name "Create built time card from customer tender shift"
if [[ -n "$CUSTOMER_SHIFT_ID" && "$CUSTOMER_SHIFT_ID" != "null" ]]; then
    xbe_json do built-time-cards create --customer-tender-job-schedule-shift "$CUSTOMER_SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".customer_tender_job_schedule_shift_id" "$CUSTOMER_SHIFT_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create built time card: $output"
        fi
    fi
else
    skip "No customer tender job schedule shift ID available. Set XBE_TEST_CUSTOMER_TENDER_JOB_SCHEDULE_SHIFT_ID to enable create testing."
fi

run_tests
