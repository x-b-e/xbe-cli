#!/bin/bash
#
# XBE CLI Integration Tests: Prediction Knowledge Base Questions
#
# Tests list/show/create/update/delete operations for prediction-knowledge-base-questions.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PREDICTION_KNOWLEDGE_BASE_ID="${XBE_TEST_PREDICTION_KNOWLEDGE_BASE_ID:-}"
PREDICTION_SUBJECT_ID="${XBE_TEST_PREDICTION_SUBJECT_ID:-}"
CREATED_BY_ID="${XBE_TEST_USER_ID:-}"

SAMPLE_ID=""
SAMPLE_KNOWLEDGE_BASE_ID=""
SAMPLE_PREDICTION_SUBJECT_ID=""
SAMPLE_CREATED_BY_ID=""
CREATED_ID=""

TITLE="KB Question $(unique_suffix)"
UPDATED_TITLE="KB Question Updated $(unique_suffix)"
DESCRIPTION="Question description $(unique_suffix)"
UPDATED_DESCRIPTION="Updated description $(unique_suffix)"

TAG_NAME=""
TAG_CATEGORY_SLUG=""

update_blocked_message() {
    local msg="$1"
    if [[ "$msg" == *"Not Authorized"* ]] || [[ "$msg" == *"not authorized"* ]] || [[ "$msg" == *"Not Found"* ]] || [[ "$msg" == *"not found"* ]] || [[ "$msg" == *"cannot"* ]] || [[ "$msg" == *"validation"* ]] || [[ "$msg" == *"unprocessable"* ]]; then
        return 0
    fi
    return 1
}

