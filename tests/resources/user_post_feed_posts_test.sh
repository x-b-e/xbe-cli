#!/bin/bash
#
# XBE CLI Integration Tests: User Post Feed Posts
#
# Tests list, show, and update operations for the user-post-feed-posts resource.
#
# COVERAGE: List filters + show + update
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

USER_POST_FEED_POST_ID=""
USER_ID=""
USER_POST_FEED_ID=""
POST_ID=""
FOLLOW_ID=""
SCORE=""
FEED_AT=""
SUBSCRIPTION_START_AT=""
SUBSCRIPTION_END_AT=""
POST_TYPE=""
CREATOR_USER_ID=""
CREATOR_FILTER=""
CURRENT_USER_ID=""
UPDATE_FEED_POST_ID=""
SKIP_SHOW=0
SKIP_UPDATE=0

DEFAULT_DATE_MIN="2024-01-01T00:00:00Z"
DEFAULT_DATE_MAX="2024-12-31T23:59:59Z"

describe "Resource: user-post-feed-posts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user post feed posts"
xbe_json view user-post-feed-posts list --limit 5
assert_success

test_name "List user post feed posts returns array"
xbe_json view user-post-feed-posts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user post feed posts"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample user post feed post"
xbe_json view user-post-feed-posts list --limit 1
if [[ $status -eq 0 ]]; then
    USER_POST_FEED_POST_ID=$(json_get ".[0].id")
    USER_ID=$(json_get ".[0].user_id")
    USER_POST_FEED_ID=$(json_get ".[0].user_post_feed_id")
    POST_ID=$(json_get ".[0].post_id")
    FOLLOW_ID=$(json_get ".[0].follow_id")
    SCORE=$(json_get ".[0].score")
    FEED_AT=$(json_get ".[0].feed_at")
    SUBSCRIPTION_START_AT=$(json_get ".[0].subscription_start_at")
    SUBSCRIPTION_END_AT=$(json_get ".[0].subscription_end_at")
    if [[ -n "$USER_POST_FEED_POST_ID" && "$USER_POST_FEED_POST_ID" != "null" ]]; then
        pass
    else
        SKIP_SHOW=1
        SKIP_UPDATE=1
        skip "No user post feed posts available"
    fi
else
    SKIP_SHOW=1
    SKIP_UPDATE=1
    fail "Failed to list user post feed posts"
fi

# ============================================================================
# Current User
# ============================================================================

test_name "Resolve current user"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No current user ID available"
    fi
else
    skip "Failed to resolve current user"
fi

# ============================================================================
# Sample Post Data
# ============================================================================

test_name "Find sample post for filters"
xbe_json view posts list --limit 1
if [[ $status -eq 0 ]]; then
    POST_ID=$(json_get ".[0].id")
    POST_TYPE=$(json_get ".[0].post_type")
    if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
        pass
    else
        skip "No post available"
    fi
else
    skip "Failed to list posts"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user post feed posts with --user filter"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json view user-post-feed-posts list --user "$CURRENT_USER_ID" --limit 5
    assert_success
else
    skip "No current user ID available"
fi

test_name "List user post feed posts with --user-id filter"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json view user-post-feed-posts list --user-id "$CURRENT_USER_ID" --limit 5
    assert_success
else
    skip "No current user ID available"
fi

test_name "List user post feed posts with --post filter"
if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    xbe_json view user-post-feed-posts list --post "$POST_ID" --limit 5
    assert_success
else
    skip "No post ID available"
fi

test_name "List user post feed posts with --post-type filter"
if [[ -n "$POST_TYPE" && "$POST_TYPE" != "null" ]]; then
    xbe_json view user-post-feed-posts list --post-type "$POST_TYPE" --limit 5
    assert_success
else
    skip "No post type available"
fi

test_name "Resolve creator ID for --creator filter"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    CREATOR_USER_ID="$CURRENT_USER_ID"
    CREATOR_FILTER="User|${CREATOR_USER_ID}"
    pass
else
    xbe_json view users list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATOR_USER_ID=$(json_get ".[0].id")
        if [[ -n "$CREATOR_USER_ID" && "$CREATOR_USER_ID" != "null" ]]; then
            CREATOR_FILTER="User|${CREATOR_USER_ID}"
            pass
        else
            skip "No user ID available"
        fi
    else
        skip "Failed to load users"
    fi
fi

test_name "List user post feed posts with --creator filter"
if [[ -n "$CREATOR_FILTER" && "$CREATOR_FILTER" != "null" ]]; then
    xbe_json view user-post-feed-posts list --creator "$CREATOR_FILTER" --limit 5
    assert_success
