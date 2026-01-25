#!/bin/bash
#
# XBE CLI Integration Tests: Version Events
#
# Tests list and show operations for the version_events resource.
# Version events record change events exported to downstream integrations.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

EVENT_ID=""
SKIP_SHOW=0

describe "Resource: version-events"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List version events"
xbe_json view version-events list --limit 5
assert_success

test_name "List version events returns array"
xbe_json view version-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list version events"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample version event"
xbe_json view version-events list --limit 1
if [[ $status -eq 0 ]]; then
    EVENT_ID=$(json_get ".[0].id")
    if [[ -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
        pass
    else
        SKIP_SHOW=1
        skip "No version events available"
    fi
else
    SKIP_SHOW=1
    fail "Failed to list version events"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List version events with --created-at-min filter"
xbe_json view version-events list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List version events with --created-at-max filter"
xbe_json view version-events list --created-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List version events with --updated-at-min filter"
xbe_json view version-events list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List version events with --updated-at-max filter"
xbe_json view version-events list --updated-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List version events with --is-created-at filter"
xbe_json view version-events list --is-created-at true --limit 5
assert_success

test_name "List version events with --is-updated-at filter"
xbe_json view version-events list --is-updated-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show version event"
if [[ $SKIP_SHOW -eq 0 && -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json view version-events show "$EVENT_ID"
    assert_success
else
    skip "No version event ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
