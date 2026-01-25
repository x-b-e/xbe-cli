#!/bin/bash
#
# XBE CLI Integration Tests: Integration Exports
#
# Tests view operations for integration exports.
#
# COVERAGE: List + filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

FIRST_EXPORT_ID=""
CREATED_BY_ID=""


describe "Resource: integration-exports (view-only)"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List integration exports"
xbe_json view integration-exports list --limit 5
assert_success

test_name "List integration exports returns array"
xbe_json view integration-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list integration exports"
fi

# Capture IDs for downstream tests
xbe_json view integration-exports list --limit 5
if [[ $status -eq 0 ]]; then
    FIRST_EXPORT_ID=$(json_get ".[0].id")
else
    FIRST_EXPORT_ID=""
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show integration export"
if [[ -n "$FIRST_EXPORT_ID" && "$FIRST_EXPORT_ID" != "null" ]]; then
    xbe_json view integration-exports show "$FIRST_EXPORT_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".created_by_id")
        pass
    else
        fail "Failed to show integration export"
    fi
else
    skip "No integration export ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List integration exports with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view integration-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
