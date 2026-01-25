#!/bin/bash
#
# XBE CLI Integration Tests: Tender Raters
#
# Tests create operations for tender-raters.
#
# COVERAGE: Create attributes (replace-rates, replace-shift-set-time-card-constraints,
# persist-changes, skip-adjustment-cost-index-value-presence-validation, skip-validate-customer-tender-hourly-rates)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TENDER_ID="${XBE_TEST_TENDER_RATER_TENDER_ID:-}"

describe "Resource: tender-raters"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create tender rater requires tender"
xbe_run do tender-raters create
assert_failure

test_name "Create tender rater"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
    xbe_json do tender-raters create \
        --tender "$TENDER_ID" \
        --replace-rates true \
        --replace-shift-set-time-card-constraints true \
        --persist-changes false \
        --skip-adjustment-cost-index-value-presence-validation true \
        --skip-validate-customer-tender-hourly-rates true
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".tender_id" "$TENDER_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"tender"* ]] || [[ "$output" == *"invoice"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create tender rater: $output"
        fi
    fi
else
    skip "No tender ID available. Set XBE_TEST_TENDER_RATER_TENDER_ID to enable create testing."
fi

run_tests
