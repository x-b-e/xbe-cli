#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Addresses
#
# Tests create operations for the invoice_addresses resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INVOICE_ID=""

describe "Resource: invoice-addresses"

# ============================================================================
# Prerequisites - Find a rejected invoice
# ============================================================================

test_name "Find rejected invoice"
xbe_json view time-card-invoices list --invoice-status rejected --limit 1

if [[ $status -eq 0 ]]; then
    INVOICE_ID=$(json_get ".[0].invoice_id")
    if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
        pass
    else
        skip "No rejected invoice available"
    fi
else
    fail "Failed to list time card invoices"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create invoice address without required invoice fails"
xbe_run do invoice-addresses create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invoice address"
if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    COMMENT="Addressed via CLI test"

    xbe_json do invoice-addresses create \
        --invoice "$INVOICE_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".invoice_id" "$INVOICE_ID"
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create invoice address"
    fi
else
    skip "No rejected invoice available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
