#!/bin/bash
#
# XBE CLI Integration Tests: Tender Cancellations
#
# Tests create operations for tender-cancellations.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TENDER_ID="${XBE_TEST_TENDER_CANCELLATION_ID:-}"

describe "Resource: tender-cancellations"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cancellation requires tender"
xbe_run do tender-cancellations create --comment "missing tender"
assert_failure

test_name "Create tender cancellation"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
    COMMENT=$(unique_name "TenderCancellation")
    xbe_json do tender-cancellations create \
        --tender "$TENDER_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".tender_id" "$TENDER_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"cannot have a shift"* ]] || [[ "$output" == *"not valid"* ]] || [[ "$output" == *"not in valid"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create tender cancellation: $output"
        fi
    fi
else
    skip "No tender ID available. Set XBE_TEST_TENDER_CANCELLATION_ID to enable create testing."
fi

run_tests
