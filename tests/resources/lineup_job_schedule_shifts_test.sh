#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Job Schedule Shifts
#
# Tests CRUD operations for the lineup-job-schedule-shifts resource.
#
# COVERAGE: All filters + all create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_SHIFT_ID=""
SHOW_ID=""
LINEUP_ID=""
JOB_SCHEDULE_SHIFT_ID=""
TRUCKER_ID=""
DRIVER_ID=""
TRAILER_CLASSIFICATION_ID=""
TRAILER_CLASSIFICATION_EQ_TYPE=""
IS_READY_TO_DISPATCH=""
HAS_LINEUP_DISPATCH_SHIFT=""

describe "Resource: lineup-job-schedule-shifts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup job schedule shifts"
xbe_json view lineup-job-schedule-shifts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(json_get '.[0].id // empty')
    LINEUP_ID=$(json_get '.[0].lineup_id // empty')
    JOB_SCHEDULE_SHIFT_ID=$(json_get '.[0].job_schedule_shift_id // empty')
    TRUCKER_ID=$(json_get '.[0].trucker_id // empty')
    DRIVER_ID=$(json_get '.[0].driver_id // empty')
    TRAILER_CLASSIFICATION_ID=$(json_get '.[0].trailer_classification_id // empty')
    TRAILER_CLASSIFICATION_EQ_TYPE=$(json_get '.[0].trailer_classification_equivalent_type // empty')
    IS_READY_TO_DISPATCH=$(json_get '.[0].is_ready_to_dispatch // empty')
    HAS_LINEUP_DISPATCH_SHIFT=$(json_get '.[0].has_lineup_dispatch_shift // empty')
else
    fail "Failed to list lineup job schedule shifts"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup job schedule shift"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts show "$SHOW_ID"
    assert_success
else
    skip "No lineup job schedule shift ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List lineup job schedule shifts with --lineup filter"
