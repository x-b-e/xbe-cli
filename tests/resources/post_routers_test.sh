#!/bin/bash
#
# XBE CLI Integration Tests: Post Routers
#
# Tests view and create operations for the post-routers resource.
# Post routers analyze posts and enqueue routing jobs.
#
# COVERAGE: Create + list filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

POST_ID="${XBE_TEST_POST_ID:-}"
POST_ROUTER_ID=""
POST_ROUTER_STATUS=""
SKIP_ID_FILTERS=0

describe "Resource: post-routers"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create post for routing"
if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    skip "Using XBE_TEST_POST_ID"
else
    xbe_json do posts create --post-type "basic" --text-content "Post router test $(date +%s)"
    if [[ $status -eq 0 ]]; then
        POST_ID=$(json_get ".id")
        if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
            register_cleanup "posts" "$POST_ID"
            pass
        else
            SKIP_CREATE=1
            fail "Created post but no ID returned"
        fi
    elif [[ "$output" == *"USER_NOT_AUTHORIZED"* || "$output" == *"FORBIDDEN"* ]]; then
        SKIP_CREATE=1
        skip "No permission to create posts"
    else
        SKIP_CREATE=1
        fail "Failed to create post for routing"
    fi
fi

if [[ -z "$POST_ID" || "$POST_ID" == "null" ]]; then
    test_name "Find existing post for routing"
    xbe_json auth whoami
    if [[ $status -eq 0 ]]; then
        USER_ID=$(json_get ".id")
        if [[ -n "$USER_ID" && "$USER_ID" != "null" ]]; then
            xbe_json view posts list --creator "User|$USER_ID" --limit 1
            if [[ $status -eq 0 ]]; then
                POST_ID=$(json_get ".[0].id")
            fi
        fi
    fi

    if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
        pass
    else
        skip "No usable post found for routing"
    fi
fi

test_name "Create post router"
if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    xbe_json do post-routers create --post "$POST_ID"

    if [[ $status -eq 0 ]]; then
        POST_ROUTER_ID=$(json_get ".id")
        POST_ROUTER_STATUS=$(json_get ".status")
        if [[ -n "$POST_ROUTER_ID" && "$POST_ROUTER_ID" != "null" ]]; then
            created_post_id=$(json_get ".post_id")
            if [[ "$created_post_id" == "$POST_ID" ]]; then
                pass
            else
                SKIP_ID_FILTERS=1
                fail "Post router post ID mismatch"
            fi
        else
            SKIP_ID_FILTERS=1
            fail "Created post router but no ID returned"
        fi
    elif [[ "$output" == *"USER_NOT_AUTHORIZED"* || "$output" == *"FORBIDDEN"* ]]; then
        SKIP_ID_FILTERS=1
        skip "No permission to create post routers"
    else
        SKIP_ID_FILTERS=1
        fail "Failed to create post router"
    fi
else
    SKIP_ID_FILTERS=1
    skip "No post available to create post router"
fi

test_name "Create post router without post fails"
xbe_json do post-routers create
assert_failure

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List post routers"
xbe_json view post-routers list --limit 5
assert_success

test_name "List post routers returns array"
xbe_json view post-routers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list post routers"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List post routers with --post filter"
if [[ -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    xbe_json view post-routers list --post "$POST_ID" --limit 5
    assert_success
else
    skip "No post ID available"
fi

test_name "List post routers with --status filter"
status_filter="$POST_ROUTER_STATUS"
if [[ -z "$status_filter" || "$status_filter" == "null" ]]; then
    status_filter="queueing"
fi
xbe_json view post-routers list --status "$status_filter" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show post router"
if [[ -z "$POST_ROUTER_ID" || "$POST_ROUTER_ID" == "null" ]]; then
    xbe_json view post-routers list --limit 1
    if [[ $status -eq 0 ]]; then
        POST_ROUTER_ID=$(json_get ".[0].id")
        POST_ROUTER_STATUS=$(json_get ".[0].status")
        if [[ -z "$POST_ID" || "$POST_ID" == "null" ]]; then
            POST_ID=$(json_get ".[0].post_id")
        fi
    fi
fi

if [[ -n "$POST_ROUTER_ID" && "$POST_ROUTER_ID" != "null" ]]; then
    xbe_json view post-routers show "$POST_ROUTER_ID"
    assert_success
else
    skip "No post router ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