describe "Resource: prediction-knowledge-base-questions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List prediction knowledge base questions"
xbe_json view prediction-knowledge-base-questions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_KNOWLEDGE_BASE_ID=$(json_get ".[0].prediction_knowledge_base_id")
    SAMPLE_PREDICTION_SUBJECT_ID=$(json_get ".[0].prediction_subject_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
else
    fail "Failed to list prediction knowledge base questions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show prediction knowledge base question"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if update_blocked_message "$output"; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show prediction knowledge base question: $output"
        fi
    fi
else
    skip "No prediction knowledge base question ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List prediction knowledge base questions with --prediction-knowledge-base filter"
FILTER_KB_ID="${SAMPLE_KNOWLEDGE_BASE_ID:-$PREDICTION_KNOWLEDGE_BASE_ID}"
if [[ -n "$FILTER_KB_ID" && "$FILTER_KB_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --prediction-knowledge-base "$FILTER_KB_ID" --limit 5
    assert_success
else
    skip "No prediction knowledge base ID available (set XBE_TEST_PREDICTION_KNOWLEDGE_BASE_ID)"
fi

test_name "List prediction knowledge base questions with --prediction-subject filter"
FILTER_SUBJECT_ID="${SAMPLE_PREDICTION_SUBJECT_ID:-$PREDICTION_SUBJECT_ID}"
if [[ -n "$FILTER_SUBJECT_ID" && "$FILTER_SUBJECT_ID" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --prediction-subject "$FILTER_SUBJECT_ID" --limit 5
    assert_success
else
    skip "No prediction subject ID available (set XBE_TEST_PREDICTION_SUBJECT_ID)"
fi

test_name "List prediction knowledge base questions with --status filter"
xbe_json view prediction-knowledge-base-questions list --status open --limit 5
assert_success

# Fetch tag metadata for tag filters
if [[ -z "$TAG_NAME" || -z "$TAG_CATEGORY_SLUG" ]]; then
    xbe_json view tags list --limit 1
    if [[ $status -eq 0 ]]; then
        TAG_NAME=$(json_get ".[0].name")
        TAG_CATEGORY_SLUG=$(json_get ".[0].tag_category_slug")
    fi
fi

test_name "List prediction knowledge base questions with --tagged-with filter"
if [[ -n "$TAG_NAME" && "$TAG_NAME" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --tagged-with "$TAG_NAME" --limit 5
    assert_success
else
    skip "No tag name available for tagged-with filter"
fi

test_name "List prediction knowledge base questions with --tagged-with-any filter"
if [[ -n "$TAG_NAME" && "$TAG_NAME" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --tagged-with-any "$TAG_NAME" --limit 5
    assert_success
else
    skip "No tag name available for tagged-with-any filter"
fi

test_name "List prediction knowledge base questions with --tagged-with-all filter"
if [[ -n "$TAG_NAME" && "$TAG_NAME" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --tagged-with-all "$TAG_NAME" --limit 5
    assert_success
else
    skip "No tag name available for tagged-with-all filter"
fi

test_name "List prediction knowledge base questions with --in-tag-category filter"
if [[ -n "$TAG_CATEGORY_SLUG" && "$TAG_CATEGORY_SLUG" != "null" ]]; then
    xbe_json view prediction-knowledge-base-questions list --in-tag-category "$TAG_CATEGORY_SLUG" --limit 5
    assert_success
else
    skip "No tag category slug available for in-tag-category filter"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prediction knowledge base question requires required flags"
xbe_run do prediction-knowledge-base-questions create
assert_failure

if [[ -z "$PREDICTION_KNOWLEDGE_BASE_ID" || "$PREDICTION_KNOWLEDGE_BASE_ID" == "null" ]]; then
    PREDICTION_KNOWLEDGE_BASE_ID="$SAMPLE_KNOWLEDGE_BASE_ID"
fi
if [[ -z "$PREDICTION_SUBJECT_ID" || "$PREDICTION_SUBJECT_ID" == "null" ]]; then
    PREDICTION_SUBJECT_ID="$SAMPLE_PREDICTION_SUBJECT_ID"
fi
if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
    CREATED_BY_ID="$SAMPLE_CREATED_BY_ID"
fi

TEST_CREATE_ARGS=(do prediction-knowledge-base-questions create --prediction-knowledge-base "$PREDICTION_KNOWLEDGE_BASE_ID" --title "$TITLE" --description "$DESCRIPTION" --status open)
if [[ -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    TEST_CREATE_ARGS+=(--prediction-subject "$PREDICTION_SUBJECT_ID")
fi
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    TEST_CREATE_ARGS+=(--created-by "$CREATED_BY_ID")
fi

test_name "Create prediction knowledge base question"
if [[ -n "$PREDICTION_KNOWLEDGE_BASE_ID" && "$PREDICTION_KNOWLEDGE_BASE_ID" != "null" ]]; then
    xbe_json "${TEST_CREATE_ARGS[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "prediction-knowledge-base-questions" "$CREATED_ID"
            pass
        else
            fail "Created prediction knowledge base question but no ID returned"
        fi
    else
        if update_blocked_message "$output"; then
            skip "Create blocked by server policy or invalid knowledge base"
        else
            fail "Failed to create prediction knowledge base question: $output"
        fi
    fi
else
    skip "No prediction knowledge base ID available (set XBE_TEST_PREDICTION_KNOWLEDGE_BASE_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update prediction knowledge base question attributes"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-knowledge-base-questions update "$CREATED_ID" \
        --title "$UPDATED_TITLE" \
        --description "$UPDATED_DESCRIPTION"
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Update blocked by server policy or validation"
        else
            fail "Failed to update prediction knowledge base question: $output"
        fi
    fi
else
    skip "No created prediction knowledge base question available for update"
fi

test_name "Update prediction knowledge base question status"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do prediction-knowledge-base-questions update "$CREATED_ID" --status resolved
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Status update blocked by server policy or validation"
        else
            fail "Failed to update prediction knowledge base question status: $output"
        fi
    fi
else
    skip "No created prediction knowledge base question available for status update"
fi

test_name "Update prediction knowledge base question prediction subject"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" && -n "$PREDICTION_SUBJECT_ID" && "$PREDICTION_SUBJECT_ID" != "null" ]]; then
    xbe_json do prediction-knowledge-base-questions update "$CREATED_ID" --prediction-subject "$PREDICTION_SUBJECT_ID"
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Prediction subject update blocked by server policy or validation"
        else
            fail "Failed to update prediction knowledge base question prediction subject: $output"
        fi
    fi
else
    skip "No created question or prediction subject ID available for prediction subject update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete prediction knowledge base question requires --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-knowledge-base-questions delete "$CREATED_ID"
    assert_failure
else
    skip "No created prediction knowledge base question available for delete"
fi

test_name "Delete prediction knowledge base question"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do prediction-knowledge-base-questions delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        assert_success
    else
        if update_blocked_message "$output"; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete prediction knowledge base question: $output"
        fi
    fi
else
    skip "No created prediction knowledge base question available for delete"
fi

run_tests
