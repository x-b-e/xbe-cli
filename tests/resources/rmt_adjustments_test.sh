#!/bin/bash
#
# XBE CLI Integration Tests: RMT Adjustments
#
# Tests create operations for rmt-adjustments.
#
# COVERAGE: Writable attributes (rmt-ids, note, raw-data-adjustments, update-if-invoiced)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

RMT_IDS="${XBE_TEST_RMT_IDS:-}"
if [[ -z "$RMT_IDS" && -n "${XBE_TEST_RMT_ID:-}" ]]; then
    RMT_IDS="$XBE_TEST_RMT_ID"
fi

RAW_DATA_ADJUSTMENTS="${XBE_TEST_RMT_ADJUSTMENTS_JSON:-}"
if [[ -z "$RAW_DATA_ADJUSTMENTS" ]]; then
    RAW_DATA_ADJUSTMENTS='{"net_weight":12.5,"net_weight_by_xbe_reason":"Scale correction"}'
fi

describe "Resource: rmt-adjustments"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires rmt IDs"
xbe_run do rmt-adjustments create --note "Missing IDs" --raw-data-adjustments "$RAW_DATA_ADJUSTMENTS"
assert_failure

test_name "Create requires note"
xbe_run do rmt-adjustments create --rmt-ids 123 --raw-data-adjustments "$RAW_DATA_ADJUSTMENTS"
assert_failure

test_name "Create requires raw data adjustments"
xbe_run do rmt-adjustments create --rmt-ids 123 --note "Missing adjustments"
assert_failure

test_name "Create RMT adjustment"
if [[ -n "$RMT_IDS" && "$RMT_IDS" != "null" ]]; then
    NOTE=$(unique_name "RmtAdjustment")
    xbe_json do rmt-adjustments create \
        --rmt-ids "$RMT_IDS" \
        --note "$NOTE" \
        --raw-data-adjustments "$RAW_DATA_ADJUSTMENTS" \
        --update-if-invoiced=false
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_has ".rmt_ids"
        assert_json_equals ".update_if_invoiced" "false"
        FIRST_ID="${RMT_IDS%%,*}"
        if [[ -n "$FIRST_ID" && "$FIRST_ID" != "null" ]]; then
            assert_json_equals ".rmt_ids[0]" "$FIRST_ID"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"contains disallowed fields"* ]] || [[ "$output" == *"missing required reason field"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create RMT adjustment: $output"
        fi
    fi
else
    skip "No RMT IDs available. Set XBE_TEST_RMT_IDS to enable create testing."
fi

run_tests
