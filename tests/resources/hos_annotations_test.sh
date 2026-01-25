#!/bin/bash
#
# XBE CLI Integration Tests: HOS Annotations
#
# Tests list and show operations for the hos-annotations resource.
#
# COVERAGE: List filters + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_HOS_DAY_ID=""
SAMPLE_HOS_EVENT_ID=""

describe "Resource: hos-annotations"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List HOS annotations"
xbe_json view hos-annotations list --limit 5
assert_success

test_name "List HOS annotations returns array"
xbe_json view hos-annotations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS annotations"
fi

# ==========================================================================
# Sample Record (used for filters/show)
# ==========================================================================

test_name "Capture sample annotation"
xbe_json view hos-annotations list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_BROKER_ID=$(json_get ".[0].broker_id")
    SAMPLE_HOS_DAY_ID=$(json_get ".[0].hos_day_id")
    SAMPLE_HOS_EVENT_ID=$(json_get ".[0].hos_event_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No annotations available for follow-on tests"
    fi
else
    skip "Could not list annotations to capture sample"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List annotations with --broker filter"
if [[ -n "$SAMPLE_BROKER_ID" && "$SAMPLE_BROKER_ID" != "null" ]]; then
    xbe_json view hos-annotations list --broker "$SAMPLE_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List annotations with --hos-day filter"
if [[ -n "$SAMPLE_HOS_DAY_ID" && "$SAMPLE_HOS_DAY_ID" != "null" ]]; then
    xbe_json view hos-annotations list --hos-day "$SAMPLE_HOS_DAY_ID" --limit 5
    assert_success
else
    skip "No HOS day ID available"
fi

test_name "List annotations with --hos-event filter"
if [[ -n "$SAMPLE_HOS_EVENT_ID" && "$SAMPLE_HOS_EVENT_ID" != "null" ]]; then
    xbe_json view hos-annotations list --hos-event "$SAMPLE_HOS_EVENT_ID" --limit 5
    assert_success
else
    skip "No HOS event ID available"
fi

test_name "List annotations with --created-at-min filter"
xbe_json view hos-annotations list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List annotations with --created-at-max filter"
xbe_json view hos-annotations list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List annotations with --updated-at-min filter"
xbe_json view hos-annotations list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List annotations with --updated-at-max filter"
xbe_json view hos-annotations list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show HOS annotation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view hos-annotations show "$SAMPLE_ID"
    assert_success
else
    skip "No annotation ID available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Delete HOS annotation requires --confirm flag"
xbe_run do hos-annotations delete "nonexistent"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
