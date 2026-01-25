#!/bin/bash
#
# XBE CLI Integration Tests: User Creator Feeds
#
# Tests list/show/create/update/delete operations for user_creator_feeds.
#
# COVERAGE: List filters + create/update attributes + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_FEED_ID=""
CREATED_FEED_NEW=0
EXISTING_FEED_ID=""
WHOAMI_USER_ID=""

CREATED_AT_FILTER="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
UPDATED_AT_FILTER="$CREATED_AT_FILTER"

describe "Resource: user_creator_feeds"

# ============================================================================
# Resolve current user
# ============================================================================

test_name "Get current user for user creator feed tests"
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

test_name "List user creator feeds"
xbe_json view user-creator-feeds list --limit 5
assert_success

test_name "List user creator feeds returns array"
xbe_json view user-creator-feeds list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user creator feeds"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user creator feeds with --user"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view user-creator-feeds list --user "$WHOAMI_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available for filter test"
fi

test_name "List user creator feeds with --created-at-min"
xbe_json view user-creator-feeds list --created-at-min "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List user creator feeds with --created-at-max"
xbe_json view user-creator-feeds list --created-at-max "$CREATED_AT_FILTER" --limit 5
assert_success

test_name "List user creator feeds with --is-created-at=true"
xbe_json view user-creator-feeds list --is-created-at true --limit 5
assert_success

test_name "List user creator feeds with --is-created-at=false"
xbe_json view user-creator-feeds list --is-created-at false --limit 5
assert_success

test_name "List user creator feeds with --updated-at-min"
xbe_json view user-creator-feeds list --updated-at-min "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List user creator feeds with --updated-at-max"
xbe_json view user-creator-feeds list --updated-at-max "$UPDATED_AT_FILTER" --limit 5
assert_success

test_name "List user creator feeds with --is-updated-at=true"
xbe_json view user-creator-feeds list --is-updated-at true --limit 5
assert_success

test_name "List user creator feeds with --is-updated-at=false"
xbe_json view user-creator-feeds list --is-updated-at false --limit 5
assert_success

# ============================================================================
# Capture existing feed (for show/update fallback)
# ============================================================================

test_name "Capture existing user creator feed"
if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
    xbe_json view user-creator-feeds list --user "$WHOAMI_USER_ID" --limit 1
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

test_name "Create user creator feed"
xbe_json do user-creator-feeds create
if [[ $status -eq 0 ]]; then
    CREATED_FEED_ID=$(json_get ".id")
    if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
        CREATED_FEED_NEW=1
        register_cleanup "user-creator-feeds" "$CREATED_FEED_ID"
        pass
    else
        fail "Created user creator feed but no ID returned"
    fi
else
    if [[ -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
        xbe_json do user-creator-feeds create --user "$WHOAMI_USER_ID"
        if [[ $status -eq 0 ]]; then
            CREATED_FEED_ID=$(json_get ".id")
            if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
                CREATED_FEED_NEW=1
                register_cleanup "user-creator-feeds" "$CREATED_FEED_ID"
                pass
            else
                fail "Created user creator feed but no ID returned"
            fi
        else
            if [[ -n "$EXISTING_FEED_ID" && "$EXISTING_FEED_ID" != "null" ]]; then
                CREATED_FEED_ID="$EXISTING_FEED_ID"
                CREATED_FEED_NEW=0
                echo "    Using existing feed: $CREATED_FEED_ID"
                pass
            else
                fail "Failed to create user creator feed"
            fi
        fi
    else
        if [[ -n "$EXISTING_FEED_ID" && "$EXISTING_FEED_ID" != "null" ]]; then
            CREATED_FEED_ID="$EXISTING_FEED_ID"
            CREATED_FEED_NEW=0
            echo "    Using existing feed: $CREATED_FEED_ID"
            pass
        else
            fail "Failed to create user creator feed"
        fi
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user creator feed"
if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    xbe_json view user-creator-feeds show "$CREATED_FEED_ID"
    assert_success
else
    skip "No user creator feed ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update user creator feed"
if [[ -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    xbe_json do user-creator-feeds update "$CREATED_FEED_ID"
    if [[ $status -ne 0 && -n "$WHOAMI_USER_ID" && "$WHOAMI_USER_ID" != "null" ]]; then
        xbe_json do user-creator-feeds update "$CREATED_FEED_ID" --user "$WHOAMI_USER_ID"
    fi
    assert_success
else
    skip "No user creator feed ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete user creator feed"
if [[ "$CREATED_FEED_NEW" -eq 1 && -n "$CREATED_FEED_ID" && "$CREATED_FEED_ID" != "null" ]]; then
    xbe_run do user-creator-feeds delete "$CREATED_FEED_ID" --confirm
    assert_success
else
    skip "Feed was not created by test run"
fi
