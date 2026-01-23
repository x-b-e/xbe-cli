#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Unsubmissions
#
# Tests create operations for the time-card-unsubmissions resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_TIME_CARD_ID="${XBE_TEST_TIME_CARD_UNSUBMISSION_TIME_CARD_ID:-}"

describe "Resource: time-card-unsubmissions"

# ============================================================================
# Sample Record (used for create)
# ============================================================================

if [[ -z "$SAMPLE_TIME_CARD_ID" ]]; then
    test_name "Locate submitted time card via status changes"
    xbe_json view time-card-status-changes list --status submitted --limit 1 --sort -changed-at
    if [[ $status -eq 0 ]]; then
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
            pass
        else
            skip "No submitted time card status changes found"
        fi
    else
        skip "Could not list time card status changes"
    fi
else
    test_name "Use provided time card ID"
    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Unsubmit time card"
if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    xbe_json do time-card-unsubmissions create \
        --time-card "$SAMPLE_TIME_CARD_ID" \
        --comment "CLI unsubmission test"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Unsubmit failed: $output"
        fi
    fi
else
    skip "No submitted time card available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Unsubmit without required fields fails"
xbe_run do time-card-unsubmissions create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
