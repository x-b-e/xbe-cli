#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Subject Recap Generations
#
# Tests create operations for the prediction-subject-recap-generations resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"

describe "Resource: prediction-subject-recap-generations"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create recap generation without required prediction subject fails"
xbe_run do prediction-subject-recap-generations create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction subject recap generation"
if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json do prediction-subject-recap-generations create \
        --prediction-subject "$PREDICTION_SUBJECT_ID"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".prediction_subject_id" "$PREDICTION_SUBJECT_ID"
    else
        fail "Failed to create prediction subject recap generation"
    fi
else
    skip "XBE_TEST_PREDICTION_SUBJECT_ID not set"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
