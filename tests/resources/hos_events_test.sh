#!/bin/bash
#
# XBE CLI Integration Tests: HOS Events
#
# Tests list and show operations for the hos-events resource.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_EVENT_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_HOS_DAY_ID=""
SAMPLE_USER_ID=""

describe "Resource: hos-events"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List HOS events"
xbe_json view hos-events list --limit 5
assert_success

test_name "List HOS events returns array"
xbe_json view hos-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS events"
fi

test_name "Capture sample HOS event (if available)"
xbe_json view hos-events list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_EVENT_ID=$(json_get ".[0].id")
        SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
        SAMPLE_HOS_DAY_ID=$(json_get ".[0].hos_day_id")
        SAMPLE_USER_ID=$(json_get ".[0].user_id")
        pass
    else
        echo "    No HOS events available; using fallback IDs for filter tests."
        pass
    fi
else
    fail "Failed to list HOS events"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

BROKER_FILTER="${SAMPLE_BROKER_ID:-${XBE_TEST_BROKER_ID:-1}}"
HOS_DAY_FILTER="${SAMPLE_HOS_DAY_ID:-1}"
USER_FILTER="${SAMPLE_USER_ID:-1}"

test_name "List HOS events with --broker filter"
xbe_json view hos-events list --broker "$BROKER_FILTER" --limit 5
assert_success

test_name "List HOS events with --hos-day filter"
xbe_json view hos-events list --hos-day "$HOS_DAY_FILTER" --limit 5
assert_success

test_name "List HOS events with --user filter"
xbe_json view hos-events list --user "$USER_FILTER" --limit 5
assert_success

test_name "List HOS events with --driver filter"
xbe_json view hos-events list --driver "$USER_FILTER" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List HOS events with --limit"
xbe_json view hos-events list --limit 3
assert_success

test_name "List HOS events with --offset"
xbe_json view hos-events list --limit 3 --offset 3
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_EVENT_ID" && "$SAMPLE_EVENT_ID" != "null" ]]; then
    test_name "Show HOS event"
    xbe_json view hos-events show "$SAMPLE_EVENT_ID"
    assert_success
else
    test_name "Show HOS event"
    skip "No HOS events available to show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
