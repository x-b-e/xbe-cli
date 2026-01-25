#!/bin/bash
#
# XBE CLI Integration Tests: Pave Frame Actual Hours
#
# Tests CRUD operations for the pave_frame_actual_hours resource.
#
# COVERAGE: All create/update attributes + list pagination
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ID=""

DATE=$(date -u +%Y-%m-%d)
HOUR=$((RANDOM % 24))
if [[ $HOUR -ge 6 && $HOUR -lt 18 ]]; then
    WINDOW="day"
else
    WINDOW="night"
fi
LAT=$(( (RANDOM % 180) - 90 ))
LON=$(( (RANDOM % 360) - 180 ))
TEMP_MIN_F="45.5"
PRECIP_1HR_IN="0.2"

describe "Resource: pave_frame_actual_hours"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create pave frame actual hour with required fields"
xbe_json do pave-frame-actual-hours create \
    --date "$DATE" \
    --hour "$HOUR" \
    --window "$WINDOW" \
    --latitude "$LAT" \
    --longitude "$LON" \
    --temp-min-f "$TEMP_MIN_F" \
    --precip-1hr-in "$PRECIP_1HR_IN"

if [[ $status -eq 0 ]]; then
    CREATED_ID=$(json_get ".id")
    if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
        register_cleanup "pave-frame-actual-hours" "$CREATED_ID"
        pass
    else
        fail "Created pave frame actual hour but no ID returned"
    fi
else
    fail "Failed to create pave frame actual hour"
fi

if [[ -z "$CREATED_ID" || "$CREATED_ID" == "null" ]]; then
    echo "Cannot continue without a valid pave frame actual hour ID"
    run_tests
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show pave frame actual hour"
xbe_json view pave-frame-actual-hours show "$CREATED_ID"
assert_success

# ============================================================================
# UPDATE Tests (cover all attributes)
# ============================================================================

UPDATED_HOUR=$(( (HOUR + 1) % 24 ))
if [[ $UPDATED_HOUR -ge 6 && $UPDATED_HOUR -lt 18 ]]; then
    UPDATED_WINDOW="day"
else
    UPDATED_WINDOW="night"
fi
UPDATED_LAT=$((LAT + 1))
if [[ $UPDATED_LAT -gt 90 ]]; then
    UPDATED_LAT=$((LAT - 1))
fi
UPDATED_LON=$((LON + 1))
if [[ $UPDATED_LON -gt 180 ]]; then
    UPDATED_LON=$((LON - 1))
fi
UPDATED_TEMP_MIN_F="50.0"
UPDATED_PRECIP_1HR_IN="0.4"


test_name "Update pave frame actual hour fields"
xbe_json do pave-frame-actual-hours update "$CREATED_ID" \
    --date "$DATE" \
    --hour "$UPDATED_HOUR" \
    --window "$UPDATED_WINDOW" \
    --latitude "$UPDATED_LAT" \
    --longitude "$UPDATED_LON" \
    --temp-min-f "$UPDATED_TEMP_MIN_F" \
    --precip-1hr-in "$UPDATED_PRECIP_1HR_IN"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List pave frame actual hours"
xbe_json view pave-frame-actual-hours list --limit 5
assert_success

test_name "List pave frame actual hours returns array"
xbe_json view pave-frame-actual-hours list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list pave frame actual hours"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List pave frame actual hours with --limit"
xbe_json view pave-frame-actual-hours list --limit 3
assert_success

test_name "List pave frame actual hours with --offset"
xbe_json view pave-frame-actual-hours list --limit 3 --offset 3
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete pave frame actual hour requires --confirm flag"
xbe_run do pave-frame-actual-hours delete "$CREATED_ID"
assert_failure

# Create a record specifically for deletion
DEL_HOUR=$(( (UPDATED_HOUR + 2) % 24 ))
if [[ $DEL_HOUR -ge 6 && $DEL_HOUR -lt 18 ]]; then
    DEL_WINDOW="day"
else
    DEL_WINDOW="night"
fi
DEL_LAT=$((UPDATED_LAT + 1))
if [[ $DEL_LAT -gt 90 ]]; then
    DEL_LAT=$((UPDATED_LAT - 1))
fi
DEL_LON=$((UPDATED_LON + 1))
if [[ $DEL_LON -gt 180 ]]; then
    DEL_LON=$((UPDATED_LON - 1))
fi

test_name "Delete pave frame actual hour with --confirm"
xbe_json do pave-frame-actual-hours create \
    --date "$DATE" \
    --hour "$DEL_HOUR" \
    --window "$DEL_WINDOW" \
    --latitude "$DEL_LAT" \
    --longitude "$DEL_LON" \
    --temp-min-f "$UPDATED_TEMP_MIN_F" \
    --precip-1hr-in "$UPDATED_PRECIP_1HR_IN"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_run do pave-frame-actual-hours delete "$DEL_ID" --confirm
    assert_success
else
    skip "Could not create pave frame actual hour for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create pave frame actual hour without date fails"
xbe_json do pave-frame-actual-hours create \
    --hour "$HOUR" \
    --window "$WINDOW" \
    --latitude "$LAT" \
    --longitude "$LON" \
    --temp-min-f "$TEMP_MIN_F" \
    --precip-1hr-in "$PRECIP_1HR_IN"
assert_failure

test_name "Update without any fields fails"
xbe_json do pave-frame-actual-hours update "$CREATED_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
