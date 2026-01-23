#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Revisionables
#
# Tests create operations for the invoice_revisionables resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INVOICE_ID=""
INVOICE_STATUS=""

describe "Resource: invoice-revisionables"

# ============================================================================
# Prerequisites - Find an eligible invoice
# ============================================================================

test_name "Find eligible invoice"
found=false
error=false

for candidate_status in revised exported non_exportable; do
    xbe_json view time-card-invoices list --invoice-status "$candidate_status" --limit 1

    if [[ $status -ne 0 ]]; then
        error=true
        break
    fi

    INVOICE_ID=$(json_get ".[0].invoice_id")
    if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
        INVOICE_STATUS="$candidate_status"
        found=true
        break
    fi
done

if [[ "$error" == "true" ]]; then
    fail "Failed to list time card invoices"
elif [[ "$found" == "true" ]]; then
    pass
else
    skip "No eligible invoice available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create invoice revisionable without required invoice fails"
xbe_run do invoice-revisionables create --comment "Missing invoice"
assert_failure

test_name "Create invoice revisionable without required comment fails"
xbe_run do invoice-revisionables create --invoice 123
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invoice revisionable"
if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    COMMENT="Revisionable via CLI test"

    xbe_json do invoice-revisionables create \
        --invoice "$INVOICE_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".invoice_id" "$INVOICE_ID"
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create invoice revisionable"
    fi
else
    skip "No eligible invoice available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
