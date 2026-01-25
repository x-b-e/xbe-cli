#!/bin/bash
#
# XBE CLI Integration Tests: Post Children
#
# Tests list, show, create, and delete operations for post-children.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEED_POST_CHILD_ID=""
SEED_PARENT_POST_ID=""
SEED_CHILD_POST_ID=""

PARENT_POST_ID="${XBE_TEST_PARENT_POST_ID:-}"
CHILD_POST_ID="${XBE_TEST_CHILD_POST_ID:-}"

CREATED_POST_CHILD_ID=""

is_forbidden() {
    [[ "$output" == *"USER_NOT_AUTHORIZED"* || "$output" == *"FORBIDDEN"* || "$output" == *"403"* ]]
}

describe "Resource: post-children"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List post children"
xbe_json view post-children list --limit 50
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    total=$(echo "$output" | jq 'length')
    if [[ "$total" -gt 0 ]]; then
        SEED_POST_CHILD_ID=$(echo "$output" | jq -r '.[0].id')
        SEED_PARENT_POST_ID=$(echo "$output" | jq -r '.[0].parent_post_id')
        SEED_CHILD_POST_ID=$(echo "$output" | jq -r '.[0].child_post_id')
    fi
else
    fail "Failed to list post children"
fi

# ============================================================================
# PREREQUISITES
# ============================================================================

test_name "Create prerequisite parent post"
if [[ -z "$PARENT_POST_ID" || "$PARENT_POST_ID" == "null" ]]; then
    xbe_json do posts create --post-type "basic" --text-content "Post child parent $(date +%s)"
    if [[ $status -eq 0 ]]; then
        PARENT_POST_ID=$(json_get ".id")
        if [[ -n "$PARENT_POST_ID" && "$PARENT_POST_ID" != "null" ]]; then
            register_cleanup "posts" "$PARENT_POST_ID"
            pass
        else
            fail "Created parent post but no ID returned"
            run_tests
        fi
    else
        if is_forbidden; then
            skip "No permission to create posts (set XBE_TEST_PARENT_POST_ID to continue)"
        else
            fail "Failed to create parent post"
            run_tests
        fi
    fi
else
    echo "    Using XBE_TEST_PARENT_POST_ID: $PARENT_POST_ID"
    pass
fi

test_name "Create prerequisite child post"
if [[ -z "$CHILD_POST_ID" || "$CHILD_POST_ID" == "null" ]]; then
    xbe_json do posts create --post-type "basic" --text-content "Post child child $(date +%s)"
    if [[ $status -eq 0 ]]; then
        CHILD_POST_ID=$(json_get ".id")
        if [[ -n "$CHILD_POST_ID" && "$CHILD_POST_ID" != "null" ]]; then
            register_cleanup "posts" "$CHILD_POST_ID"
            pass
        else
            fail "Created child post but no ID returned"
            run_tests
        fi
    else
        if is_forbidden; then
            skip "No permission to create posts (set XBE_TEST_CHILD_POST_ID to continue)"
        else
            fail "Failed to create child post"
            run_tests
        fi
    fi
else
    echo "    Using XBE_TEST_CHILD_POST_ID: $CHILD_POST_ID"
    pass
fi

if [[ -z "$PARENT_POST_ID" || "$PARENT_POST_ID" == "null" ]]; then
    PARENT_POST_ID="$SEED_PARENT_POST_ID"
fi
if [[ -z "$CHILD_POST_ID" || "$CHILD_POST_ID" == "null" ]]; then
    CHILD_POST_ID="$SEED_CHILD_POST_ID"
fi

if [[ -n "$PARENT_POST_ID" && -n "$CHILD_POST_ID" && "$PARENT_POST_ID" == "$CHILD_POST_ID" ]]; then
    echo "    Parent post ID and child post ID are the same; skipping create tests."
    PARENT_POST_ID=""
    CHILD_POST_ID=""
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create post child"
if [[ -n "$PARENT_POST_ID" && "$PARENT_POST_ID" != "null" && -n "$CHILD_POST_ID" && "$CHILD_POST_ID" != "null" ]]; then
    xbe_json do post-children create --parent-post "$PARENT_POST_ID" --child-post "$CHILD_POST_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_POST_CHILD_ID=$(json_get ".id")
        if [[ -n "$CREATED_POST_CHILD_ID" && "$CREATED_POST_CHILD_ID" != "null" ]]; then
            register_cleanup "post-children" "$CREATED_POST_CHILD_ID"
            pass
        else
            fail "Created post child but no ID returned"
        fi
    else
        if is_forbidden; then
            skip "No permission to create post children"
        else
            fail "Failed to create post child"
        fi
    fi
else
    skip "No parent/child post IDs available for creation"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show post child"
SHOW_POST_CHILD_ID="$CREATED_POST_CHILD_ID"
if [[ -z "$SHOW_POST_CHILD_ID" || "$SHOW_POST_CHILD_ID" == "null" ]]; then
    SHOW_POST_CHILD_ID="$SEED_POST_CHILD_ID"
fi

if [[ -n "$SHOW_POST_CHILD_ID" && "$SHOW_POST_CHILD_ID" != "null" ]]; then
    xbe_json view post-children show "$SHOW_POST_CHILD_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to show post child"
    fi
else
    skip "No post child available to show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by parent post"
if [[ -n "$PARENT_POST_ID" && "$PARENT_POST_ID" != "null" ]]; then
    xbe_json view post-children list --parent-post "$PARENT_POST_ID" --limit 5
    assert_success
elif [[ -n "$SEED_PARENT_POST_ID" && "$SEED_PARENT_POST_ID" != "null" ]]; then
    xbe_json view post-children list --parent-post "$SEED_PARENT_POST_ID" --limit 5
    assert_success
else
    skip "No parent post ID available for filter"
fi

test_name "Filter by child post"
if [[ -n "$CHILD_POST_ID" && "$CHILD_POST_ID" != "null" ]]; then
    xbe_json view post-children list --child-post "$CHILD_POST_ID" --limit 5
    assert_success
elif [[ -n "$SEED_CHILD_POST_ID" && "$SEED_CHILD_POST_ID" != "null" ]]; then
    xbe_json view post-children list --child-post "$SEED_CHILD_POST_ID" --limit 5
    assert_success
else
    skip "No child post ID available for filter"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete post child"
if [[ -n "$CREATED_POST_CHILD_ID" && "$CREATED_POST_CHILD_ID" != "null" ]]; then
    xbe_run do post-children delete "$CREATED_POST_CHILD_ID" --confirm
    assert_success
else
    skip "No created post child to delete"
fi

run_tests
