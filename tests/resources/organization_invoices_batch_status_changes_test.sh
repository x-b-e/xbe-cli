#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Status Changes
#
# Tests view operations for organization-invoices-batch-status-changes.
#
# COVERAGE: List filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_BATCH_ID=""
SAMPLE_STATUS=""

describe "Resource: organization-invoices-batch-status-changes"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List organization invoices batch status changes"
xbe_json view organization-invoices-batch-status-changes list --limit 5
assert_success

test_name "Capture sample organization invoices batch status change (if available)"
xbe_json view organization-invoices-batch-status-changes list --limit 10
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_BATCH_ID=$(json_get ".[0].organization_invoices_batch_id")
        SAMPLE_STATUS=$(json_get ".[0].status")
        pass
    else
        echo "    No organization invoices batch status changes available; skipping show test."
        pass
    fi
else
    fail "Failed to list organization invoices batch status changes"
fi

if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    test_name "Show organization invoices batch status change"
    xbe_json view organization-invoices-batch-status-changes show "$SAMPLE_ID"
    assert_success
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List organization invoices batch status changes with --organization-invoices-batch filter"
if [[ -n "$SAMPLE_BATCH_ID" && "$SAMPLE_BATCH_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-status-changes list --organization-invoices-batch "$SAMPLE_BATCH_ID" --limit 5
    assert_success
else
    skip "No organization invoices batch ID available for filter test"
fi


test_name "List organization invoices batch status changes with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view organization-invoices-batch-status-changes list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available for filter test"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
