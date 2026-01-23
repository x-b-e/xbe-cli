#!/bin/bash
#
# XBE CLI Integration Tests: Service Types
#
# Tests view operations for the service_types resource.
# Service types categorize different service offerings.
#
# COVERAGE: List + filters + pagination (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: service_types (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List service types"
xbe_json view service-types list
assert_success

test_name "List service types returns array"
xbe_json view service-types list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list service types"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List service types with --name filter"
xbe_json view service-types list --name "haul"
assert_success

test_name "List service types with --abbreviation filter"
xbe_json view service-types list --abbreviation "TL"
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
