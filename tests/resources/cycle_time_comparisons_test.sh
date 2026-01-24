#!/bin/bash
#
# XBE CLI Integration Tests: Cycle Time Comparisons
#
# Tests create operations for the cycle-time-comparisons resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: cycle-time-comparisons"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create cycle time comparison without required fields fails"
xbe_run do cycle-time-comparisons create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create cycle time comparison"

coordinates_one="${XBE_TEST_CYCLE_TIME_COORDINATES_ONE:-}"
coordinates_two="${XBE_TEST_CYCLE_TIME_COORDINATES_TWO:-}"
proximity_meters="${XBE_TEST_CYCLE_TIME_PROXIMITY_METERS:-5000}"

if [[ -z "$coordinates_one" || -z "$coordinates_two" ]]; then
    if [[ -z "$XBE_TOKEN" ]]; then
        skip "XBE_TOKEN not set; skipping coordinate lookup"
    else
        base_url="${XBE_BASE_URL%/}"

        job_sites_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/job-sites?page[limit]=1&fields[job-sites]=address-latitude,address-longitude" || true)

        job_lat=$(echo "$job_sites_json" | jq -r '.data[0].attributes["address-latitude"] // empty' 2>/dev/null || true)
        job_lon=$(echo "$job_sites_json" | jq -r '.data[0].attributes["address-longitude"] // empty' 2>/dev/null || true)

        material_sites_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/material-sites?page[limit]=1&fields[material-sites]=address-latitude,address-longitude" || true)

        material_lat=$(echo "$material_sites_json" | jq -r '.data[0].attributes["address-latitude"] // empty' 2>/dev/null || true)
        material_lon=$(echo "$material_sites_json" | jq -r '.data[0].attributes["address-longitude"] // empty' 2>/dev/null || true)

        if [[ -n "$job_lat" && -n "$job_lon" && -z "$coordinates_one" ]]; then
            coordinates_one="[$job_lat,$job_lon]"
        fi

        if [[ -n "$material_lat" && -n "$material_lon" && -z "$coordinates_two" ]]; then
            coordinates_two="[$material_lat,$material_lon]"
        fi
    fi
fi

if [[ -n "$coordinates_one" && -n "$coordinates_two" ]]; then
    transaction_at_min="${XBE_TEST_CYCLE_TIME_TRANSACTION_AT_MIN:-2024-01-01T00:00:00Z}"
    transaction_at_max="${XBE_TEST_CYCLE_TIME_TRANSACTION_AT_MAX:-2024-12-31T23:59:59Z}"
    cycle_count="${XBE_TEST_CYCLE_TIME_CYCLE_COUNT:-100}"

    xbe_json do cycle-time-comparisons create \
        --coordinates-one "$coordinates_one" \
        --coordinates-two "$coordinates_two" \
        --proximity-meters "$proximity_meters" \
        --transaction-at-min "$transaction_at_min" \
        --transaction-at-max "$transaction_at_max" \
        --cycle-count "$cycle_count"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_has ".coordinates_one"
        assert_json_has ".coordinates_two"
        assert_json_has ".proximity_meters"
    else
        fail "Failed to create cycle time comparison"
    fi
else
    skip "Missing coordinates for cycle time comparison"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
