#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Status Changes
#
# Tests list/show operations for the time-card-status-changes resource.
#
# COVERAGE: All filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_STATUS_CHANGE_ID=""
SAMPLE_TIME_CARD_ID=""
SAMPLE_STATUS=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""

describe "Resource: time-card-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card status changes"
xbe_json view time-card-status-changes list --limit 5
assert_success

test_name "List time card status changes returns array"
xbe_json view time-card-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time card status changes"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate time card status change for filters"
xbe_json view time-card-status-changes list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_STATUS_CHANGE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        SAMPLE_STATUS=$(json_get ".[0].status")
        SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
        SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
        if [[ -z "$SAMPLE_TIME_CARD_ID" || "$SAMPLE_TIME_CARD_ID" == "null" ]]; then
            xbe_json view time-card-status-changes show "$SAMPLE_STATUS_CHANGE_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_TIME_CARD_ID=$(json_get ".time_card_id")
                SAMPLE_STATUS=$(json_get ".status")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
            fi
        fi
        pass
    else
        if [[ -n "$XBE_TEST_TIME_CARD_STATUS_CHANGE_ID" ]]; then
            xbe_json view time-card-status-changes show "$XBE_TEST_TIME_CARD_STATUS_CHANGE_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_STATUS_CHANGE_ID=$(json_get ".id")
                SAMPLE_TIME_CARD_ID=$(json_get ".time_card_id")
                SAMPLE_STATUS=$(json_get ".status")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
                pass
            else
                skip "Failed to load XBE_TEST_TIME_CARD_STATUS_CHANGE_ID"
            fi
        else
            skip "No status changes found. Set XBE_TEST_TIME_CARD_STATUS_CHANGE_ID for filter tests."
        fi
    fi
else
    fail "Failed to list time card status changes for filters"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_STATUS_CHANGE_ID" && "$SAMPLE_STATUS_CHANGE_ID" != "null" ]]; then
    test_name "Show time card status change"
    xbe_json view time-card-status-changes show "$SAMPLE_STATUS_CHANGE_ID"
    assert_success
else
    test_name "Show time card status change"
    skip "No sample status change available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$XBE_TEST_TIME_CARD_ID" ]]; then
    SAMPLE_TIME_CARD_ID="$XBE_TEST_TIME_CARD_ID"
fi

if [[ -z "$SAMPLE_TIME_CARD_ID" || "$SAMPLE_TIME_CARD_ID" == "null" ]]; then
    xbe_json view time-cards list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    test_name "Filter by time card"
    xbe_json view time-card-status-changes list --time-card "$SAMPLE_TIME_CARD_ID"
    assert_success
else
    test_name "Filter by time card"
    skip "No time card ID available"
fi

if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    test_name "Filter by status"
    xbe_json view time-card-status-changes list --status "$SAMPLE_STATUS"
    assert_success
else
    test_name "Filter by status"
    skip "No status available"
fi

if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    test_name "Filter by created-at min/max"
    xbe_json view time-card-status-changes list \
        --created-at-min "$SAMPLE_CREATED_AT" \
        --created-at-max "$SAMPLE_CREATED_AT"
    assert_success

    test_name "Filter by is-created-at"
    xbe_json view time-card-status-changes list --is-created-at true --limit 5
    assert_success
else
    test_name "Filter by created-at min/max"
    skip "No created-at available"
    test_name "Filter by is-created-at"
    skip "No created-at available"
fi

if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    test_name "Filter by updated-at min/max"
    xbe_json view time-card-status-changes list \
        --updated-at-min "$SAMPLE_UPDATED_AT" \
        --updated-at-max "$SAMPLE_UPDATED_AT"
    assert_success

    test_name "Filter by is-updated-at"
    xbe_json view time-card-status-changes list --is-updated-at true --limit 5
    assert_success
else
    test_name "Filter by updated-at min/max"
    skip "No updated-at available"
    test_name "Filter by is-updated-at"
    skip "No updated-at available"
fi

test_name "List time card status changes with --offset"
xbe_json view time-card-status-changes list --limit 3 --offset 1
assert_success

test_name "List time card status changes with --sort"
xbe_json view time-card-status-changes list --sort created-at --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
