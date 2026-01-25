#!/bin/bash
#
# XBE CLI Integration Tests: Place Predictions
#
# Tests view operations for the place-predictions resource.
# Place predictions provide location autocomplete suggestions.
#
# COVERAGE: List + filters (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: place-predictions (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List place predictions"
xbe_json view place-predictions list --q "Austin"
assert_success

test_name "List place predictions returns array"
xbe_json view place-predictions list --q "Austin"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list place predictions"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List place predictions with --q filter"
xbe_json view place-predictions list --q "Dallas"
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
