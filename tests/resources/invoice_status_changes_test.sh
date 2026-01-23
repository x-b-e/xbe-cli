#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Status Changes
#
# Tests view operations for the invoice-status-changes resource.
# Invoice status changes record status transitions with timestamps and comments.
#
# COVERAGE: List + show + filters (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: invoice-status-changes (view-only)"

SAMPLE_ID=""
SAMPLE_INVOICE_ID=""
SAMPLE_STATUS=""

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice status changes"
xbe_json view invoice-status-changes list --limit 5
assert_success

test_name "List invoice status changes returns array"
xbe_json view invoice-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list invoice status changes"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample invoice status change"
xbe_json view invoice-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_INVOICE_ID=$(json_get ".[0].invoice_id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No invoice status changes available for follow-on tests"
    fi
else
    skip "Could not list invoice status changes to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoice status changes with --invoice filter"
if [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
    xbe_json view invoice-status-changes list --invoice "$SAMPLE_INVOICE_ID" --limit 5
    assert_success
else
    skip "No sample invoice ID available"
fi

test_name "List invoice status changes with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view invoice-status-changes list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No sample status available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show invoice status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view invoice-status-changes show "$SAMPLE_ID"
    assert_success
else
    skip "No invoice status change ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
