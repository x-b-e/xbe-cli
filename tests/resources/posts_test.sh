#!/bin/bash
#
# XBE CLI Integration Tests: Posts
#
# Tests CRUD operations for the posts resource.
# Posts are content items with various types and statuses.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_POST_ID=""

describe "Resource: posts"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create post with required fields"
xbe_json do posts create --post-type "basic" --text-content "Test post content $(date +%s)"

if [[ $status -eq 0 ]]; then
    CREATED_POST_ID=$(json_get ".id")
    if [[ -n "$CREATED_POST_ID" && "$CREATED_POST_ID" != "null" ]]; then
        register_cleanup "posts" "$CREATED_POST_ID"
        pass
    else
        fail "Created post but no ID returned"
    fi
else
    fail "Failed to create post"
fi

# Only continue if we successfully created a post
if [[ -z "$CREATED_POST_ID" || "$CREATED_POST_ID" == "null" ]]; then
    echo "Cannot continue without a valid post ID"
    run_tests
fi

test_name "Create post with status"
xbe_json do posts create --post-type "basic" --text-content "Published post $(date +%s)" --status "published"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "posts" "$id"
    pass
else
    fail "Failed to create post with status"
fi

test_name "Create private post"
xbe_json do posts create --post-type "basic" --text-content "Private post $(date +%s)" --private

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "posts" "$id"
    pass
else
    fail "Failed to create private post"
fi

# ============================================================================
# UPDATE Tests - String Attributes
# ============================================================================

test_name "Update post text-content"
xbe_json do posts update "$CREATED_POST_ID" --text-content "Updated content $(date +%s)"
assert_success

test_name "Update post status to published"
xbe_json do posts update "$CREATED_POST_ID" --status "published"
assert_success

test_name "Update post status to draft"
xbe_json do posts update "$CREATED_POST_ID" --status "draft"
assert_success

test_name "Update post post-type"
xbe_json do posts update "$CREATED_POST_ID" --post-type "notification"
assert_success

# ============================================================================
# UPDATE Tests - Boolean Attributes
# ============================================================================

test_name "Update post private to true"
xbe_json do posts update "$CREATED_POST_ID" --private true
assert_success

test_name "Update post private to false"
xbe_json do posts update "$CREATED_POST_ID" --private false
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List posts"
xbe_json view posts list --limit 5
assert_success

test_name "List posts returns array"
xbe_json view posts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list posts"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List posts with --post-type filter"
xbe_json view posts list --post-type "basic" --limit 5
assert_success

test_name "List posts with --status filter"
xbe_json view posts list --status "published" --limit 5
assert_success

test_name "List posts with --is-private filter (true)"
xbe_json view posts list --is-private true --limit 5
assert_success

test_name "List posts with --is-private filter (false)"
xbe_json view posts list --is-private false --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List posts with --limit"
xbe_json view posts list --limit 3
assert_success

test_name "List posts with --offset"
xbe_json view posts list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete post requires --confirm flag"
xbe_json do posts delete "$CREATED_POST_ID"
assert_failure

test_name "Delete post with --confirm"
# Create a post specifically for deletion
xbe_json do posts create --post-type "basic" --text-content "Delete me $(date +%s)"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do posts delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create post for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create post without post-type fails"
xbe_json do posts create --text-content "Missing type"
assert_failure

test_name "Create post without text-content fails"
xbe_json do posts create --post-type "basic"
assert_failure

test_name "Update without any fields fails"
xbe_json do posts update "$CREATED_POST_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
