#!/bin/bash
#
# XBE CLI Integration Tests: HOS Availability Snapshots
#
# Tests list and show operations for the hos_availability_snapshots resource.
# HOS availability snapshots capture driver hours-of-service availability.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SNAPSHOT_ID=""
BROKER_ID=""
HOS_DAY_ID=""
USER_ID=""
SKIP_ID_FILTERS=0

describe "Resource: hos-availability-snapshots"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List HOS availability snapshots"
xbe_json view hos-availability-snapshots list --limit 5
assert_success

test_name "List HOS availability snapshots returns array"
xbe_json view hos-availability-snapshots list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list HOS availability snapshots"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample HOS availability snapshot"
xbe_json view hos-availability-snapshots list --limit 1
if [[ $status -eq 0 ]]; then
    SNAPSHOT_ID=$(json_get ".[0].id")
    BROKER_ID=$(json_get ".[0].broker_id")
    HOS_DAY_ID=$(json_get ".[0].hos_day_id")
    USER_ID=$(json_get ".[0].user_id")
    if [[ -n "$SNAPSHOT_ID" && "$SNAPSHOT_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No HOS availability snapshots available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list HOS availability snapshots"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List HOS availability snapshots with --broker filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view hos-availability-snapshots list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List HOS availability snapshots with --hos-day filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$HOS_DAY_ID" && "$HOS_DAY_ID" != "null" ]]; then
    xbe_json view hos-availability-snapshots list --hos-day "$HOS_DAY_ID" --limit 5
    assert_success
else
    skip "No HOS day ID available"
fi

test_name "List HOS availability snapshots with --user filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view hos-availability-snapshots list --user "$USER_ID" --limit 5
    assert_success
else
    skip "No user ID available"
fi

test_name "List HOS availability snapshots with --driver filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$USER_ID" && "$USER_ID" != "null" ]]; then
    xbe_json view hos-availability-snapshots list --driver "$USER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show HOS availability snapshot"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SNAPSHOT_ID" && "$SNAPSHOT_ID" != "null" ]]; then
    xbe_json view hos-availability-snapshots show "$SNAPSHOT_ID"
    assert_success
else
    skip "No HOS availability snapshot ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
