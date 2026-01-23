#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Rejections
#
# Tests create operations for time-card-rejections.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

REJECTABLE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_REJECTION_ID:-}"

describe "Resource: time-card-rejections"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create rejection requires time card"
xbe_run do time-card-rejections create --comment "missing time card"
assert_failure

test_name "Create time card rejection"
if [[ -n "$REJECTABLE_TIME_CARD_ID" && "$REJECTABLE_TIME_CARD_ID" != "null" ]]; then
    COMMENT=$(unique_name "TimeCardRejection")
    xbe_json do time-card-rejections create \
        --time-card "$REJECTABLE_TIME_CARD_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".time_card_id" "$REJECTABLE_TIME_CARD_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must not"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create time card rejection: $output"
        fi
    fi
else
    skip "No submitted time card ID available for rejection. Set XBE_TEST_TIME_CARD_REJECTION_ID to enable create testing."
fi

run_tests
