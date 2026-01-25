#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Knowledge Base Answers
#
# Tests view operations for prediction-knowledge-base-answers.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_QUESTION_ID=""
LIST_SUPPORTED="false"

describe "Resource: prediction-knowledge-base-answers (view-only)"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List prediction knowledge base answers"
xbe_json view prediction-knowledge-base-answers list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Prediction knowledge base answers list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List prediction knowledge base answers returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view prediction-knowledge-base-answers list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list prediction knowledge base answers"
    fi
else
    skip "Prediction knowledge base answers list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show/filter)
# ==========================================================================

test_name "Capture sample prediction knowledge base answer"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view prediction-knowledge-base-answers list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_QUESTION_ID=$(json_get ".[0].prediction_knowledge_base_question_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No prediction knowledge base answers available for show"
        fi
    else
        skip "Could not list prediction knowledge base answers to capture sample"
    fi
else
    skip "Prediction knowledge base answers list endpoint not available"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List prediction knowledge base answers with --prediction-knowledge-base-question filter"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    if [[ -n "$SAMPLE_QUESTION_ID" && "$SAMPLE_QUESTION_ID" != "null" ]]; then
        xbe_json view prediction-knowledge-base-answers list --prediction-knowledge-base-question "$SAMPLE_QUESTION_ID" --limit 5
        assert_success
    else
        xbe_json view prediction-knowledge-base-answers list --prediction-knowledge-base-question "nonexistent" --limit 5
        assert_success
    fi
else
    skip "Prediction knowledge base answers list endpoint not available"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show prediction knowledge base answer"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-base-answers show "$SAMPLE_ID"
    assert_success
else
    skip "No prediction knowledge base answer ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
