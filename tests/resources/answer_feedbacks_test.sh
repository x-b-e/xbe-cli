#!/bin/bash
#
# XBE CLI Integration Tests: Answer Feedbacks
#
# Tests CRUD operations for the answer-feedbacks resource.
# Answer feedbacks capture scores and notes on answers.
#
# NOTE: Creating answer feedbacks requires an existing answer.
# Use XBE_TEST_ANSWER_ID to supply a valid answer ID.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FEEDBACK_ID=""
ANSWER_ID="${XBE_TEST_ANSWER_ID:-}"

describe "Resource: answer-feedbacks"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List answer feedbacks"
xbe_json view answer-feedbacks list --limit 5
assert_success

test_name "List answer feedbacks returns array"
xbe_json view answer-feedbacks list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list answer feedbacks"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List answer feedbacks with --answer filter"
if [[ -n "$ANSWER_ID" ]]; then
    xbe_json view answer-feedbacks list --answer "$ANSWER_ID" --limit 5
    assert_success
else
    xbe_json view answer-feedbacks list --answer "nonexistent" --limit 5
    assert_success
fi

# ============================================================================
# CREATE Tests - Validation
# ============================================================================

test_name "Create answer feedback requires --answer"
xbe_run do answer-feedbacks create --score 0.5
assert_failure

test_name "Create answer feedback rejects invalid --score"
xbe_run do answer-feedbacks create --answer "123" --score 1.5
assert_failure

# ============================================================================
# CREATE Tests - Success (requires answer)
# ============================================================================

if [[ -n "$ANSWER_ID" ]]; then
    test_name "Create answer feedback with optional fields"
    xbe_json do answer-feedbacks create \
        --answer "$ANSWER_ID" \
        --score 0.6 \
        --notes "Test notes" \
        --better-content "Test better content"

    if [[ $status -eq 0 ]]; then
        CREATED_FEEDBACK_ID=$(json_get ".id")
        if [[ -n "$CREATED_FEEDBACK_ID" && "$CREATED_FEEDBACK_ID" != "null" ]]; then
            register_cleanup "answer-feedbacks" "$CREATED_FEEDBACK_ID"
            pass
        else
            fail "Created answer feedback but no ID returned"
        fi
    else
        fail "Failed to create answer feedback"
    fi
else
    skip "XBE_TEST_ANSWER_ID not set; skipping create/update/delete success tests"
fi

# ============================================================================
# UPDATE Tests - Error cases
# ============================================================================

test_name "Update answer feedback without any fields fails"
xbe_run do answer-feedbacks update "nonexistent"
assert_failure

# ============================================================================
# UPDATE Tests - Success (requires created feedback)
# ============================================================================

if [[ -n "$CREATED_FEEDBACK_ID" && "$CREATED_FEEDBACK_ID" != "null" ]]; then
    test_name "Update answer feedback score"
    xbe_json do answer-feedbacks update "$CREATED_FEEDBACK_ID" --score 0.9
    assert_success

    test_name "Update answer feedback notes"
    xbe_json do answer-feedbacks update "$CREATED_FEEDBACK_ID" --notes "Updated notes"
    assert_success

    test_name "Update answer feedback better content"
    xbe_json do answer-feedbacks update "$CREATED_FEEDBACK_ID" --better-content "Updated better content"
    assert_success
fi

# ============================================================================
# DELETE Tests - Error cases
# ============================================================================

test_name "Delete answer feedback requires --confirm flag"
xbe_run do answer-feedbacks delete "nonexistent"
assert_failure

# ============================================================================
# DELETE Tests - Success (requires created feedback)
# ============================================================================

if [[ -n "$CREATED_FEEDBACK_ID" && "$CREATED_FEEDBACK_ID" != "null" ]]; then
    test_name "Delete answer feedback"
    xbe_run do answer-feedbacks delete "$CREATED_FEEDBACK_ID" --confirm
    assert_success
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
