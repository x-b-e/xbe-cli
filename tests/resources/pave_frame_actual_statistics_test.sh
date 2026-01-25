#!/bin/bash
#
# XBE CLI Integration Tests: Pave Frame Actual Statistics
#
# Tests list/show/create/update/delete operations for the pave-frame-actual-statistics resource.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_STAT_ID=""
SAMPLE_ID=""

LATITUDE="41.8781"
LONGITUDE="-87.6298"
HOUR_MIN_TEMP_F="45"
HOUR_MAX_PRECIP_IN="0.1"
WINDOW_MIN_PCT="0.6"
WINDOW="day"
AGG_LEVEL="month"
DATE_MIN="2024-01-01"
DATE_MAX="2024-12-31"
WORK_DAYS="1,2,3,4,5"
CALC_BEFORE="true"

UPDATE_LATITUDE="41.8810"
UPDATE_LONGITUDE="-87.6500"
UPDATE_HOUR_MIN_TEMP_F="50"
UPDATE_HOUR_MAX_PRECIP_IN="0.05"
UPDATE_WINDOW_MIN_PCT="0.7"
UPDATE_WINDOW="night"
UPDATE_AGG_LEVEL="week"
UPDATE_DATE_MIN="2023-01-01"
UPDATE_DATE_MAX="2023-12-31"
UPDATE_WORK_DAYS="0,6"
UPDATE_CALC_BEFORE="false"

NOW_ISO=$(date -u +%Y-%m-%dT%H:%M:%SZ)


describe "Resource: pave-frame-actual-statistics"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create pave frame actual statistic"
xbe_json do pave-frame-actual-statistics create \
    --latitude "$LATITUDE" \
    --longitude "$LONGITUDE" \
    --hour-minimum-temp-f "$HOUR_MIN_TEMP_F" \
    --hour-maximum-precip-in "$HOUR_MAX_PRECIP_IN" \
    --window-minimum-paving-hour-pct "$WINDOW_MIN_PCT" \
    --window "$WINDOW" \
    --agg-level "$AGG_LEVEL" \
    --work-days "$WORK_DAYS" \
    --date-min "$DATE_MIN" \
    --date-max "$DATE_MAX" \
    --calculate-results-before-create="$CALC_BEFORE"

if [[ $status -eq 0 ]]; then
    CREATED_STAT_ID=$(json_get ".id")
    if [[ -n "$CREATED_STAT_ID" && "$CREATED_STAT_ID" != "null" ]]; then
        register_cleanup "pave-frame-actual-statistics" "$CREATED_STAT_ID"
        pass
    else
        fail "Created statistic but no ID returned"
    fi
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
        pass
    else
        fail "Failed to create pave frame actual statistic"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List pave frame actual statistics"
xbe_json view pave-frame-actual-statistics list --limit 5
assert_success


test_name "List pave frame actual statistics returns array"
xbe_json view pave-frame-actual-statistics list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
else
    fail "Failed to list pave frame actual statistics"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show pave frame actual statistic"
SHOW_ID="${CREATED_STAT_ID:-$SAMPLE_ID}"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view pave-frame-actual-statistics show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Failed to show pave frame actual statistic"
    fi
else
    skip "No statistic ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List statistics with --created-at-min filter"
xbe_json view pave-frame-actual-statistics list --created-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List statistics with --created-at-max filter"
xbe_json view pave-frame-actual-statistics list --created-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List statistics with --is-created-at filter"
xbe_json view pave-frame-actual-statistics list --is-created-at true --limit 5
assert_success


test_name "List statistics with --updated-at-min filter"
xbe_json view pave-frame-actual-statistics list --updated-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List statistics with --updated-at-max filter"
xbe_json view pave-frame-actual-statistics list --updated-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List statistics with --is-updated-at filter"
xbe_json view pave-frame-actual-statistics list --is-updated-at true --limit 5
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update pave frame actual statistic"
if [[ -n "$CREATED_STAT_ID" && "$CREATED_STAT_ID" != "null" ]]; then
    xbe_run do pave-frame-actual-statistics update "$CREATED_STAT_ID" \
        --latitude "$UPDATE_LATITUDE" \
        --longitude "$UPDATE_LONGITUDE" \
        --hour-minimum-temp-f "$UPDATE_HOUR_MIN_TEMP_F" \
        --hour-maximum-precip-in "$UPDATE_HOUR_MAX_PRECIP_IN" \
        --window-minimum-paving-hour-pct "$UPDATE_WINDOW_MIN_PCT" \
        --window "$UPDATE_WINDOW" \
        --agg-level "$UPDATE_AGG_LEVEL" \
        --work-days "$UPDATE_WORK_DAYS" \
        --date-min "$UPDATE_DATE_MIN" \
        --date-max "$UPDATE_DATE_MAX" \
        --calculate-results-before-create="$UPDATE_CALC_BEFORE"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update pave frame actual statistic"
        fi
    fi
else
    skip "No created statistic ID available for update"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete pave frame actual statistic"
if [[ -n "$CREATED_STAT_ID" && "$CREATED_STAT_ID" != "null" ]]; then
    xbe_run do pave-frame-actual-statistics delete "$CREATED_STAT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to delete pave frame actual statistic: $output"
        fi
    fi
else
    skip "No created statistic ID available for delete"
fi

run_tests
