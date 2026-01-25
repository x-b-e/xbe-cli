#!/bin/bash
#
# XBE CLI Integration Tests: Comment Reactions
#
# Tests list, show, create, and delete operations for comment-reactions.
#
# COVERAGE: All list filters + create + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_COMMENT_REACTION_ID=""
COMMENT_ID="${XBE_TEST_COMMENT_ID:-}"
REACTION_CLASSIFICATION_ID="${XBE_TEST_REACTION_CLASSIFICATION_ID:-}"
CREATED_COMMENT_REACTION_ID=""

CREATED_BROKER_ID=""
CREATED_DEVELOPER_ID=""
CREATED_PROJECT_ID=""
CREATED_COMMENT_ID=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: comment-reactions"

# ============================================================================
# Prerequisites - Create comment if needed
# ============================================================================

if [[ -z "$COMMENT_ID" || "$COMMENT_ID" == "null" ]]; then
    test_name "Create prerequisite broker for comment reaction tests"
    BROKER_NAME=$(unique_name "CommentReactionTestBroker")

    xbe_json do brokers create --name "$BROKER_NAME"

    if [[ $status -eq 0 ]]; then
        CREATED_BROKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
            register_cleanup "brokers" "$CREATED_BROKER_ID"
            pass
        else
            fail "Created broker but no ID returned"
            run_tests
        fi
    else
        if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
            CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
            echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
            pass
        else
            fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
            run_tests
        fi
    fi

    test_name "Create prerequisite developer for comment reaction tests"
    DEVELOPER_NAME=$(unique_name "CommentReactionTestDev")

    xbe_json do developers create \
        --name "$DEVELOPER_NAME" \
        --broker "$CREATED_BROKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_DEVELOPER_ID=$(json_get ".id")
        if [[ -n "$CREATED_DEVELOPER_ID" && "$CREATED_DEVELOPER_ID" != "null" ]]; then
            register_cleanup "developers" "$CREATED_DEVELOPER_ID"
            pass
        else
            fail "Created developer but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create developer"
        run_tests
    fi

    test_name "Create prerequisite project for comment reaction tests"
    PROJECT_NAME=$(unique_name "CommentReactionTestProject")

    xbe_json do projects create \
        --name "$PROJECT_NAME" \
        --developer "$CREATED_DEVELOPER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_PROJECT_ID=$(json_get ".id")
        if [[ -n "$CREATED_PROJECT_ID" && "$CREATED_PROJECT_ID" != "null" ]]; then
            register_cleanup "projects" "$CREATED_PROJECT_ID"
            pass
        else
            fail "Created project but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create project"
        run_tests
    fi

    test_name "Create prerequisite comment for comment reaction tests"
    xbe_json do comments create \
        --body "Test comment for reactions" \
        --commentable-type "projects" \
        --commentable-id "$CREATED_PROJECT_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_COMMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_COMMENT_ID" && "$CREATED_COMMENT_ID" != "null" ]]; then
            register_cleanup "comments" "$CREATED_COMMENT_ID"
            COMMENT_ID="$CREATED_COMMENT_ID"
            pass
        else
            fail "Created comment but no ID returned"
            run_tests
        fi
    else
        fail "Failed to create comment"
        run_tests
    fi
fi

# Resolve a reaction classification if needed
if [[ -z "$REACTION_CLASSIFICATION_ID" || "$REACTION_CLASSIFICATION_ID" == "null" ]]; then
    test_name "Resolve reaction classification"
    xbe_json view reaction-classifications list --limit 1
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
        total=$(echo "$output" | jq 'length')
        if [[ "$total" -gt 0 ]]; then
            REACTION_CLASSIFICATION_ID=$(echo "$output" | jq -r '.[0].id')
            pass
        else
            skip "No reaction classifications available"
        fi
    else
        skip "Failed to list reaction classifications"
    fi
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List comment reactions"
xbe_json view comment-reactions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_COMMENT_REACTION_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$COMMENT_ID" || "$COMMENT_ID" == "null" ]]; then
            COMMENT_ID=$(echo "$output" | jq -r '.[0].comment_id')
        fi
        if [[ -z "$REACTION_CLASSIFICATION_ID" || "$REACTION_CLASSIFICATION_ID" == "null" ]]; then
            REACTION_CLASSIFICATION_ID=$(echo "$output" | jq -r '.[0].reaction_classification_id')
        fi
    fi
else
    fail "Failed to list comment reactions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show comment reaction"
if [[ -n "$SEED_COMMENT_REACTION_ID" && "$SEED_COMMENT_REACTION_ID" != "null" ]]; then
    xbe_json view comment-reactions show "$SEED_COMMENT_REACTION_ID"
    assert_success
else
    skip "No comment reaction available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create comment reaction"
if [[ -n "$COMMENT_ID" && "$COMMENT_ID" != "null" && -n "$REACTION_CLASSIFICATION_ID" && "$REACTION_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json do comment-reactions create --comment "$COMMENT_ID" --reaction-classification "$REACTION_CLASSIFICATION_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_COMMENT_REACTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_COMMENT_REACTION_ID" && "$CREATED_COMMENT_REACTION_ID" != "null" ]]; then
            register_cleanup "comment-reactions" "$CREATED_COMMENT_REACTION_ID"
            pass
        else
            fail "Created comment reaction but no ID returned"
        fi
    else
        if [[ "$output" == *"already exists"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"unprocessable"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create comment reaction: $output"
        fi
    fi
else
    skip "Missing comment ID or reaction classification ID for creation"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete comment reaction"
if [[ -n "$CREATED_COMMENT_REACTION_ID" && "$CREATED_COMMENT_REACTION_ID" != "null" ]]; then
    xbe_run do comment-reactions delete "$CREATED_COMMENT_REACTION_ID" --confirm
    assert_success
else
    skip "No created comment reaction to delete"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List comment reactions with --created-at-min"
xbe_json view comment-reactions list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List comment reactions with --created-at-max"
xbe_json view comment-reactions list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List comment reactions with --is-created-at=true"
xbe_json view comment-reactions list --is-created-at true --limit 5
assert_success

test_name "List comment reactions with --is-created-at=false"
xbe_json view comment-reactions list --is-created-at false --limit 5
assert_success

test_name "List comment reactions with --updated-at-min"
xbe_json view comment-reactions list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List comment reactions with --updated-at-max"
xbe_json view comment-reactions list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List comment reactions with --is-updated-at=true"
xbe_json view comment-reactions list --is-updated-at true --limit 5
assert_success

test_name "List comment reactions with --is-updated-at=false"
xbe_json view comment-reactions list --is-updated-at false --limit 5
assert_success

test_name "List comment reactions with --not-id"
NOT_ID_TARGET="${CREATED_COMMENT_REACTION_ID:-$SEED_COMMENT_REACTION_ID}"
if [[ -n "$NOT_ID_TARGET" && "$NOT_ID_TARGET" != "null" ]]; then
    xbe_json view comment-reactions list --not-id "$NOT_ID_TARGET" --limit 5
    assert_success
else
    skip "No comment reaction ID available for --not-id filter"
fi

run_tests
