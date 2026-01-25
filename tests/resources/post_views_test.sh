#!/bin/bash
#
# XBE CLI Integration Tests: Post Views
#
# Tests list/show/create operations for post-views.
#
# COVERAGE: All list filters + create
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_POST_VIEW_ID=""
POST_ID="${XBE_TEST_POST_ID:-}"
VIEWER_USER_ID=""
CREATED_POST_VIEW_ID=""
CREATED_POST_ID=""

VIEWED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
CREATED_AT_FILTER="$VIEWED_AT_FILTER"
UPDATED_AT_FILTER="$VIEWED_AT_FILTER"

describe "Resource: post-views"

# ============================================================================
# Resolve current user
# ============================================================================

test_name "Get current user for post view tests"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    VIEWER_USER_ID=$(json_get ".id")
    if [[ -n "$VIEWER_USER_ID" && "$VIEWER_USER_ID" != "null" ]]; then
        pass
    else
        fail "Whoami returned no user ID"
    fi
else
    if [[ -n "$XBE_TEST_USER_ID" ]]; then
        VIEWER_USER_ID="$XBE_TEST_USER_ID"
        pass
    else
        skip "Unable to resolve current user"
    fi
fi

# ============================================================================
# Prerequisites - Create post if needed
# ============================================================================

if [[ -z "$POST_ID" || "$POST_ID" == "null" ]]; then
    test_name "Create prerequisite post for post view tests"
    POST_TEXT=$(unique_name "PostViewTest")

    xbe_json do posts create --post-type basic --text-content "Post view test ${POST_TEXT}"

    if [[ $status -eq 0 ]]; then
        CREATED_POST_ID=$(json_get ".id")
        if [[ -n "$CREATED_POST_ID" && "$CREATED_POST_ID" != "null" ]]; then
            register_cleanup "posts" "$CREATED_POST_ID"
            POST_ID="$CREATED_POST_ID"
            pass
        else
            fail "Created post but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_POST_ID" ]]; then
            POST_ID="$XBE_TEST_POST_ID"
            pass
        else
            xbe_json view posts list --limit 1
            if [[ $status -eq 0 ]]; then
                assert_json_is_array
                total=$(echo "$output" | jq 'length')
                if [[ "$total" -gt 0 ]]; then
                    POST_ID=$(echo "$output" | jq -r '.[0].id')
                    pass
                else
                    skip "No posts available"
                fi
            else
                skip "Failed to create or list posts"
            fi
        fi
    fi
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List post views"
xbe_json view post-views list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_POST_VIEW_ID=$(echo "$output" | jq -r '.[0].id')
        if [[ -z "$POST_ID" || "$POST_ID" == "null" ]]; then
            POST_ID=$(echo "$output" | jq -r '.[0].post_id')
        fi
        if [[ -z "$VIEWER_USER_ID" || "$VIEWER_USER_ID" == "null" ]]; then
            VIEWER_USER_ID=$(echo "$output" | jq -r '.[0].viewer_id')
        fi
    fi
else
    fail "Failed to list post views"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show post view"
SHOW_ID="${SEED_POST_VIEW_ID:-$CREATED_POST_VIEW_ID}"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view post-views show "$SHOW_ID"
    assert_success
else
    skip "No post view available to show"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create post view"
if [[ -n "$POST_ID" && "$POST_ID" != "null" && -n "$VIEWER_USER_ID" && "$VIEWER_USER_ID" != "null" ]]; then
    xbe_json do post-views create --post "$POST_ID" --viewer "$VIEWER_USER_ID" --viewed-at "$VIEWED_AT_FILTER"
    if [[ $status -eq 0 ]]; then
        CREATED_POST_VIEW_ID=$(json_get ".id")
        if [[ -n "$CREATED_POST_VIEW_ID" && "$CREATED_POST_VIEW_ID" != "null" ]]; then
            pass
        else
            fail "Created post view but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"unprocessable"* ]] || [[ "$output" == *"Validation"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create post view: $output"
        fi
    fi
else
    skip "Missing post ID or viewer ID for creation"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List post views with --post"
if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    xbe_json view post-views list --post "$POST_ID" --limit 5
    assert_success
else
    skip "No post ID available for --post filter"
fi

test_name "List post views with --viewer"
if [[ -n "$VIEWER_USER_ID" && "$VIEWER_USER_ID" != "null" ]]; then
    xbe_json view post-views list --viewer "$VIEWER_USER_ID" --limit 5
    assert_success
else
    skip "No viewer ID available for --viewer filter"
fi

test_name "List post views with --viewed-at-min"
xbe_json view post-views list --viewed-at-min "$VIEWED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --viewed-at-max"
xbe_json view post-views list --viewed-at-max "$VIEWED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --is-viewed-at=true"
xbe_json view post-views list --is-viewed-at true --limit 5
assert_success

test_name "List post views with --is-viewed-at=false"
xbe_json view post-views list --is-viewed-at false --limit 5
assert_success

test_name "List post views with --created-at-min"
xbe_json view post-views list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --created-at-max"
xbe_json view post-views list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --is-created-at=true"
xbe_json view post-views list --is-created-at true --limit 5
assert_success

test_name "List post views with --is-created-at=false"
xbe_json view post-views list --is-created-at false --limit 5
assert_success

test_name "List post views with --updated-at-min"
xbe_json view post-views list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --updated-at-max"
xbe_json view post-views list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List post views with --is-updated-at=true"
xbe_json view post-views list --is-updated-at true --limit 5
assert_success

test_name "List post views with --is-updated-at=false"
xbe_json view post-views list --is-updated-at false --limit 5
assert_success

run_tests
