#!/bin/bash
#
# XBE CLI Integration Tests: Unit of Measures
#
# Tests view operations for the unit_of_measures resource.
# Unit of measures define measurement units used in the system.
#
# COVERAGE: List + filters + pagination (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: unit_of_measures (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List unit of measures"
xbe_json view unit-of-measures list
assert_success

test_name "List unit of measures returns array"
xbe_json view unit-of-measures list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list unit of measures"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List unit of measures with --metric filter"
xbe_json view unit-of-measures list --metric "mass"
assert_success

test_name "List unit of measures with --name filter"
xbe_json view unit-of-measures list --name "ton"
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
