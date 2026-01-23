#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Movement Trip Job Production Plans
#
# Tests list, show, create, and delete operations for the equipment-movement-trip-job-production-plans resource.
#
# COVERAGE: List filters + show + create/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TRIP_ID=""
SAMPLE_JOB_PRODUCTION_PLAN_ID=""

CREATED_ID=""
CREATE_TRIP_ID=""

describe "Resource: equipment-movement-trip-job-production-plans"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment movement trip job production plans"
xbe_json view equipment-movement-trip-job-production-plans list --limit 5
assert_success

test_name "List equipment movement trip job production plans returns array"
xbe_json view equipment-movement-trip-job-production-plans list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment movement trip job production plans"
fi

# ============================================================================
# Sample Record (used for filters/show/create)
# ============================================================================

test_name "Capture sample link"
xbe_json view equipment-movement-trip-job-production-plans list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TRIP_ID=$(json_get ".[0].equipment_movement_trip_id")
    SAMPLE_JOB_PRODUCTION_PLAN_ID=$(json_get ".[0].job_production_plan_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No links available for follow-on tests"
    fi
else
    skip "Could not list links to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List links with --equipment-movement-trip filter"
if [[ -n "$SAMPLE_TRIP_ID" && "$SAMPLE_TRIP_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-job-production-plans list --equipment-movement-trip "$SAMPLE_TRIP_ID" --limit 5
    assert_success
else
    skip "No equipment movement trip ID available"
fi

test_name "List links with --job-production-plan filter"
if [[ -n "$SAMPLE_JOB_PRODUCTION_PLAN_ID" && "$SAMPLE_JOB_PRODUCTION_PLAN_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-job-production-plans list --job-production-plan "$SAMPLE_JOB_PRODUCTION_PLAN_ID" --limit 5
    assert_success
else
    skip "No job production plan ID available"
fi

test_name "List links with --created-at-min filter"
xbe_json view equipment-movement-trip-job-production-plans list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List links with --created-at-max filter"
xbe_json view equipment-movement-trip-job-production-plans list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "List links with --updated-at-min filter"
xbe_json view equipment-movement-trip-job-production-plans list --updated-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "List links with --updated-at-max filter"
xbe_json view equipment-movement-trip-job-production-plans list --updated-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show equipment movement trip job production plan link"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view equipment-movement-trip-job-production-plans show "$SAMPLE_ID"
    assert_success
else
    skip "No link ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create link with equipment movement trip"
if [[ -n "$SAMPLE_TRIP_ID" && "$SAMPLE_TRIP_ID" != "null" ]]; then
    CREATE_TRIP_ID="$SAMPLE_TRIP_ID"
elif [[ -n "$XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID" ]]; then
    CREATE_TRIP_ID="$XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID"
    echo "    Using XBE_TEST_EQUIPMENT_MOVEMENT_TRIP_ID: $CREATE_TRIP_ID"
fi

if [[ -n "$CREATE_TRIP_ID" && "$CREATE_TRIP_ID" != "null" ]]; then
    xbe_json do equipment-movement-trip-job-production-plans create --equipment-movement-trip "$CREATE_TRIP_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "equipment-movement-trip-job-production-plans" "$CREATED_ID"
            pass
        else
            fail "Created link but no ID returned"
        fi
    else
        if [[ "$output" == *"valid for dispatch"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Create blocked by policy or trip invalid for dispatch"
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No equipment movement trip ID available for create"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete link requires --confirm flag"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do equipment-movement-trip-job-production-plans delete "$CREATED_ID"
    assert_failure
else
    skip "No created link ID available"
fi

test_name "Delete link with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do equipment-movement-trip-job-production-plans delete "$CREATED_ID" --confirm
    assert_success
else
    skip "No created link ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create link without required fields fails"
xbe_run do equipment-movement-trip-job-production-plans create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
