#!/bin/bash
#
# XBE CLI Integration Tests: Retainer Earning Statuses
#
# Tests list and show operations for the retainer_earning_statuses resource.
# Retainer earning statuses track expected and actual earnings for a retainer on a calculated date.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

STATUS_ID=""
RETAINER_ID=""
CALCULATED_ON=""
SKIP_ID_FILTERS=0

describe "Resource: retainer-earning-statuses"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List retainer earning statuses"
xbe_json view retainer-earning-statuses list --limit 5
assert_success

test_name "List retainer earning statuses returns array"
xbe_json view retainer-earning-statuses list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list retainer earning statuses"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample retainer earning status"
xbe_json view retainer-earning-statuses list --limit 1
if [[ $status -eq 0 ]]; then
    STATUS_ID=$(json_get ".[0].id")
    RETAINER_ID=$(json_get ".[0].retainer_id")
    CALCULATED_ON=$(json_get ".[0].calculated_on")
    if [[ -n "$STATUS_ID" && "$STATUS_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No retainer earning statuses available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list retainer earning statuses"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List retainer earning statuses with --retainer filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$RETAINER_ID" && "$RETAINER_ID" != "null" ]]; then
    xbe_json view retainer-earning-statuses list --retainer "$RETAINER_ID" --limit 5
    assert_success
else
    skip "No retainer ID available"
fi

test_name "List retainer earning statuses with --calculated-on filter"
if [[ -n "$CALCULATED_ON" && "$CALCULATED_ON" != "null" ]]; then
    xbe_json view retainer-earning-statuses list --calculated-on "$CALCULATED_ON" --limit 5
    assert_success
else
    skip "No calculated-on date available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show retainer earning status"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$STATUS_ID" && "$STATUS_ID" != "null" ]]; then
    xbe_json view retainer-earning-statuses show "$STATUS_ID"
    assert_success
else
    skip "No retainer earning status ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
