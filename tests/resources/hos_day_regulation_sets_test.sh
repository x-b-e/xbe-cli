#!/bin/bash
#
# XBE CLI Integration Tests: HOS Day Regulation Sets
#
# Tests list/show filters for the hos-day-regulation-sets resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_HOS_DAY_ID=""
SAMPLE_USER_ID=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""
describe "Resource: hos-day-regulation-sets"

# =========================================================================
# LIST Tests - Basic
# =========================================================================

test_name "List HOS day regulation sets"
xbe_json view hos-day-regulation-sets list --limit 5
assert_success

test_name "List HOS day regulation sets returns array"
xbe_json view hos-day-regulation-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS day regulation sets"
fi

# =========================================================================
# Sample Record (used for filters/show)
# =========================================================================

test_name "Capture sample HOS day regulation set"
xbe_json view hos-day-regulation-sets list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_HOS_DAY_ID=$(json_get ".[0].hos_day_id")
    SAMPLE_USER_ID=$(json_get ".[0].user_id")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No HOS day regulation sets available for follow-on tests"
    fi
else
    skip "Could not list HOS day regulation sets to capture sample"
fi

# =========================================================================
# SHOW Tests
# =========================================================================

test_name "Show HOS day regulation set"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view hos-day-regulation-sets show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_CREATED_AT=$(json_get ".created_at")
        SAMPLE_UPDATED_AT=$(json_get ".updated_at")
        pass
    else
        fail "Failed to show HOS day regulation set"
    fi
else
    skip "No regulation set ID available"
fi

# =========================================================================
# LIST Tests - Filters
# =========================================================================

test_name "List regulation sets with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List regulation sets with --hos-day filter"
if [[ -n "$SAMPLE_HOS_DAY_ID" && "$SAMPLE_HOS_DAY_ID" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --hos-day "$SAMPLE_HOS_DAY_ID" --limit 5
    assert_success
else
    skip "No HOS day ID available"
fi

test_name "List regulation sets with --user filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --user "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List regulation sets with --driver filter"
if [[ -n "$SAMPLE_USER_ID" && "$SAMPLE_USER_ID" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --driver "$SAMPLE_USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List regulation sets with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created_at available"
fi

test_name "List regulation sets with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created_at available"
fi

test_name "List regulation sets with --updated-at-min filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --updated-at-min "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated_at available"
fi

test_name "List regulation sets with --updated-at-max filter"
if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    xbe_json view hos-day-regulation-sets list --updated-at-max "$SAMPLE_UPDATED_AT" --limit 5
    assert_success
else
    skip "No updated_at available"
fi

# =========================================================================
# Summary
# =========================================================================

run_tests
