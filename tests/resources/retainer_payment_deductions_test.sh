#!/bin/bash
#
# XBE CLI Integration Tests: Retainer Payment Deductions
#
# Tests list and show operations for the retainer-payment-deductions resource.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_DEDUCTION_ID=""

describe "Resource: retainer-payment-deductions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List retainer payment deductions"
xbe_json view retainer-payment-deductions list --limit 5
assert_success

test_name "List retainer payment deductions returns array"
xbe_json view retainer-payment-deductions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list retainer payment deductions"
fi

test_name "Capture sample retainer payment deduction (if available)"
xbe_json view retainer-payment-deductions list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_DEDUCTION_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No retainer payment deductions available; show test will be skipped."
        pass
    fi
else
    fail "Failed to list retainer payment deductions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List retainer payment deductions with --created-at-min filter"
xbe_json view retainer-payment-deductions list --created-at-min "2000-01-01T00:00:00Z" --limit 5
assert_success

test_name "List retainer payment deductions with --created-at-max filter"
xbe_json view retainer-payment-deductions list --created-at-max "2099-12-31T23:59:59Z" --limit 5
assert_success

test_name "List retainer payment deductions with --updated-at-min filter"
xbe_json view retainer-payment-deductions list --updated-at-min "2000-01-01T00:00:00Z" --limit 5
assert_success

test_name "List retainer payment deductions with --updated-at-max filter"
xbe_json view retainer-payment-deductions list --updated-at-max "2099-12-31T23:59:59Z" --limit 5
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List retainer payment deductions with --offset"
xbe_json view retainer-payment-deductions list --limit 5 --offset 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_DEDUCTION_ID" && "$SAMPLE_DEDUCTION_ID" != "null" ]]; then
    test_name "Show retainer payment deduction"
    xbe_json view retainer-payment-deductions show "$SAMPLE_DEDUCTION_ID"
    assert_success
else
    test_name "Show retainer payment deduction"
    skip "No retainer payment deductions available to show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
