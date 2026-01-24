#!/bin/bash
#
# XBE CLI Integration Tests: Vehicle Location Events
#
# Tests view operations for vehicle location events.
# Vehicle location events represent tractor and trailer positions.
#
# COVERAGE: List + filters + show (when available)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FIRST_EVENT_ID=""
FIRST_TRACTOR_ID=""
FIRST_TRAILER_ID=""

FALLBACK_TRACTOR_ID="1"
FALLBACK_TRAILER_ID="1"


describe "Resource: vehicle-location-events (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List vehicle location events"
xbe_json view vehicle-location-events list --limit 5
assert_success

test_name "List vehicle location events returns array"
xbe_json view vehicle-location-events list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list vehicle location events"
fi

# Capture IDs for downstream tests
xbe_json view vehicle-location-events list --limit 5
if [[ $status -eq 0 ]]; then
    FIRST_EVENT_ID=$(json_get ".[0].id")
    FIRST_TRACTOR_ID=$(json_get ".[0].tractor_id")
    FIRST_TRAILER_ID=$(json_get ".[0].trailer_id")
else
    FIRST_EVENT_ID=""
    FIRST_TRACTOR_ID=""
    FIRST_TRAILER_ID=""
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show vehicle location event"
if [[ -n "$FIRST_EVENT_ID" && "$FIRST_EVENT_ID" != "null" ]]; then
    xbe_json view vehicle-location-events show "$FIRST_EVENT_ID"
    assert_success
else
    skip "No vehicle location event ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List vehicle location events with --tractor filter"
TRACTOR_FILTER_ID="$FALLBACK_TRACTOR_ID"
if [[ -n "$FIRST_TRACTOR_ID" && "$FIRST_TRACTOR_ID" != "null" ]]; then
    TRACTOR_FILTER_ID="$FIRST_TRACTOR_ID"
fi
xbe_json view vehicle-location-events list --tractor "$TRACTOR_FILTER_ID" --limit 5
assert_success

test_name "List vehicle location events with --trailer filter"
TRAILER_FILTER_ID="$FALLBACK_TRAILER_ID"
if [[ -n "$FIRST_TRAILER_ID" && "$FIRST_TRAILER_ID" != "null" ]]; then
    TRAILER_FILTER_ID="$FIRST_TRAILER_ID"
fi
xbe_json view vehicle-location-events list --trailer "$TRAILER_FILTER_ID" --limit 5
assert_success

test_name "List vehicle location events with --event-at-min filter"
xbe_json view vehicle-location-events list --event-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --event-at-max filter"
xbe_json view vehicle-location-events list --event-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --is-event-at filter"
xbe_json view vehicle-location-events list --is-event-at true --limit 5
assert_success

test_name "List vehicle location events with --include-device-location-events filter"
INCLUDE_DEVICE_TRACTOR_ID="$TRACTOR_FILTER_ID"
if [[ -z "$INCLUDE_DEVICE_TRACTOR_ID" || "$INCLUDE_DEVICE_TRACTOR_ID" == "null" ]]; then
    INCLUDE_DEVICE_TRACTOR_ID="$FALLBACK_TRACTOR_ID"
fi
xbe_json view vehicle-location-events list --tractor "$INCLUDE_DEVICE_TRACTOR_ID" --include-device-location-events true --limit 5
assert_success

test_name "List vehicle location events with --time-slice-min filter"
xbe_json view vehicle-location-events list --time-slice-min 5 --limit 5
assert_success

test_name "List vehicle location events with --total-count-max filter"
xbe_json view vehicle-location-events list --total-count-max 50 --limit 5
assert_success

test_name "List vehicle location events with --created-at-min filter"
xbe_json view vehicle-location-events list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --created-at-max filter"
xbe_json view vehicle-location-events list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --is-created-at filter"
xbe_json view vehicle-location-events list --is-created-at true --limit 5
assert_success

test_name "List vehicle location events with --updated-at-min filter"
xbe_json view vehicle-location-events list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --updated-at-max filter"
xbe_json view vehicle-location-events list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List vehicle location events with --is-updated-at filter"
xbe_json view vehicle-location-events list --is-updated-at true --limit 5
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "List vehicle location events with include-device-location-events missing tractor/trailer fails"
xbe_json view vehicle-location-events list --include-device-location-events true --limit 5
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
