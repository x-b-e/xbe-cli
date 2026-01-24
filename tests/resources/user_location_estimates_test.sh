#!/bin/bash
#
# XBE CLI Integration Tests: User Location Estimates
#
# Tests list operations and filters for user location estimates.
#
# COVERAGE: Required filter + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CURRENT_USER_ID=""
USER_FILTER_ID=""

AS_OF="2025-01-01T12:00:00Z"
EARLIEST_EVENT_AT="2025-01-01T00:00:00Z"
LATEST_EVENT_AT="2025-01-02T00:00:00Z"

describe "Resource: user-location-estimates (view-only)"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Fetch current user (optional)"
xbe_json auth whoami
if [[ $status -eq 0 ]]; then
    CURRENT_USER_ID=$(json_get ".id")
    if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
        pass
    else
        skip "No user ID returned"
    fi
else
    skip "Failed to fetch current user"
fi

if [[ -n "$CURRENT_USER_ID" && "$CURRENT_USER_ID" != "null" ]]; then
    USER_FILTER_ID="$CURRENT_USER_ID"
else
    USER_FILTER_ID="1"
fi

# ============================================================================
# LIST Tests - Required Filter
# ============================================================================

test_name "List user location estimates requires --user"
xbe_json view user-location-estimates list
assert_failure

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List user location estimates"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID"
assert_success

test_name "List user location estimates returns array"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list user location estimates"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List user location estimates with --as-of filter"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID" --as-of "$AS_OF"
assert_success

test_name "List user location estimates with --earliest-event-at filter"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID" --earliest-event-at "$EARLIEST_EVENT_AT"
assert_success

test_name "List user location estimates with --latest-event-at filter"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID" --latest-event-at "$LATEST_EVENT_AT"
assert_success

test_name "List user location estimates with --max-abs-latency-seconds filter"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID" --max-abs-latency-seconds 3600
assert_success

test_name "List user location estimates with --max-latest-seconds filter"
xbe_json view user-location-estimates list --user "$USER_FILTER_ID" --max-latest-seconds 86400
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
