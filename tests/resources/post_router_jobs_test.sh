#!/bin/bash
#
# XBE CLI Integration Tests: Post Router Jobs
#
# Tests list and show operations for the post_router_jobs resource.
# Post router jobs track background worker jobs created for routed posts.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

JOB_ID=""
POST_ROUTER_ID=""
POST_ID=""
POST_WORKER_CLASS_NAME=""
SKIP_ID_FILTERS=0

describe "Resource: post-router-jobs"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List post router jobs"
xbe_json view post-router-jobs list --limit 5
assert_success

test_name "List post router jobs returns array"
xbe_json view post-router-jobs list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list post router jobs"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample post router job"
xbe_json view post-router-jobs list --limit 1
if [[ $status -eq 0 ]]; then
    JOB_ID=$(json_get ".[0].id")
    POST_ROUTER_ID=$(json_get ".[0].post_router_id")
    POST_ID=$(json_get ".[0].post_id")
    POST_WORKER_CLASS_NAME=$(json_get ".[0].post_worker_class_name")
    if [[ -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No post router jobs available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list post router jobs"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List post router jobs with --post-router filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$POST_ROUTER_ID" && "$POST_ROUTER_ID" != "null" ]]; then
    xbe_json view post-router-jobs list --post-router "$POST_ROUTER_ID" --limit 5
    assert_success
else
    skip "No post router ID available"
fi

test_name "List post router jobs with --post filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$POST_ID" && "$POST_ID" != "null" ]]; then
    xbe_json view post-router-jobs list --post "$POST_ID" --limit 5
    assert_success
else
    skip "No post ID available"
fi

test_name "List post router jobs with --post-worker-class-name filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$POST_WORKER_CLASS_NAME" && "$POST_WORKER_CLASS_NAME" != "null" ]]; then
    xbe_json view post-router-jobs list --post-worker-class-name "$POST_WORKER_CLASS_NAME" --limit 5
    assert_success
else
    skip "No post worker class name available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show post router job"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$JOB_ID" && "$JOB_ID" != "null" ]]; then
    xbe_json view post-router-jobs show "$JOB_ID"
    assert_success
else
    skip "No post router job ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
