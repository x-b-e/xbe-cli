#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Invoice Payments
#
# Tests view behavior for trucker-invoice-payments (QuickBooks bill payments).
#
# COVERAGE: List + list filters + show + required filter failure
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: trucker-invoice-payments"

SAMPLE_ID=""
TRUCKER_ID="${XBE_TEST_TRUCKER_ID:-}"
BROKER_ID=""

# ============================================================================
# LIST Tests - Required filter
# ============================================================================

test_name "List trucker invoice payments without filter fails"
xbe_run view trucker-invoice-payments list
assert_failure

# ============================================================================
# Resolve a trucker ID (QuickBooks-enabled broker if possible)
# ============================================================================

if [[ -z "$TRUCKER_ID" || "$TRUCKER_ID" == "null" ]]; then
    xbe_json view brokers list --quickbooks-enabled true --limit 1
    if [[ $status -eq 0 ]]; then
        BROKER_ID=$(json_get ".[0].id")
    fi

    if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
        xbe_json view truckers list --broker "$BROKER_ID" --limit 1
        if [[ $status -eq 0 ]]; then
            TRUCKER_ID=$(json_get ".[0].id")
        fi
    fi
fi

# ============================================================================
# LIST Tests - Basic + Filters
# ============================================================================

test_name "List trucker invoice payments with --trucker-id"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view trucker-invoice-payments list --trucker-id "$TRUCKER_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        skip "List failed; ensure trucker has QuickBooks payments and broker has QuickBooks enabled"
    fi
else
    skip "No trucker ID available (set XBE_TEST_TRUCKER_ID or ensure a QuickBooks-enabled broker has truckers)"
fi

test_name "List trucker invoice payments with --trucker filter"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view trucker-invoice-payments list --trucker "$TRUCKER_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        skip "List failed; ensure trucker has QuickBooks payments and broker has QuickBooks enabled"
    fi
else
    skip "No trucker ID available"
fi

# Capture sample ID for show test
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    test_name "Capture sample trucker invoice payment"
    xbe_json view trucker-invoice-payments list --trucker-id "$TRUCKER_ID" --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No trucker invoice payments available"
        fi
    else
        skip "Failed to list trucker invoice payments"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker invoice payment without filter fails"
xbe_run view trucker-invoice-payments show 123
assert_failure

test_name "Show trucker invoice payment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view trucker-invoice-payments show "$SAMPLE_ID" --trucker-id "$TRUCKER_ID"
    assert_success
else
    skip "No payment ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
