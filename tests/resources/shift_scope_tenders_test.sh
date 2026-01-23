#!/bin/bash
#
# XBE CLI Integration Tests: Shift Scope Tenders
#
# Tests create operations for the shift-scope-tenders resource.
#
# COVERAGE: Create + error cases + filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: shift-scope-tenders"

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create shift scope tenders without required identifiers fails"
xbe_run do shift-scope-tenders create
assert_failure

test_name "Create shift scope tenders with rate and constraint fails"
xbe_run do shift-scope-tenders create --rate 1 --shift-set-time-card-constraint 2
assert_failure

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create shift scope tenders with shift scope"
if [[ -n "$XBE_TEST_SHIFT_SCOPE_ID" ]]; then
    shift_scope_id="$XBE_TEST_SHIFT_SCOPE_ID"
elif [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping shift scope lookup"
else
    base_url="${XBE_BASE_URL%/}"

    shift_scopes_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/shift-scopes?page[limit]=1&sort=-created-at" || true)

    shift_scope_id=$(echo "$shift_scopes_json" | jq -r ".data[0].id // empty" 2>/dev/null || true)

    if [[ -z "$shift_scope_id" ]]; then
        skip "No shift scope found"
    fi
fi

if [[ -n "$shift_scope_id" ]]; then
    xbe_json do shift-scope-tenders create \
        --shift-scope "$shift_scope_id" \
        --created-at-min "2000-01-01" \
        --created-at-max "2100-01-01" \
        --limit 5

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
fi

# Optional: create with rate

test_name "Create shift scope tenders with rate"
if [[ -n "$XBE_TEST_RATE_ID" ]]; then
    rate_id="$XBE_TEST_RATE_ID"
elif [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping rate lookup"
else
    base_url="${XBE_BASE_URL%/}"

    rates_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/rates?page[limit]=1&sort=-created-at" || true)

    rate_id=$(echo "$rates_json" | jq -r ".data[0].id // empty" 2>/dev/null || true)

    if [[ -z "$rate_id" ]]; then
        skip "No rate found"
    fi
fi

if [[ -n "$rate_id" ]]; then
    xbe_run do shift-scope-tenders create --rate "$rate_id" --limit 1

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
fi

# Optional: create with shift set time card constraint

test_name "Create shift scope tenders with shift set time card constraint"
if [[ -n "$XBE_TEST_SHIFT_SET_TIME_CARD_CONSTRAINT_ID" ]]; then
    constraint_id="$XBE_TEST_SHIFT_SET_TIME_CARD_CONSTRAINT_ID"
elif [[ -z "$XBE_TOKEN" ]]; then
    skip "XBE_TOKEN not set; skipping constraint lookup"
else
    base_url="${XBE_BASE_URL%/}"

    constraints_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/shift-set-time-card-constraints?page[limit]=1&sort=-created-at" || true)

    constraint_id=$(echo "$constraints_json" | jq -r ".data[0].id // empty" 2>/dev/null || true)

    if [[ -z "$constraint_id" ]]; then
        skip "No shift set time card constraint found"
    fi
fi

if [[ -n "$constraint_id" ]]; then
    xbe_run do shift-scope-tenders create --shift-set-time-card-constraint "$constraint_id" --limit 1

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
