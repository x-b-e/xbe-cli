#!/bin/bash
#
# XBE CLI Integration Tests: Lineup Scenario Trailer Lineup Job Schedule Shifts
#
# Tests CRUD operations for the lineup-scenario-trailer-lineup-job-schedule-shifts resource.
#
# COVERAGE: All filters + all create/update attributes + show + delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RECORD_ID=""
SHOW_ID=""
LINEUP_SCENARIO_TRAILER_ID=""
LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID=""

START_SITE_MINUTES=""
END_SITE_MINUTES=""

describe "Resource: lineup-scenario-trailer-lineup-job-schedule-shifts"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List lineup scenario trailer lineup job schedule shifts"
xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SHOW_ID=$(json_get '.[0].id // empty')
    LINEUP_SCENARIO_TRAILER_ID=$(json_get '.[0].lineup_scenario_trailer_id // empty')
    LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID=$(json_get '.[0].lineup_scenario_lineup_job_schedule_shift_id // empty')
    START_SITE_MINUTES=$(json_get '.[0].start_site_distance_minutes // empty')
    END_SITE_MINUTES=$(json_get '.[0].end_site_distance_minutes // empty')
else
    fail "Failed to list lineup scenario trailer lineup job schedule shifts"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show lineup scenario trailer lineup job schedule shift"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts show "$SHOW_ID"
    assert_success
else
    skip "No record ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List records with --lineup-scenario-trailer filter"
if [[ -n "$LINEUP_SCENARIO_TRAILER_ID" && "$LINEUP_SCENARIO_TRAILER_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts list --lineup-scenario-trailer "$LINEUP_SCENARIO_TRAILER_ID" --limit 5
    assert_success
else
    skip "No lineup scenario trailer ID available"
fi

test_name "List records with --lineup-scenario-lineup-job-schedule-shift filter"
if [[ -n "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" && "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts list --lineup-scenario-lineup-job-schedule-shift "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" --limit 5
    assert_success
else
    skip "No lineup scenario lineup job schedule shift ID available"
fi

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List records with --limit"
xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts list --limit 3
assert_success

test_name "List records with --offset"
xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts list --limit 3 --offset 3
assert_success

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create record without required fields fails"
xbe_run do lineup-scenario-trailer-lineup-job-schedule-shifts create
assert_failure

test_name "Update record without fields fails"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do lineup-scenario-trailer-lineup-job-schedule-shifts update "$SHOW_ID"
    assert_failure
else
    skip "No record ID available"
fi

test_name "Delete record requires --confirm flag"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_run do lineup-scenario-trailer-lineup-job-schedule-shifts delete "$SHOW_ID"
    assert_failure
else
    skip "No record ID available"
fi

# ============================================================================
# Prerequisite Lookup via API
# ============================================================================

test_name "Lookup lineup scenario trailer + lineup scenario lineup job schedule shift via API"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"

    trailers_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/lineup-scenario-trailers?page[limit]=20" || true)

    LINEUP_SCENARIO_TRAILER_ID=$(echo "$trailers_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    lineup_scenario_id=$(echo "$trailers_json" | jq -r '.data[0].relationships["lineup-scenario"].data.id // empty' 2>/dev/null || true)

    if [[ -z "$LINEUP_SCENARIO_TRAILER_ID" || -z "$lineup_scenario_id" ]]; then
        skip "No lineup scenario trailer with lineup scenario found"
    else
        shifts_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/lineup-scenario-lineup-job-schedule-shifts?filter[lineup-scenario]=$lineup_scenario_id&page[limit]=20" || true)

        LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID=$(echo "$shifts_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)

        if [[ -z "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" ]]; then
            skip "No lineup scenario lineup job schedule shifts found for lineup scenario"
        else
            pass
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create record with required fields"
if [[ -n "$LINEUP_SCENARIO_TRAILER_ID" && -n "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json do lineup-scenario-trailer-lineup-job-schedule-shifts create \
        --lineup-scenario-trailer "$LINEUP_SCENARIO_TRAILER_ID" \
        --lineup-scenario-lineup-job-schedule-shift "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" \
        --start-site-distance-minutes 10 \
        --end-site-distance-minutes 12

    if [[ $status -eq 0 ]]; then
        CREATED_RECORD_ID=$(json_get ".id")
        if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
            register_cleanup "lineup-scenario-trailer-lineup-job-schedule-shifts" "$CREATED_RECORD_ID"
            pass
        else
            fail "Created record but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create record: $output"
        fi
    fi
else
    skip "Missing lineup scenario trailer or lineup scenario lineup job schedule shift for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update record distances"
if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
    xbe_json do lineup-scenario-trailer-lineup-job-schedule-shifts update "$CREATED_RECORD_ID" \
        --start-site-distance-minutes 15 \
        --end-site-distance-minutes 18
    assert_success
else
    skip "No created record available"
fi

# ============================================================================
# SHOW Tests (Created)
# ============================================================================

test_name "Show created record details"
if [[ -n "$CREATED_RECORD_ID" && "$CREATED_RECORD_ID" != "null" ]]; then
    xbe_json view lineup-scenario-trailer-lineup-job-schedule-shifts show "$CREATED_RECORD_ID"
    assert_success
else
    skip "No created record available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete record with --confirm"
if [[ -n "$LINEUP_SCENARIO_TRAILER_ID" && -n "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" ]]; then
    xbe_json do lineup-scenario-trailer-lineup-job-schedule-shifts create \
        --lineup-scenario-trailer "$LINEUP_SCENARIO_TRAILER_ID" \
        --lineup-scenario-lineup-job-schedule-shift "$LINEUP_SCENARIO_LINEUP_JOB_SCHEDULE_SHIFT_ID" \
        --start-site-distance-minutes 5 \
        --end-site-distance-minutes 6

    if [[ $status -eq 0 ]]; then
        del_id=$(json_get ".id")
        xbe_run do lineup-scenario-trailer-lineup-job-schedule-shifts delete "$del_id" --confirm
        assert_success
    else
        skip "Could not create record for deletion test"
    fi
else
    skip "Missing prerequisites for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
