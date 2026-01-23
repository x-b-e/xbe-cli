#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Rejections
#
# Tests create operations for the invoice_rejections resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INVOICE_ID=""

describe "Resource: invoice-rejections"

# ============================================================================
# Prerequisites - Find a sent invoice
# ============================================================================

test_name "Find sent invoice"
xbe_json view time-card-invoices list --invoice-status sent --limit 1

if [[ $status -eq 0 ]]; then
    INVOICE_ID=$(json_get ".[0].invoice_id")
    if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
        pass
    else
        skip "No sent invoice available"
    fi
else
    fail "Failed to list time card invoices"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create invoice rejection without required invoice fails"
xbe_run do invoice-rejections create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invoice rejection"
if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    COMMENT="Rejected via CLI test"

    xbe_json do invoice-rejections create \
        --invoice "$INVOICE_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".invoice_id" "$INVOICE_ID"
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create invoice rejection"
    fi
else
    skip "No sent invoice available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
