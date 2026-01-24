#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Invoice Status Changes
#
# Tests list/show operations for organization-invoices-batch-invoice-status-changes.
#
# COVERAGE: List filters (organization-invoices-batch-invoice, status) + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_INVOICE_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_STATUS_CHANGE_ORGANIZATION_INVOICES_BATCH_INVOICE_ID:-}"
SAMPLE_ID=""

describe "Resource: organization-invoices-batch-invoice-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List organization invoices batch invoice status changes"
xbe_json view organization-invoices-batch-invoice-status-changes list --limit 5
assert_success

test_name "List organization invoices batch invoice status changes returns array"
xbe_json view organization-invoices-batch-invoice-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list organization invoices batch invoice status changes"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List organization invoices batch invoice status changes with --status filter"
xbe_json view organization-invoices-batch-invoice-status-changes list --status successful --limit 5
assert_success

test_name "List organization invoices batch invoice status changes with --organization-invoices-batch-invoice filter"
if [[ -n "$BATCH_INVOICE_ID" && "$BATCH_INVOICE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoice-status-changes list --organization-invoices-batch-invoice "$BATCH_INVOICE_ID" --limit 5
    assert_success
else
    skip "No batch invoice ID available. Set XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_STATUS_CHANGE_ORGANIZATION_INVOICES_BATCH_INVOICE_ID to enable filter testing."
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample organization invoices batch invoice status change"
xbe_json view organization-invoices-batch-invoice-status-changes list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No organization invoices batch invoice status changes available for show test"
    fi
else
    skip "Could not list organization invoices batch invoice status changes to capture sample"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch invoice status change"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoice-status-changes show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show organization invoices batch invoice status change: $output"
        fi
    fi
else
    skip "No organization invoices batch invoice status change ID available for show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
