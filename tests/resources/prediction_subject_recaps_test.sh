#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Recaps
#
# Tests list/show operations for prediction-subject-recaps.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_RECAP_SUBJECT_ID:-}"
if [[ -z "$PREDICTION_SUBJECT_ID" ]]; then
    PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
fi

SAMPLE_ID=""
SAMPLE_PREDICTION_SUBJECT_ID=""

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"Not Found"* ]] || [[ "$msg" == *"not found"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: prediction-subject-recaps"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction subject recaps"
xbe_json view prediction-subject-recaps list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_PREDICTION_SUBJECT_ID=$(json_get ".[0].prediction_subject_id")
else
    fail "Failed to list prediction subject recaps"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction subject recap"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-subject-recaps show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show prediction subject recap: $output"
        fi
    fi
else
    skip "No prediction subject recap ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction subject recaps with --prediction-subject filter"
FILTER_PREDICTION_SUBJECT_ID="${SAMPLE_PREDICTION_SUBJECT_ID:-$PREDICTION_SUBJECT_ID}"
if [[ -n "$FILTER_PREDICTION_SUBJECT_ID" && "$FILTER_PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json view prediction-subject-recaps list --prediction-subject "$FILTER_PREDICTION_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No prediction subject ID available (set XBE_TEST_PREDICTION_SUBJECT_ID)"
fi

run_tests
