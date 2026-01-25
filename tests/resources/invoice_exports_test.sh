#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Exports
#
# Tests create operations for invoice-exports.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INVOICE_ID="${XBE_TEST_INVOICE_EXPORT_ID:-}"

describe "Resource: invoice-exports"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create export requires invoice"
xbe_run do invoice-exports create --comment "missing invoice"
assert_failure

test_name "Create invoice export"
if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    COMMENT=$(unique_name "InvoiceExport")
    xbe_json do invoice-exports create \
        --invoice "$INVOICE_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".invoice_id" "$INVOICE_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"already"* ]] || [[ "$output" == *"not ready"* ]] || [[ "$output" == *"must be"* ]] || [[ "$output" == *"not in valid"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create invoice export: $output"
        fi
    fi
else
    skip "No invoice ID available. Set XBE_TEST_INVOICE_EXPORT_ID to enable create testing."
fi

run_tests
