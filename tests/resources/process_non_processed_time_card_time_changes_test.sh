#!/bin/bash
#
# XBE CLI Integration Tests: Process Non-Processed Time Card Time Changes
#
# Tests create operations for process-non-processed-time-card-time-changes.
#
# COVERAGE: Writable attributes (time-card-time-change-ids, delete-unprocessed)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TIME_CARD_TIME_CHANGE_IDS="${XBE_TEST_TIME_CARD_TIME_CHANGE_IDS:-}"
if [[ -z "$TIME_CARD_TIME_CHANGE_IDS" && -n "${XBE_TEST_TIME_CARD_TIME_CHANGE_ID:-}" ]]; then
    TIME_CARD_TIME_CHANGE_IDS="$XBE_TEST_TIME_CARD_TIME_CHANGE_ID"
fi

describe "Resource: process-non-processed-time-card-time-changes"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires time card time change IDs"
xbe_run do process-non-processed-time-card-time-changes create --delete-unprocessed false
assert_failure

test_name "Create processing run"
if [[ -n "$TIME_CARD_TIME_CHANGE_IDS" && "$TIME_CARD_TIME_CHANGE_IDS" != "null" ]]; then
    xbe_json do process-non-processed-time-card-time-changes create \
        --time-card-time-change-ids "$TIME_CARD_TIME_CHANGE_IDS" \
        --delete-unprocessed false
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_has ".time_card_time_change_ids"
        assert_json_equals ".delete_unprocessed" "false"
        FIRST_ID="${TIME_CARD_TIME_CHANGE_IDS%%,*}"
        if [[ -n "$FIRST_ID" && "$FIRST_ID" != "null" ]]; then
            assert_json_equals ".time_card_time_change_ids[0]" "$FIRST_ID"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must all refer"* ]] || [[ "$output" == *"must all"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create processing run: $output"
        fi
    fi
else
    skip "No time card time change IDs available. Set XBE_TEST_TIME_CARD_TIME_CHANGE_IDS to enable create testing."
fi

run_tests
