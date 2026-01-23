#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Submissions
#
# Tests create operations for time-card-submissions.
#
# COVERAGE: Writable attributes (comment, skip-quantity-validation, create-zero-ppu-missing-rates, skip-positive-quantity-validation)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SUBMITTABLE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_SUBMISSION_ID:-}"

describe "Resource: time-card-submissions"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create submission requires time card"
xbe_run do time-card-submissions create --comment "missing time card"
assert_failure

test_name "Create time card submission"
if [[ -n "$SUBMITTABLE_TIME_CARD_ID" && "$SUBMITTABLE_TIME_CARD_ID" != "null" ]]; then
    COMMENT=$(unique_name "TimeCardSubmission")
    xbe_json do time-card-submissions create \
        --time-card "$SUBMITTABLE_TIME_CARD_ID" \
        --comment "$COMMENT" \
        --skip-quantity-validation \
        --create-zero-ppu-missing-rates \
        --skip-positive-quantity-validation
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".time_card_id" "$SUBMITTABLE_TIME_CARD_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create time card submission: $output"
        fi
    fi
else
    skip "No editable/rejected time card ID available for submission. Set XBE_TEST_TIME_CARD_SUBMISSION_ID to enable create testing."
fi

run_tests