else
    skip "No creator filter available"
fi

test_name "Find user post feed post for update"
if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    xbe_json view user-post-feed-posts list --user "$CURRENT_USER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        UPDATE_FEED_POST_ID=$(json_get ".[0].id")
        if [[ -n "$UPDATE_FEED_POST_ID" && "$UPDATE_FEED_POST_ID" != "null" ]]; then
            pass
        else
            SKIP_UPDATE=1
            skip "No feed posts available for current user"
        fi
    else
        SKIP_UPDATE=1
        skip "Failed to list user post feed posts for update"
    fi
else
    SKIP_UPDATE=1
    skip "No current user ID available"
fi

test_name "List user post feed posts with --score filter"
if [[ -n "$SCORE" && "$SCORE" != "null" ]]; then
    xbe_json view user-post-feed-posts list --score "$SCORE" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --score 0.5 --limit 5
    assert_success
fi

test_name "List user post feed posts with --feed-at-min filter"
if [[ -n "$FEED_AT" && "$FEED_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --feed-at-min "$FEED_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --feed-at-min "$DEFAULT_DATE_MIN" --limit 5
    assert_success
fi

test_name "List user post feed posts with --feed-at-max filter"
if [[ -n "$FEED_AT" && "$FEED_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --feed-at-max "$FEED_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --feed-at-max "$DEFAULT_DATE_MAX" --limit 5
    assert_success
fi

test_name "List user post feed posts with --is-feed-at filter"
xbe_json view user-post-feed-posts list --is-feed-at true --limit 5
assert_success

test_name "List user post feed posts with --subscription-start-at-min filter"
if [[ -n "$SUBSCRIPTION_START_AT" && "$SUBSCRIPTION_START_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --subscription-start-at-min "$SUBSCRIPTION_START_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --subscription-start-at-min "$DEFAULT_DATE_MIN" --limit 5
    assert_success
fi

test_name "List user post feed posts with --subscription-start-at-max filter"
if [[ -n "$SUBSCRIPTION_START_AT" && "$SUBSCRIPTION_START_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --subscription-start-at-max "$SUBSCRIPTION_START_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --subscription-start-at-max "$DEFAULT_DATE_MAX" --limit 5
    assert_success
fi

test_name "List user post feed posts with --is-subscription-start-at filter"
xbe_json view user-post-feed-posts list --is-subscription-start-at true --limit 5
assert_success

test_name "List user post feed posts with --subscription-end-at-min filter"
if [[ -n "$SUBSCRIPTION_END_AT" && "$SUBSCRIPTION_END_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --subscription-end-at-min "$SUBSCRIPTION_END_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --subscription-end-at-min "$DEFAULT_DATE_MIN" --limit 5
    assert_success
fi

test_name "List user post feed posts with --subscription-end-at-max filter"
if [[ -n "$SUBSCRIPTION_END_AT" && "$SUBSCRIPTION_END_AT" != "null" ]]; then
    xbe_json view user-post-feed-posts list --subscription-end-at-max "$SUBSCRIPTION_END_AT" --limit 5
    assert_success
else
    xbe_json view user-post-feed-posts list --subscription-end-at-max "$DEFAULT_DATE_MAX" --limit 5
    assert_success
fi

test_name "List user post feed posts with --is-subscription-end-at filter"
xbe_json view user-post-feed-posts list --is-subscription-end-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user post feed post"
if [[ $SKIP_SHOW -eq 0 && -n "$USER_POST_FEED_POST_ID" && "$USER_POST_FEED_POST_ID" != "null" ]]; then
    xbe_json view user-post-feed-posts show "$USER_POST_FEED_POST_ID"
    assert_success
else
    skip "No user post feed post ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user post feed post"
if [[ $SKIP_UPDATE -eq 0 && -n "$UPDATE_FEED_POST_ID" && "$UPDATE_FEED_POST_ID" != "null" ]]; then
    xbe_json do user-post-feed-posts update "$UPDATE_FEED_POST_ID" \
        --is-bookmarked=true \
        --subscription-start-at "2025-01-01T00:00:00Z" \
        --subscription-end-at "2025-01-31T23:59:59Z"
    assert_success
else
    skip "No user post feed post ID available"
fi

test_name "Update user post feed post without fields fails"
if [[ $SKIP_UPDATE -eq 0 && -n "$UPDATE_FEED_POST_ID" && "$UPDATE_FEED_POST_ID" != "null" ]]; then
    xbe_json do user-post-feed-posts update "$UPDATE_FEED_POST_ID"
    assert_failure
else
    skip "No user post feed post ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
