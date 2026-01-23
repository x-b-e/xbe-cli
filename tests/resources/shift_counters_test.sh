#!/bin/bash
#
# XBE CLI Integration Tests: Shift Counters
#
# Tests list and create behavior for shift counters.
#
# COVERAGE: List + create + optional start-at-min
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: shift-counters"

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List shift counters"
xbe_json view shift-counters list
assert_success

test_name "List shift counters returns array"
xbe_json view shift-counters list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list shift counters"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create shift counter (default start)"
xbe_json do shift-counters create
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to create shift counter"
fi

test_name "Create shift counter with --start-at-min"
xbe_json do shift-counters create --start-at-min "2025-01-01T00:00:00Z"
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
else
    fail "Failed to create shift counter with start-at-min"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
