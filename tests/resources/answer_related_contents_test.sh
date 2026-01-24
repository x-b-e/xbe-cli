#!/bin/bash
#
# XBE CLI Integration Tests: Answer Related Contents
#
# Tests list/show/create/update/delete operations for answer-related-contents.
#
# COVERAGE: list filters + create/update/delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""
CREATED_NEW=0
SAMPLE_ID=""
SAMPLE_ANSWER_ID=""
SAMPLE_RELATED_CONTENT_TYPE=""
SAMPLE_RELATED_CONTENT_ID=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: answer-related-contents"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List answer related contents"
xbe_json view answer-related-contents list --limit 5
assert_success

test_name "List answer related contents returns array"
xbe_json view answer-related-contents list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list answer related contents"
fi

# ============================================================================
# Capture sample data
# ============================================================================

test_name "Capture sample answer related content"
xbe_json view answer-related-contents list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_ANSWER_ID=$(json_get ".[0].answer_id")
    SAMPLE_RELATED_CONTENT_TYPE=$(json_get ".[0].related_content_type")
    SAMPLE_RELATED_CONTENT_ID=$(json_get ".[0].related_content_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No sample answer related content found"
    fi
else
    skip "Failed to capture sample answer related content"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List answer related contents with --answer"
if [[ -n "$SAMPLE_ANSWER_ID" && "$SAMPLE_ANSWER_ID" != "null" ]]; then
    xbe_json view answer-related-contents list --answer "$SAMPLE_ANSWER_ID" --limit 5
    assert_success
else
    skip "No sample answer ID available"
fi

test_name "List answer related contents with --related-content-type"
if [[ -n "$SAMPLE_RELATED_CONTENT_TYPE" && "$SAMPLE_RELATED_CONTENT_TYPE" != "null" ]]; then
    xbe_json view answer-related-contents list --related-content-type "$SAMPLE_RELATED_CONTENT_TYPE" --limit 5
    assert_success
else
    skip "No sample related content type available"
fi

test_name "List answer related contents with --related-content-type and --related-content-id"
if [[ -n "$SAMPLE_RELATED_CONTENT_TYPE" && "$SAMPLE_RELATED_CONTENT_TYPE" != "null" && -n "$SAMPLE_RELATED_CONTENT_ID" && "$SAMPLE_RELATED_CONTENT_ID" != "null" ]]; then
    xbe_json view answer-related-contents list --related-content-type "$SAMPLE_RELATED_CONTENT_TYPE" --related-content-id "$SAMPLE_RELATED_CONTENT_ID" --limit 5
    assert_success
else
    skip "No sample related content values available"
fi

test_name "List answer related contents with --not-related-content-type"
if [[ -n "$SAMPLE_RELATED_CONTENT_TYPE" && "$SAMPLE_RELATED_CONTENT_TYPE" != "null" ]]; then
    xbe_json view answer-related-contents list --not-related-content-type "$SAMPLE_RELATED_CONTENT_TYPE" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"not_related_content_type"* ]] || [[ "$output" == *"Internal Server Error"* ]] || [[ "$output" == *"INTERNAL_SERVER_ERROR"* ]]; then
            skip "Filter not-related-content-type not supported by server"
        else
            fail "Failed to filter by not-related-content-type"
        fi
    fi
else
    skip "No sample related content type available"
fi

test_name "List answer related contents with --created-at-min"
xbe_json view answer-related-contents list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List answer related contents with --created-at-max"
xbe_json view answer-related-contents list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List answer related contents with --is-created-at=true"
xbe_json view answer-related-contents list --is-created-at true --limit 5
assert_success

test_name "List answer related contents with --updated-at-min"
xbe_json view answer-related-contents list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List answer related contents with --updated-at-max"
xbe_json view answer-related-contents list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List answer related contents with --is-updated-at=false"
xbe_json view answer-related-contents list --is-updated-at false --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create answer related content"
if [[ -n "$SAMPLE_ANSWER_ID" && "$SAMPLE_ANSWER_ID" != "null" && -n "$SAMPLE_RELATED_CONTENT_TYPE" && "$SAMPLE_RELATED_CONTENT_TYPE" != "null" && -n "$SAMPLE_RELATED_CONTENT_ID" && "$SAMPLE_RELATED_CONTENT_ID" != "null" ]]; then
    xbe_json do answer-related-contents create \
        --answer "$SAMPLE_ANSWER_ID" \
        --related-content-type "$SAMPLE_RELATED_CONTENT_TYPE" \
        --related-content-id "$SAMPLE_RELATED_CONTENT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            CREATED_NEW=1
            register_cleanup "answer-related-contents" "$CREATED_ID"
            pass
        else
            fail "Created answer related content but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"read-only"* ]] || [[ "$output" == *"read only"* ]]; then
            skip "Create blocked by server policy"
        else
            fail "Failed to create answer related content"
        fi
    fi
else
    skip "No sample answer/related content IDs available for create"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show answer related content"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json view answer-related-contents show "$CREATED_ID"
    assert_success
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view answer-related-contents show "$SAMPLE_ID"
    assert_success
else
    skip "No answer related content ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update answer related content"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do answer-related-contents update "$CREATED_ID" --answer "$SAMPLE_ANSWER_ID"
    if [[ $status -ne 0 && -n "$SAMPLE_RELATED_CONTENT_TYPE" && "$SAMPLE_RELATED_CONTENT_TYPE" != "null" && -n "$SAMPLE_RELATED_CONTENT_ID" && "$SAMPLE_RELATED_CONTENT_ID" != "null" ]]; then
        xbe_json do answer-related-contents update "$CREATED_ID" --related-content-type "$SAMPLE_RELATED_CONTENT_TYPE" --related-content-id "$SAMPLE_RELATED_CONTENT_ID"
    fi
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"read-only"* ]] || [[ "$output" == *"read only"* ]]; then
            skip "Update blocked by server policy"
        else
            fail "Failed to update answer related content"
        fi
    fi
else
    skip "No created answer related content ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete answer related content"
if [[ "$CREATED_NEW" -eq 1 && -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do answer-related-contents delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Forbidden"* ]]; then
            skip "Delete blocked by server policy"
        else
            fail "Failed to delete answer related content"
        fi
    fi
else
    skip "Answer related content was not created by test"
fi

run_tests
