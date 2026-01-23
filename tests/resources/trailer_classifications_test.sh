#!/bin/bash
#
# XBE CLI Integration Tests: Trailer Classifications
#
# Tests view operations for the trailer_classifications resource.
# Trailer classifications categorize different types of trailers.
#
# COVERAGE: List + pagination (view-only resource, no filters)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: trailer_classifications (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trailer classifications"
xbe_json view trailer-classifications list
assert_success

test_name "List trailer classifications returns array"
xbe_json view trailer-classifications list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trailer classifications"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