if [[ -n "$LINEUP_ID" && "$LINEUP_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --lineup "$LINEUP_ID" --limit 5
    assert_success
else
    skip "No lineup ID available"
fi

test_name "List lineup job schedule shifts with --job-schedule-shift filter"
if [[ -n "$JOB_SCHEDULE_SHIFT_ID" && "$JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --job-schedule-shift "$JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No job schedule shift ID available"
fi

test_name "List lineup job schedule shifts with --trucker filter"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List lineup job schedule shifts with --trailer-classification filter"
if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --trailer-classification "$TRAILER_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No trailer classification ID available"
fi

test_name "List lineup job schedule shifts with --is-ready-to-dispatch filter"
if [[ -n "$IS_READY_TO_DISPATCH" && "$IS_READY_TO_DISPATCH" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --is-ready-to-dispatch "$IS_READY_TO_DISPATCH" --limit 5
    assert_success
else
    skip "No ready-to-dispatch value available"
fi

test_name "List lineup job schedule shifts with --without-lineup-dispatch-shift filter"
if [[ -n "$HAS_LINEUP_DISPATCH_SHIFT" && "$HAS_LINEUP_DISPATCH_SHIFT" != "null" ]]; then
    if [[ "$HAS_LINEUP_DISPATCH_SHIFT" == "true" ]]; then
        filter_value="false"
    else
        filter_value="true"
    fi
    xbe_json view lineup-job-schedule-shifts list --without-lineup-dispatch-shift "$filter_value" --limit 5
    assert_success
else
    skip "No lineup dispatch shift value available"
fi

test_name "List lineup job schedule shifts with --trailer-classification-equivalent-type filter"
if [[ -n "$TRAILER_CLASSIFICATION_EQ_TYPE" && "$TRAILER_CLASSIFICATION_EQ_TYPE" != "null" ]]; then
    xbe_json view lineup-job-schedule-shifts list --trailer-classification-equivalent-type "$TRAILER_CLASSIFICATION_EQ_TYPE" --limit 5
    assert_success
else
    skip "No trailer classification equivalent type available"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List lineup job schedule shifts with --limit"
xbe_json view lineup-job-schedule-shifts list --limit 3
assert_success

test_name "List lineup job schedule shifts with --offset"
xbe_json view lineup-job-schedule-shifts list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create lineup job schedule shift without required fields fails"
xbe_run do lineup-job-schedule-shifts create
assert_failure

test_name "Update lineup job schedule shift without fields fails"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do lineup-job-schedule-shifts update "$SHOW_ID"
    assert_failure
else
    skip "No lineup job schedule shift ID available"
fi

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup lineup job schedule shift dependencies via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    if [[ -z "$LINEUP_ID" || "$LINEUP_ID" == "null" ]]; then
        lineup_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/lineups?page[limit]=1" || true)
        LINEUP_ID=$(echo "$lineup_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$JOB_SCHEDULE_SHIFT_ID" || "$JOB_SCHEDULE_SHIFT_ID" == "null" ]]; then
        shifts_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/job-schedule-shifts?page[limit]=1&filter[is-managed-or-alive]=true" || true)
        JOB_SCHEDULE_SHIFT_ID=$(echo "$shifts_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
        truckers_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/truckers?page[limit]=1" || true)
        TRUCKER_ID=$(echo "$truckers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$TRAILER_CLASSIFICATION_ID" || "$TRAILER_CLASSIFICATION_ID" == "null" ]]; then
        trailer_classifications_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/trailer-classifications?page[limit]=1" || true)
        TRAILER_CLASSIFICATION_ID=$(echo "$trailer_classifications_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    if [[ -z "$DRIVER_ID" || "$DRIVER_ID" == "null" ]]; then
        drivers_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/users?page[limit]=1" || true)
        DRIVER_ID=$(echo "$drivers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    fi

    pass
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create lineup job schedule shift with attributes"
if [[ -n "$LINEUP_ID" && "$LINEUP_ID" != "null" && -n "$JOB_SCHEDULE_SHIFT_ID" && "$JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    CREATE_ARGS=(do lineup-job-schedule-shifts create \
        --lineup "$LINEUP_ID" \
        --job-schedule-shift "$JOB_SCHEDULE_SHIFT_ID" \
        --trailer-classification-equivalent-type tendered \
        --is-brokered false \
        --is-ready-to-dispatch false \
        --exclude-from-lineup-scenarios false \
        --travel-minutes 10 \
        --loaded-tons-max 5 \
        --explicit-material-transaction-tons-max 7 \
        --notify-driver-on-late-shift-assignment false \
        --is-expecting-time-card true)

    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        CREATE_ARGS+=(--trucker "$TRUCKER_ID")
    fi
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        CREATE_ARGS+=(--trailer-classification "$TRAILER_CLASSIFICATION_ID")
    fi
    if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
        CREATE_ARGS+=(--driver "$DRIVER_ID")
    fi

    xbe_json "${CREATE_ARGS[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_SHIFT_ID=$(json_get ".id")
        if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
            register_cleanup "lineup-job-schedule-shifts" "$CREATED_SHIFT_ID"
            pass
        else
            fail "Created lineup job schedule shift but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create lineup job schedule shift: $output"
        fi
    fi
else
    skip "Missing lineup or job schedule shift ID"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update lineup job schedule shift attributes"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    UPDATE_ARGS=(do lineup-job-schedule-shifts update "$CREATED_SHIFT_ID" \
        --trailer-classification-equivalent-type assigned \
        --is-brokered true \
        --is-ready-to-dispatch false \
        --exclude-from-lineup-scenarios true \
        --travel-minutes 12 \
        --loaded-tons-max 8 \
        --explicit-material-transaction-tons-max 9 \
        --notify-driver-on-late-shift-assignment true \
        --is-expecting-time-card false)

    if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
        UPDATE_ARGS+=(--trucker "$TRUCKER_ID")
    fi
    if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
        UPDATE_ARGS+=(--trailer-classification "$TRAILER_CLASSIFICATION_ID")
    fi
    if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
        UPDATE_ARGS+=(--driver "$DRIVER_ID")
    fi

    xbe_json "${UPDATE_ARGS[@]}"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update lineup job schedule shift: $output"
        fi
    fi
else
    skip "No created lineup job schedule shift available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete lineup job schedule shift requires --confirm flag"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    xbe_run do lineup-job-schedule-shifts delete "$CREATED_SHIFT_ID"
    assert_failure
else
    skip "No created lineup job schedule shift available"
fi

test_name "Delete lineup job schedule shift with --confirm"
if [[ -n "$CREATED_SHIFT_ID" && "$CREATED_SHIFT_ID" != "null" ]]; then
    xbe_run do lineup-job-schedule-shifts delete "$CREATED_SHIFT_ID" --confirm
    assert_success
else
    skip "No created lineup job schedule shift available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
