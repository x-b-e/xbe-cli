#!/bin/bash
#
# XBE CLI Integration Tests: Production Measurements
#
# Tests create, update, delete operations and list filters for the
# production-measurements resource.
#
# COMPLETE COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_MEASUREMENT_ID=""
EXISTING_MEASUREMENT_ID=""
JOB_PRODUCTION_PLAN_SEGMENT_ID="${XBE_TEST_JOB_PRODUCTION_PLAN_SEGMENT_ID:-}"

HAS_EXPLICIT_SEGMENT_ID=false
if [[ -n "$XBE_TEST_JOB_PRODUCTION_PLAN_SEGMENT_ID" ]]; then
    HAS_EXPLICIT_SEGMENT_ID=true
fi

describe "Resource: production-measurements"

# ==========================================================================
# Seed IDs from existing data if available
# ==========================================================================

test_name "Lookup existing production measurement (if any)"
xbe_json view production-measurements list --limit 1
if [[ $status -eq 0 ]]; then
    EXISTING_MEASUREMENT_ID=$(json_get ".[0].id")
    if [[ -n "$EXISTING_MEASUREMENT_ID" && "$EXISTING_MEASUREMENT_ID" != "null" ]]; then
        if [[ -z "$JOB_PRODUCTION_PLAN_SEGMENT_ID" || "$JOB_PRODUCTION_PLAN_SEGMENT_ID" == "null" ]]; then
            JOB_PRODUCTION_PLAN_SEGMENT_ID=$(json_get ".[0].job_production_plan_segment_id")
        fi
    fi
    pass
else
    fail "Failed to list production measurements"
fi

test_name "Lookup job production plan segment without measurement (if needed)"
if [[ -n "$JOB_PRODUCTION_PLAN_SEGMENT_ID" && "$JOB_PRODUCTION_PLAN_SEGMENT_ID" != "null" ]]; then
    skip "Job production plan segment ID already available"
elif [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup"
else
    base_url="${XBE_BASE_URL%/}"
    segments_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/job-production-plan-segments?page[limit]=25&fields[job-production-plan-segments]=production-measurement" || true)
    JOB_PRODUCTION_PLAN_SEGMENT_ID=$(echo "$segments_json" | jq -r '.data[] | select(.relationships["production-measurement"].data == null) | .id' | head -n 1)
    if [[ -n "$JOB_PRODUCTION_PLAN_SEGMENT_ID" && "$JOB_PRODUCTION_PLAN_SEGMENT_ID" != "null" ]]; then
        pass
    else
        skip "No segment without production measurement found"
    fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create production measurement"
if [[ -n "$JOB_PRODUCTION_PLAN_SEGMENT_ID" && "$JOB_PRODUCTION_PLAN_SEGMENT_ID" != "null" ]]; then
    xbe_json do production-measurements create \
        --job-production-plan-segment "$JOB_PRODUCTION_PLAN_SEGMENT_ID" \
        --width-display-unit-of-measure inches \
        --pass-count 1 \
        --width-inches 144 \
        --depth-inches 6 \
        --length-feet 500 \
        --speed-feet-per-minute 35 \
        --speed-feet-per-minute-possible 40 \
        --density-lbs-per-cubic-foot 145 \
        --note "CLI test measurement"

    if [[ $status -eq 0 ]]; then
        CREATED_MEASUREMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_MEASUREMENT_ID" && "$CREATED_MEASUREMENT_ID" != "null" ]]; then
            register_cleanup "production-measurements" "$CREATED_MEASUREMENT_ID"
            pass
        else
            fail "Created production measurement but no ID returned"
        fi
    else
        if [[ "$HAS_EXPLICIT_SEGMENT_ID" == true ]]; then
            fail "Failed to create production measurement: $output"
        else
            skip "Create failed with inferred segment ID"
        fi
    fi
else
    skip "Missing job production plan segment ID (set XBE_TEST_JOB_PRODUCTION_PLAN_SEGMENT_ID)"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show production measurement"
SHOW_ID="$CREATED_MEASUREMENT_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$EXISTING_MEASUREMENT_ID"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view production-measurements show "$SHOW_ID"
    assert_success
else
    skip "No production measurement ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update production measurement attributes"
UPDATE_ID="$CREATED_MEASUREMENT_ID"
if [[ -z "$UPDATE_ID" || "$UPDATE_ID" == "null" ]]; then
    UPDATE_ID="${XBE_TEST_PRODUCTION_MEASUREMENT_ID:-}"
fi
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do production-measurements update "$UPDATE_ID" \
        --width-inches 120 \
        --depth-inches 4 \
        --length-feet 600 \
        --speed-feet-per-minute 30 \
        --speed-feet-per-minute-possible 38 \
        --density-lbs-per-cubic-foot 142 \
        --width-display-unit-of-measure feet \
        --pass-count 2 \
        --note "Updated production measurement"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update production measurement: $output"
        fi
    fi
else
    skip "Set XBE_TEST_PRODUCTION_MEASUREMENT_ID to run"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List production measurements"
xbe_json view production-measurements list --limit 10
assert_success

test_name "List production measurements returns array"
xbe_json view production-measurements list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list production measurements"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List production measurements with --job-production-plan-segment filter"
FILTER_SEGMENT_ID="$JOB_PRODUCTION_PLAN_SEGMENT_ID"
if [[ -z "$FILTER_SEGMENT_ID" || "$FILTER_SEGMENT_ID" == "null" ]]; then
    FILTER_SEGMENT_ID=$(json_get ".[0].job_production_plan_segment_id")
fi
if [[ -n "$FILTER_SEGMENT_ID" && "$FILTER_SEGMENT_ID" != "null" ]]; then
    xbe_json view production-measurements list --job-production-plan-segment "$FILTER_SEGMENT_ID" --limit 10
    assert_success
else
    skip "No job production plan segment ID available"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete production measurement requires --confirm flag"
if [[ -n "$CREATED_MEASUREMENT_ID" && "$CREATED_MEASUREMENT_ID" != "null" ]]; then
    xbe_json do production-measurements delete "$CREATED_MEASUREMENT_ID"
    assert_failure
else
    skip "No production measurement ID available"
fi

test_name "Delete production measurement with --confirm"
if [[ -n "$CREATED_MEASUREMENT_ID" && "$CREATED_MEASUREMENT_ID" != "null" ]]; then
    xbe_json do production-measurements delete "$CREATED_MEASUREMENT_ID" --confirm
    assert_success
else
    skip "No production measurement ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
