#!/bin/bash
#
# XBE CLI Integration Tests: Digital Fleet Ticket Events
#
# Tests list and show operations for the digital-fleet-ticket-events resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

EVENT_ID=""
BROKER_ID=""
SKIP_SHOW=0

describe "Resource: digital-fleet-ticket-events"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List digital fleet ticket events"
xbe_json view digital-fleet-ticket-events list --limit 5
assert_success

test_name "List digital fleet ticket events returns array"
xbe_json view digital-fleet-ticket-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list digital fleet ticket events"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample digital fleet ticket event"
xbe_json view digital-fleet-ticket-events list --limit 1
if [[ $status -eq 0 ]]; then
    EVENT_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    if [[ -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
        pass
    else
        SKIP_SHOW=1
        skip "No digital fleet ticket events available"
    fi
else
    SKIP_SHOW=1
    fail "Failed to list digital fleet ticket events"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List digital fleet ticket events with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view digital-fleet-ticket-events list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List digital fleet ticket events with --event-at-min filter"
xbe_json view digital-fleet-ticket-events list --event-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List digital fleet ticket events with --event-at-max filter"
xbe_json view digital-fleet-ticket-events list --event-at-max "2024-12-31T23:59:59Z" --limit 5
assert_success

test_name "List digital fleet ticket events with --is-event-at filter"
xbe_json view digital-fleet-ticket-events list --is-event-at true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show digital fleet ticket event"
if [[ $SKIP_SHOW -eq 0 && -n "$EVENT_ID" && "$EVENT_ID" != "null" ]]; then
    xbe_json view digital-fleet-ticket-events show "$EVENT_ID"
    assert_success
else
    skip "No digital fleet ticket event ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
