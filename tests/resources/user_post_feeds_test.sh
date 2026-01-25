#!/bin/bash
#
# XBE CLI Integration Tests: User Post Feeds
#
# Tests list/show/create/update/delete operations for user_post_feeds.
#
# COVERAGE: List filters + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FEED_ID=""
CREATED_FEED_NEW=0
EXISTING_FEED_ID=""
WHOAMI_USER_ID=""
CURRENT_VECTOR_INDEXING=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: user_post_feeds"

# ============================================================================
# Resolve current user
# ============================================================================

test_name "Get current user for user post feed tests"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    WHOAMI_USER_ID=$(json_get ".id")
    if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
        pass
    else
        fail "Whoami returned no user ID"
    fi
else
    if [[ -n "$XBE_TEST_USER_ID" ]]; then
        WHOAMI_USER_ID="$XBE_TEST_USER_ID"
        pass
    else
        skip "Unable to resolve current user"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user post feeds"
xbe_json view user-post-feeds list --limit 5
assert_success

test_name "List user post feeds returns array"
xbe_json view user-post-feeds list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user post feeds"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user post feeds with --user"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view user-post-feeds list --user "$WHOAMI_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List user post feeds with --created-at-min"
xbe_json view user-post-feeds list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List user post feeds with --created-at-max"
xbe_json view user-post-feeds list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List user post feeds with --is-created-at=true"
xbe_json view user-post-feeds list --is-created-at true --limit 5
assert_success

test_name "List user post feeds with --is-created-at=false"
xbe_json view user-post-feeds list --is-created-at false --limit 5
assert_success

test_name "List user post feeds with --updated-at-min"
xbe_json view user-post-feeds list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List user post feeds with --updated-at-max"
xbe_json view user-post-feeds list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List user post feeds with --is-updated-at=true"
xbe_json view user-post-feeds list --is-updated-at true --limit 5
assert_success

test_name "List user post feeds with --is-updated-at=false"
xbe_json view user-post-feeds list --is-updated-at false --limit 5
assert_success

# ============================================================================
# Capture existing feed (for show/update fallback)
# ============================================================================

test_name "Capture existing user post feed"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view user-post-feeds list --user "$WHOAMI_USER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        EXISTING_FEED_ID=$(json_get ".[0].id")
        if [[ -n "$EXISTING_FEED_ID" && "$EXISTING_FEED_ID" != "null" ]]; then
            pass
        else
            skip "No existing feed found for current user"
        fi
    else
        skip "Could not list feeds to capture sample"
    fi
else
    skip "No user ID available for sample feed"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create user post feed"
xbe_json do user-post-feeds create
if [[ $status -eq 0 ]]; then
    CREATED_FEED_ID=$(json_get ".id")
    if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
        CREATED_FEED_NEW=1
        register_cleanup "user-post-feeds" "$CREATED_FEED_ID"
        pass
    else
        fail "Created user post feed but no ID returned"
    fi
else
    if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
        xbe_json do user-post-feeds create --user "$WHOAMI_USER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_FEED_ID=$(json_get ".id")
            if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
                CREATED_FEED_NEW=1
                register_cleanup "user-post-feeds" "$CREATED_FEED_ID"
                pass
            else
                fail "Created user post feed but no ID returned"
            fi
        else
            if [[ -n "$EXISTING_FEED_ID" && "$EXISTING_FEED_ID" != "null" ]]; then
                CREATED_FEED_ID="$EXISTING_FEED_ID"
                CREATED_FEED_NEW=0
                echo "    Using existing feed: $CREATED_FEED_ID"
                pass
            else
                fail "Failed to create user post feed"
            fi
        fi
    else
        if [[ -n "$EXISTING_FEED_ID" && "$EXISTING_FEED_ID" != "null" ]]; then
            CREATED_FEED_ID="$EXISTING_FEED_ID"
            CREATED_FEED_NEW=0
            echo "    Using existing feed: $CREATED_FEED_ID"
            pass
        else
            fail "Failed to create user post feed"
        fi
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user post feed"
if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    xbe_json view user-post-feeds show "$CREATED_FEED_ID"
    if [[ $status -eq 0 ]]; then
        CURRENT_VECTOR_INDEXING=$(json_get ".enable_vector_indexing")
        pass
    else
        fail "Failed to show user post feed"
    fi
else
    skip "No user post feed ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user post feed"
if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    if [[ -z "$CURRENT_VECTOR_INDEXING" || "$CURRENT_VECTOR_INDEXING" == "null" ]]; then
        xbe_json view user-post-feeds show "$CREATED_FEED_ID"
        if [[ $status -eq 0 ]]; then
            CURRENT_VECTOR_INDEXING=$(json_get ".enable_vector_indexing")
        fi
    fi

    if [[ "$CURRENT_VECTOR_INDEXING" == "true" || "$CURRENT_VECTOR_INDEXING" == "false" ]]; then
        xbe_json do user-post-feeds update "$CREATED_FEED_ID" --enable-vector-indexing "$CURRENT_VECTOR_INDEXING"
        assert_success
    else
        fail "Unable to determine enable_vector_indexing state"
    fi
else
    skip "No user post feed ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete user post feed"
if [[ "$CREATED_FEED_NEW" -eq 1 && -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    xbe_run do user-post-feeds delete "$CREATED_FEED_ID" --confirm
    assert_success
else
    skip "Feed was not created by test run"
fi

run_tests
