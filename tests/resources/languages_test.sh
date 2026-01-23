#!/bin/bash
#
# XBE CLI Integration Tests: Languages
#
# Tests view operations for the languages resource.
# Languages are supported locales in the system.
#
# COVERAGE: List + filters + pagination (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: languages (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List languages"
xbe_json view languages list
assert_success

test_name "List languages returns array"
xbe_json view languages list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list languages"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List languages with --code filter"
xbe_json view languages list --code "en"
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
