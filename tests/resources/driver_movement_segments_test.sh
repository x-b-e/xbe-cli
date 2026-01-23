#!/bin/bash
#
# XBE CLI Integration Tests: Driver Movement Segments
#
# Tests list and show operations for the driver_movement_segments resource.
# Driver movement segments represent contiguous moving or stationary intervals.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SEGMENT_ID=""
SEGMENT_SET_ID=""
TRAILER_ID=""
TRACTOR_ID=""
SITE_KIND=""
START_AT=""
END_AT=""
IS_MOVING=""
SKIP_ID_FILTERS=0

describe "Resource: driver-movement-segments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver movement segments"
xbe_json view driver-movement-segments list --limit 5
assert_success

test_name "List driver movement segments returns array"
xbe_json view driver-movement-segments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver movement segments"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample driver movement segment"
xbe_json view driver-movement-segments list --limit 1
if [[ $status -eq 0 ]]; then
    SEGMENT_ID=$(json_get ".[0].id")
    SEGMENT_SET_ID=$(json_get ".[0].driver_movement_segment_set_id")
    TRAILER_ID=$(json_get ".[0].trailer_id")
    TRACTOR_ID=$(json_get ".[0].tractor_id")
    SITE_KIND=$(json_get ".[0].site_kind")
    START_AT=$(json_get ".[0].start_at")
    END_AT=$(json_get ".[0].end_at")
    IS_MOVING=$(json_get ".[0].is_moving")
    if [[ -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No driver movement segments available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list driver movement segments"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List driver movement segments with --is-moving filter"
if [[ -n "$IS_MOVING" && "$IS_MOVING" != "null" ]]; then
    xbe_json view driver-movement-segments list --is-moving "$IS_MOVING" --limit 5
else
    xbe_json view driver-movement-segments list --is-moving true --limit 5
fi
assert_success

test_name "List driver movement segments with --is-stationary filter"
xbe_json view driver-movement-segments list --is-stationary true --limit 5
assert_success

test_name "List driver movement segments with --at-job-site filter"
xbe_json view driver-movement-segments list --at-job-site true --limit 5
assert_success

test_name "List driver movement segments with --at-material-site filter"
xbe_json view driver-movement-segments list --at-material-site true --limit 5
assert_success

test_name "List driver movement segments with --at-parking-site filter"
xbe_json view driver-movement-segments list --at-parking-site true --limit 5
assert_success

test_name "List driver movement segments with --site-kind filter"
if [[ -n "$SITE_KIND" && "$SITE_KIND" != "null" ]]; then
    xbe_json view driver-movement-segments list --site-kind "$SITE_KIND" --limit 5
else
    xbe_json view driver-movement-segments list --site-kind job_site --limit 5
fi
assert_success

test_name "List driver movement segments with --driver-movement-segment-set filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SEGMENT_SET_ID" && "$SEGMENT_SET_ID" != "null" ]]; then
    xbe_json view driver-movement-segments list --driver-movement-segment-set "$SEGMENT_SET_ID" --limit 5
    assert_success
else
    skip "No segment set ID available"
fi

test_name "List driver movement segments with --trailer filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
    xbe_json view driver-movement-segments list --trailer "$TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "List driver movement segments with --tractor filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view driver-movement-segments list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "List driver movement segments with --start-at-min filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view driver-movement-segments list --start-at-min "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List driver movement segments with --start-at-max filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view driver-movement-segments list --start-at-max "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List driver movement segments with --end-at-min filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view driver-movement-segments list --end-at-min "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

test_name "List driver movement segments with --end-at-max filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view driver-movement-segments list --end-at-max "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver movement segment"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SEGMENT_ID" && "$SEGMENT_ID" != "null" ]]; then
    xbe_json view driver-movement-segments show "$SEGMENT_ID"
    assert_success
else
    skip "No driver movement segment ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
