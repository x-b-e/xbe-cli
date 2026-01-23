#!/bin/bash
#
# XBE CLI Integration Tests: HOS Violations
#
# Tests list and show operations for the hos_violations resource.
# HOS violations capture hours-of-service rule breaches for drivers.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

VIOLATION_ID=""
BROKER_ID=""
HOS_DAY_ID=""
USER_ID=""
START_AT=""
END_AT=""
SKIP_ID_FILTERS=0

describe "Resource: hos-violations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List HOS violations"
xbe_json view hos-violations list --limit 5
assert_success

test_name "List HOS violations returns array"
xbe_json view hos-violations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS violations"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample HOS violation"
xbe_json view hos-violations list --limit 1
if [[ $status -eq 0 ]]; then
    VIOLATION_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    HOS_DAY_ID=$(json_get ".[0].hos_day_id")
    USER_ID=$(json_get ".[0].user_id")
    START_AT=$(json_get ".[0].start_at")
    END_AT=$(json_get ".[0].end_at")
    if [[ -n "$VIOLATION_ID" && "$VIOLATION_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No HOS violations available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list HOS violations"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List HOS violations with --broker filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view hos-violations list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List HOS violations with --hos-day filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$HOS_DAY_ID" && "$HOS_DAY_ID" != "null" ]]; then
    xbe_json view hos-violations list --hos-day "$HOS_DAY_ID" --limit 5
    assert_success
else
    skip "No HOS day ID available"
fi

test_name "List HOS violations with --user filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view hos-violations list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List HOS violations with --driver filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view hos-violations list --driver "$USER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List HOS violations with --start-at-min filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view hos-violations list --start-at-min "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List HOS violations with --start-at-max filter"
if [[ -n "$START_AT" && "$START_AT" != "null" ]]; then
    xbe_json view hos-violations list --start-at-max "$START_AT" --limit 5
    assert_success
else
    skip "No start-at timestamp available"
fi

test_name "List HOS violations with --end-at-min filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view hos-violations list --end-at-min "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

test_name "List HOS violations with --end-at-max filter"
if [[ -n "$END_AT" && "$END_AT" != "null" ]]; then
    xbe_json view hos-violations list --end-at-max "$END_AT" --limit 5
    assert_success
else
    skip "No end-at timestamp available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show HOS violation"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$VIOLATION_ID" && "$VIOLATION_ID" != "null" ]]; then
    xbe_json view hos-violations show "$VIOLATION_ID"
    assert_success
else
    skip "No HOS violation ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
