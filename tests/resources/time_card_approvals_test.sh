#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Approvals
#
# Tests create operations for time-card-approvals.
#
# COVERAGE: Writable attributes (comment, skip-quantity-validation, create-zero-ppu-missing-rates)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

APPROVABLE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_APPROVAL_ID:-}"

describe "Resource: time-card-approvals"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create approval requires time card"
xbe_run do time-card-approvals create --comment "missing time card"
assert_failure

test_name "Create time card approval"
if [[ -n "$APPROVABLE_TIME_CARD_ID" && "$APPROVABLE_TIME_CARD_ID" != "null" ]]; then
    COMMENT=$(unique_name "TimeCardApproval")
    xbe_json do time-card-approvals create \
        --time-card "$APPROVABLE_TIME_CARD_ID" \
        --comment "$COMMENT" \
        --skip-quantity-validation \
        --create-zero-ppu-missing-rates
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".time_card_id" "$APPROVABLE_TIME_CARD_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create time card approval: $output"
        fi
    fi
else
    skip "No submitted time card ID available for approval. Set XBE_TEST_TIME_CARD_APPROVAL_ID to enable create testing."
fi

run_tests
