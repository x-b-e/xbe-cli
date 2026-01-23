#!/bin/bash
#
# XBE CLI Integration Tests: Shift Scope Matches
#
# Tests create operations for the shift-scope-matches resource.
# Requires a tender with a rate or shift set time card constraint.
#
# COVERAGE: create attributes + relationships + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TENDER_ID=""
RATE_ID=""
CONSTRAINT_ID=""
TENDER_JOB_SCHEDULE_SHIFT_ID=""
SHIFT_SCOPE_ID=""

DIRECT_API_AVAILABLE=0
if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

api_get() {
    local path="$1"
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        output="Missing XBE_TOKEN for direct API calls"
        status=1
        return
    fi
    run curl -sS -X GET "$XBE_BASE_URL$path" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json"
}

describe "Resource: shift-scope-matches"

# ============================================================================
# Sample data lookup (tender + rate/constraint)
# ============================================================================

test_name "Lookup tender with rates or constraints"
if [[ $DIRECT_API_AVAILABLE -eq 1 ]]; then
    api_get "/v1/tenders?page[limit]=10&fields[tenders]=rates,shift-set-time-card-constraints,tender-job-schedule-shifts"
    if [[ $status -eq 0 ]]; then
        TENDER_ID=$(echo "$output" | jq -r '.data[] | select(((.relationships.rates.data // []) | length) > 0 or ((.relationships["shift-set-time-card-constraints"].data // []) | length) > 0) | .id' | head -n1)
        if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" ]]; then
            RATE_ID=$(echo "$output" | jq -r --arg id "$TENDER_ID" '.data[] | select(.id == $id) | .relationships.rates.data[0].id // empty' | head -n1)
            CONSTRAINT_ID=$(echo "$output" | jq -r --arg id "$TENDER_ID" '.data[] | select(.id == $id) | .relationships["shift-set-time-card-constraints"].data[0].id // empty' | head -n1)
            TENDER_JOB_SCHEDULE_SHIFT_ID=$(echo "$output" | jq -r --arg id "$TENDER_ID" '.data[] | select(.id == $id) | .relationships["tender-job-schedule-shifts"].data[0].id // empty' | head -n1)
            pass
        else
            skip "No tender with rates or constraints available"
        fi
    else
        skip "Failed to fetch tenders"
    fi
else
    skip "XBE_TOKEN not set for direct API access"
fi

test_name "Lookup shift scope for rate"
if [[ -n "$RATE_ID" && $DIRECT_API_AVAILABLE -eq 1 ]]; then
    api_get "/v1/rates/$RATE_ID?fields[rates]=shift-scope"
    if [[ $status -eq 0 ]]; then
        SHIFT_SCOPE_ID=$(echo "$output" | jq -r '.data.relationships["shift-scope"].data.id // empty')
        if [[ -n "$SHIFT_SCOPE_ID" && "$SHIFT_SCOPE_ID" != "null" ]]; then
            pass
        else
            skip "No shift scope on rate"
        fi
    else
        skip "Failed to fetch rate"
    fi
else
    skip "No rate ID available"
fi

# ============================================================================
# Error cases
# ============================================================================

test_name "Create shift scope match requires --tender"
xbe_run do shift-scope-matches create --rate "123"
assert_failure

test_name "Create shift scope match requires --rate or --shift-set-time-card-constraint"
xbe_run do shift-scope-matches create --tender "123"
assert_failure

test_name "Create shift scope match forbids --rate with --shift-set-time-card-constraint"
xbe_run do shift-scope-matches create --tender "123" --rate "1" --shift-set-time-card-constraint "2"
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create shift scope match with rate"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" && -n "$RATE_ID" && "$RATE_ID" != "null" ]]; then
    cmd=(do shift-scope-matches create --tender "$TENDER_ID" --rate "$RATE_ID" --show-matching-shift-sql)
    if [[ -n "$SHIFT_SCOPE_ID" && "$SHIFT_SCOPE_ID" != "null" ]]; then
        cmd+=(--shift-scope "$SHIFT_SCOPE_ID")
    fi
    if [[ -n "$TENDER_JOB_SCHEDULE_SHIFT_ID" && "$TENDER_JOB_SCHEDULE_SHIFT_ID" != "null" ]]; then
        cmd+=(--tender-job-schedule-shift "$TENDER_JOB_SCHEDULE_SHIFT_ID")
    fi
    xbe_json "${cmd[@]}"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create shift scope match with rate"
    fi
else
    skip "No tender/rate available"
fi

test_name "Create shift scope match with shift set time card constraint"
if [[ -n "$TENDER_ID" && "$TENDER_ID" != "null" && -n "$CONSTRAINT_ID" && "$CONSTRAINT_ID" != "null" ]]; then
    xbe_json do shift-scope-matches create \
        --tender "$TENDER_ID" \
        --shift-set-time-card-constraint "$CONSTRAINT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create shift scope match with constraint"
    fi
else
    skip "No tender/constraint available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
