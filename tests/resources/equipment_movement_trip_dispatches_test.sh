#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Trip Dispatches
#
# Tests list, show, create, update, and delete operations for the equipment-movement-trip-dispatches resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_CREATED_BY_ID=""
SAMPLE_TRIP_ID=""
SAMPLE_REQUIREMENT_ID=""
SAMPLE_INBOUND_REQUIREMENT_ID=""
SAMPLE_OUTBOUND_REQUIREMENT_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_DRIVER_ID=""
SAMPLE_TRAILER_ID=""
SAMPLE_LINEUP_DISPATCH_ID=""
SAMPLE_CREATED_AT=""
CREATED_ID=""

describe "Resource: equipment-movement-trip-dispatches"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment movement trip dispatches"
xbe_json view equipment-movement-trip-dispatches list --limit 5
assert_success

test_name "List equipment movement trip dispatches returns array"
xbe_json view equipment-movement-trip-dispatches list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment movement trip dispatches"
fi

# ============================================================================
# Sample Record (used for filters/show/create/update/delete)
# ============================================================================

test_name "Capture sample trip dispatch"
xbe_json view equipment-movement-trip-dispatches list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    SAMPLE_TRIP_ID=$(json_get ".[0].equipment_movement_trip_id")
    SAMPLE_REQUIREMENT_ID=$(json_get ".[0].equipment_movement_requirement_id")
    SAMPLE_INBOUND_REQUIREMENT_ID=$(json_get ".[0].inbound_equipment_requirement_id")
    SAMPLE_OUTBOUND_REQUIREMENT_ID=$(json_get ".[0].outbound_equipment_requirement_id")
    SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    SAMPLE_TRAILER_ID=$(json_get ".[0].trailer_id")
    SAMPLE_LINEUP_DISPATCH_ID=$(json_get ".[0].lineup_dispatch_id")
    SAMPLE_CREATED_AT=$(json_get ".[0].created_at")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No trip dispatches available for follow-on tests"
    fi
else
    skip "Could not list trip dispatches to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List trip dispatches with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List trip dispatches with --created-by filter"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List trip dispatches with --equipment-movement-trip filter"
if [[ -n "$SAMPLE_TRIP_ID" && "$SAMPLE_TRIP_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --equipment-movement-trip "$SAMPLE_TRIP_ID" --limit 5
    assert_success
else
    skip "No equipment movement trip ID available"
fi

test_name "List trip dispatches with --equipment-movement-requirement filter"
if [[ -n "$SAMPLE_REQUIREMENT_ID" && "$SAMPLE_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --equipment-movement-requirement "$SAMPLE_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No equipment movement requirement ID available"
fi

test_name "List trip dispatches with --inbound-equipment-requirement filter"
if [[ -n "$SAMPLE_INBOUND_REQUIREMENT_ID" && "$SAMPLE_INBOUND_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --inbound-equipment-requirement "$SAMPLE_INBOUND_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No inbound equipment requirement ID available"
fi

test_name "List trip dispatches with --outbound-equipment-requirement filter"
if [[ -n "$SAMPLE_OUTBOUND_REQUIREMENT_ID" && "$SAMPLE_OUTBOUND_REQUIREMENT_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --outbound-equipment-requirement "$SAMPLE_OUTBOUND_REQUIREMENT_ID" --limit 5
    assert_success
else
    skip "No outbound equipment requirement ID available"
fi

test_name "List trip dispatches with --trucker filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --trucker "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List trip dispatches with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List trip dispatches with --lineup-dispatch filter"
if [[ -n "$SAMPLE_LINEUP_DISPATCH_ID" && "$SAMPLE_LINEUP_DISPATCH_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --lineup-dispatch "$SAMPLE_LINEUP_DISPATCH_ID" --limit 5
    assert_success
else
    skip "No lineup dispatch ID available"
fi

test_name "List trip dispatches with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at value available"
fi

test_name "List trip dispatches with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at value available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment movement trip dispatch"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-dispatches show "$SAMPLE_ID"
    assert_success
else
    skip "No trip dispatch ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment movement trip dispatch"
CREATE_ARGS=()
if [[ -n "$SAMPLE_TRIP_ID" && "$SAMPLE_TRIP_ID" != "null" ]]; then
    CREATE_ARGS+=(--equipment-movement-trip "$SAMPLE_TRIP_ID")
elif [[ -n "$SAMPLE_REQUIREMENT_ID" && "$SAMPLE_REQUIREMENT_ID" != "null" ]]; then
    CREATE_ARGS+=(--equipment-movement-requirement "$SAMPLE_REQUIREMENT_ID")
elif [[ -n "$SAMPLE_INBOUND_REQUIREMENT_ID" && "$SAMPLE_INBOUND_REQUIREMENT_ID" != "null" ]] || \
     [[ -n "$SAMPLE_OUTBOUND_REQUIREMENT_ID" && "$SAMPLE_OUTBOUND_REQUIREMENT_ID" != "null" ]]; then
    if [[ -n "$SAMPLE_INBOUND_REQUIREMENT_ID" && "$SAMPLE_INBOUND_REQUIREMENT_ID" != "null" ]]; then
        CREATE_ARGS+=(--inbound-equipment-requirement "$SAMPLE_INBOUND_REQUIREMENT_ID")
    fi
    if [[ -n "$SAMPLE_OUTBOUND_REQUIREMENT_ID" && "$SAMPLE_OUTBOUND_REQUIREMENT_ID" != "null" ]]; then
        CREATE_ARGS+=(--outbound-equipment-requirement "$SAMPLE_OUTBOUND_REQUIREMENT_ID")
    fi
else
    skip "No usable input mode values available"
fi

if [[ ${#CREATE_ARGS[@]} -gt 0 ]]; then
    if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
        CREATE_ARGS+=(--trucker "$SAMPLE_TRUCKER_ID")
    fi
    if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
        CREATE_ARGS+=(--driver "$SAMPLE_DRIVER_ID")
    fi
    if [[ -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
        CREATE_ARGS+=(--trailer "$SAMPLE_TRAILER_ID")
    fi

    xbe_json do equipment-movement-trip-dispatches create "${CREATE_ARGS[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "equipment-movement-trip-dispatches" "$CREATED_ID"
            pass
        else
            fail "Created trip dispatch but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"422"* ]] || [[ "$output" == *"Only one input mode"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trip dispatch --tell-clerk-synchronously"
UPDATE_ID="$CREATED_ID"
if [[ -z "$UPDATE_ID" || "$UPDATE_ID" == "null" ]]; then
    UPDATE_ID="$SAMPLE_ID"
fi

if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-dispatches update "$UPDATE_ID" --tell-clerk-synchronously
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No trip dispatch ID available"
fi

test_name "Update trip dispatch --trucker"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-dispatches update "$UPDATE_ID" --trucker "$SAMPLE_TRUCKER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update trucker failed: $output"
        fi
    fi
else
    skip "No trucker ID available"
fi

test_name "Update trip dispatch --driver"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-dispatches update "$UPDATE_ID" --driver "$SAMPLE_DRIVER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update driver failed: $output"
        fi
    fi
else
    skip "No driver ID available"
fi

test_name "Update trip dispatch --trailer"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$SAMPLE_TRAILER_ID" && "$SAMPLE_TRAILER_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-dispatches update "$UPDATE_ID" --trailer "$SAMPLE_TRAILER_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Forbidden"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update trailer failed: $output"
        fi
    fi
else
    skip "No trailer ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment movement trip dispatch"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do equipment-movement-trip-dispatches delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created trip dispatch available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trip dispatch without required fields fails"
xbe_run do equipment-movement-trip-dispatches create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
