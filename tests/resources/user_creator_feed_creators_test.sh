#!/bin/bash
#
# XBE CLI Integration Tests: User Creator Feed Creators
#
# Tests view operations for the user_creator_feed_creators resource.
# User creator feed creators represent the ordered creators in a user's feed.
#
# COVERAGE: list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: user_creator_feed_creators"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user creator feed creators"
xbe_json view user-creator-feed-creators list --limit 5
assert_success

test_name "List user creator feed creators returns array"
xbe_json view user-creator-feed-creators list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user creator feed creators"
fi

ENTRY_ID=$(json_get ".[0].id")
USER_CREATOR_FEED_FILTER=$(json_get ".[0].user_creator_feed_id")
USER_FILTER=$(json_get ".[0].user_id")
USER_ID_FILTER=$(json_get ".[0].user_id")
CREATOR_TYPE_FILTER=$(json_get ".[0].creator_type")
CREATOR_ID_FILTER=$(json_get ".[0].creator_id")
FOLLOW_FILTER=$(json_get ".[0].follow_id")

if [[ -z "$USER_CREATOR_FEED_FILTER" || "$USER_CREATOR_FEED_FILTER" == "null" ]]; then
    USER_CREATOR_FEED_FILTER="1"
fi

if [[ -z "$USER_FILTER" || "$USER_FILTER" == "null" ]]; then
    USER_FILTER="1"
fi

if [[ -z "$USER_ID_FILTER" || "$USER_ID_FILTER" == "null" ]]; then
    USER_ID_FILTER="$USER_FILTER"
fi

if [[ -z "$CREATOR_TYPE_FILTER" || "$CREATOR_TYPE_FILTER" == "null" ]]; then
    CREATOR_TYPE_FILTER="User"
fi

if [[ -z "$CREATOR_ID_FILTER" || "$CREATOR_ID_FILTER" == "null" ]]; then
    CREATOR_ID_FILTER="1"
fi

if [[ -z "$FOLLOW_FILTER" || "$FOLLOW_FILTER" == "null" ]]; then
    FOLLOW_FILTER="1"
fi

CREATOR_FILTER="${CREATOR_TYPE_FILTER}|${CREATOR_ID_FILTER}"

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user creator feed creators with --user-creator-feed filter"
xbe_json view user-creator-feed-creators list --user-creator-feed "$USER_CREATOR_FEED_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --user filter"
xbe_json view user-creator-feed-creators list --user "$USER_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --user-id filter"
xbe_json view user-creator-feed-creators list --user-id "$USER_ID_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --creator filter"
xbe_json view user-creator-feed-creators list --creator "$CREATOR_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --creator-type filter"
xbe_json view user-creator-feed-creators list --creator-type "$CREATOR_TYPE_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --creator-type and --creator-id filters"
xbe_json view user-creator-feed-creators list --creator-type "$CREATOR_TYPE_FILTER" --creator-id "$CREATOR_ID_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --not-creator-type filter"
xbe_json view user-creator-feed-creators list --not-creator-type "$CREATOR_TYPE_FILTER" --limit 5
assert_success

test_name "List user creator feed creators with --follow filter"
xbe_json view user-creator-feed-creators list --follow "$FOLLOW_FILTER" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show user creator feed creator"
if [[ -n "$ENTRY_ID" && "$ENTRY_ID" != "null" ]]; then
    xbe_json view user-creator-feed-creators show "$ENTRY_ID"
    assert_success
else
    skip "No user creator feed creator available for show test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
