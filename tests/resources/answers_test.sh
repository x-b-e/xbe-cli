#!/bin/bash
#
# XBE CLI Integration Tests: Answers
#
# Tests view operations for the answers resource.
# Answers are generated for questions and can have feedback.
#
# NOTE: Showing a specific answer requires XBE_TEST_ANSWER_ID.
# Optionally set XBE_TEST_QUESTION_ID for question filter coverage.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

ANSWER_ID="${XBE_TEST_ANSWER_ID:-}"
QUESTION_ID="${XBE_TEST_QUESTION_ID:-}"

describe "Resource: answers (view-only)"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List answers"
xbe_json view answers list --limit 5
assert_success

test_name "List answers returns array"
xbe_json view answers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list answers"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List answers with --question filter"
if [[ -n "$QUESTION_ID" ]]; then
    xbe_json view answers list --question "$QUESTION_ID" --limit 5
    assert_success
else
    xbe_json view answers list --question "nonexistent" --limit 5
    assert_success
fi

test_name "List answers with --with-feedback filter"
xbe_json view answers list --with-feedback true --limit 5
assert_success

test_name "List answers with --without-feedback filter"
xbe_json view answers list --without-feedback true --limit 5
assert_success

# ==========================================================================
# SHOW Tests
# ==========================================================================

if [[ -n "$ANSWER_ID" ]]; then
    test_name "Show answer"
    xbe_json view answers show "$ANSWER_ID"
    assert_success
else
    skip "XBE_TEST_ANSWER_ID not set; skipping show test"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
