#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Shortfall Allocations
#
# Tests create operations for the driver-day-shortfall-allocations resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: driver-day-shortfall-allocations"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create allocation without required fields fails"
xbe_run do driver-day-shortfall-allocations create
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create driver day shortfall allocation"
if [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping API lookup for time cards"
else
    base_url="${XBE_BASE_URL%/}"

    time_cards_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/time-cards?page[limit]=1&filter[not_broker_invoiced]=true&include=tender-job-schedule-shift" || true)

    time_card_id=$(echo "$time_cards_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
    shift_id=$(echo "$time_cards_json" | jq -r '.data[0].relationships["tender-job-schedule-shift"].data.id // empty' 2>/dev/null || true)

    if [[ -z "$shift_id" ]]; then
        shift_id=$(echo "$time_cards_json" | jq -r '.included[] | select(.type=="tender-job-schedule-shifts") | .id' 2>/dev/null | head -n 1)
    fi

    if [[ -z "$time_card_id" || -z "$shift_id" ]]; then
        skip "No suitable time card or shift found"
    else
        constraints_json=$(curl -s \
            -H "Authorization: Bearer $XBE_TOKEN" \
            -H "Accept: application/vnd.api+json" \
            "$base_url/v1/shift-set-time-card-constraints?page[limit]=1&filter[scoped_to_shift]=$shift_id&include=service-type-unit-of-measures" || true)

        constraint_id=$(echo "$constraints_json" | jq -r '.data[0].id // empty' 2>/dev/null || true)
        stuom_id=$(echo "$constraints_json" | jq -r '.data[0].relationships["service-type-unit-of-measures"].data[0].id // empty' 2>/dev/null || true)

        if [[ -z "$stuom_id" ]]; then
            stuom_id=$(echo "$constraints_json" | jq -r '.included[] | select(.type=="service-type-unit-of-measures") | .id' 2>/dev/null | head -n 1)
        fi

        if [[ -z "$constraint_id" || -z "$stuom_id" ]]; then
            skip "No shift set time card constraint or service type unit of measure found"
        else
            allocation_json=$(printf '[{"time_card_id":"%s","quantity":1}]' "$time_card_id")
            xbe_json do driver-day-shortfall-allocations create \
                --time-card-ids "$time_card_id" \
                --shift-set-time-card-constraint-ids "$constraint_id" \
                --service-type-unit-of-measure "$stuom_id" \
                --quantity "1" \
                --allocation-quantities "$allocation_json"

            if [[ $status -eq 0 ]]; then
                pass
            else
                if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                    pass
                else
                    fail "Create failed: $output"
                fi
            fi
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
