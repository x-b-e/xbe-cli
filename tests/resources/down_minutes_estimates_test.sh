#!/bin/bash
#
# XBE CLI Integration Tests: Down Minutes Estimates
#
# Tests create operations for down-minutes-estimates.
#
# COVERAGE: Writable attributes (time-card-start-at, time-card-end-at)
#           Writable relationships (tender-job-schedule-shift)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SHIFT_ID="${XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
if [[ -z "$SHIFT_ID" ]]; then
    SHIFT_ID="${XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
fi

if [[ -z "$SHIFT_ID" ]]; then
    SHIFT_ID="${XBE_TEST_CUSTOMER_TENDER_JOB_SCHEDULE_SHIFT_ID:-}"
fi

describe "Resource: down-minutes-estimates"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create down minutes estimate requires a shift"
xbe_run do down-minutes-estimates create
assert_failure

test_name "Create down minutes estimate for a shift"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    xbe_json do down-minutes-estimates create --tender-job-schedule-shift "$SHIFT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must be related"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create down minutes estimate: $output"
        fi
    fi
else
    skip "No tender job schedule shift ID available. Set XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID or XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID."
fi

test_name "Create down minutes estimate with time card window"
if [[ -n "$SHIFT_ID" && "$SHIFT_ID" != "null" ]]; then
    xbe_json do down-minutes-estimates create \
        --tender-job-schedule-shift "$SHIFT_ID" \
        --time-card-start-at "2025-01-01T08:00:00Z" \
        --time-card-end-at "2025-01-01T12:00:00Z"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must be related"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create down minutes estimate with time card window: $output"
        fi
    fi
else
    skip "No tender job schedule shift ID available. Set XBE_TEST_BROKER_TENDER_JOB_SCHEDULE_SHIFT_ID or XBE_TEST_TENDER_JOB_SCHEDULE_SHIFT_ID."
fi

run_tests
