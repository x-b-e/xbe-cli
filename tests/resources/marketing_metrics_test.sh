#!/bin/bash
#
# XBE CLI Integration Tests: Marketing Metrics
#
# Tests view and create operations for the marketing-metrics resource.
#
# COVERAGE: create + list + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: marketing-metrics"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Refresh marketing metrics"
xbe_json do marketing-metrics create
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_has ".shift_count"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Create blocked by server policy"
    else
        fail "Failed to refresh marketing metrics: $output"
    fi
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List marketing metrics"
xbe_json view marketing-metrics list
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    assert_json_array_not_empty
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "List blocked by server policy"
    else
        fail "Failed to list marketing metrics: $output"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show marketing metrics"
xbe_json view marketing-metrics show
if [[ $status -eq 0 ]]; then
    assert_json_has ".id"
    assert_json_has ".shift_count"
else
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"401"* ]] || [[ "$output" == *"403"* ]]; then
        skip "Show blocked by server policy"
    else
        fail "Failed to show marketing metrics: $output"
    fi
fi

run_tests
