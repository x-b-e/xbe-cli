#!/bin/bash
#
# XBE CLI Integration Tests: Transport Routes
#
# Tests view operations for the transport-routes resource.
# Transport routes are computed routes between origin and destination coordinates.
#
# COVERAGE: List + show + filters + pagination (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

ROUTE_ID=""
ORIGIN_LAT=""
ORIGIN_LNG=""
DEST_LAT=""
DEST_LNG=""

describe "Resource: transport-routes (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport routes"
xbe_json view transport-routes list --limit 5
assert_success

test_name "List transport routes returns array"
xbe_json view transport-routes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list transport routes"
fi

# ==========================================================================
# Sample Route for Show/Filter Tests
# ==========================================================================

test_name "Get a transport route for detail/filter tests"
xbe_json view transport-routes list --limit 1
if [[ $status -eq 0 ]]; then
    ROUTE_ID=$(json_get ".[0].id")
    ORIGIN_LAT=$(json_get ".[0].origin_latitude")
    ORIGIN_LNG=$(json_get ".[0].origin_longitude")
    DEST_LAT=$(json_get ".[0].destination_latitude")
    DEST_LNG=$(json_get ".[0].destination_longitude")

    if [[ -n "$ROUTE_ID" && "$ROUTE_ID" != "null" ]]; then
        pass
    else
        skip "No transport routes found in the system"
        run_tests
    fi
else
    fail "Failed to list transport routes"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show transport route"
xbe_json view transport-routes show "$ROUTE_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

if [[ -n "$ORIGIN_LAT" && "$ORIGIN_LAT" != "null" && -n "$ORIGIN_LNG" && "$ORIGIN_LNG" != "null" ]]; then
    test_name "List transport routes with --near-origin-location filter"
    NEAR_ORIGIN_FILTER="$ORIGIN_LAT|$ORIGIN_LNG|10"
    xbe_json view transport-routes list --near-origin-location "$NEAR_ORIGIN_FILTER" --limit 5
    assert_success
else
    skip "Origin coordinates not available for near-origin-location test"
fi

if [[ -n "$DEST_LAT" && "$DEST_LAT" != "null" && -n "$DEST_LNG" && "$DEST_LNG" != "null" ]]; then
    test_name "List transport routes with --near-destination-location filter"
    NEAR_DEST_FILTER="$DEST_LAT|$DEST_LNG|10"
    xbe_json view transport-routes list --near-destination-location "$NEAR_DEST_FILTER" --limit 5
    assert_success
else
    skip "Destination coordinates not available for near-destination-location test"
fi

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List transport routes with --limit"
xbe_json view transport-routes list --limit 3
assert_success

test_name "List transport routes with --offset"
xbe_json view transport-routes list --limit 3 --offset 3
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
