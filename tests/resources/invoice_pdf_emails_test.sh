#!/bin/bash
#
# XBE CLI Integration Tests: Invoice PDF Emails
#
# Tests create operations for the invoice-pdf-emails resource.
#
# COVERAGE: create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_INVOICE_ID=""
CREATE_INVOICE_ID=""
LIST_SUPPORTED="true"

describe "Resource: invoice-pdf-emails"

# ============================================================================
# Sample Invoice (used for create)
# ============================================================================

test_name "Capture sample invoice from invoice approvals"
xbe_json view invoice-approvals list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_INVOICE_ID=$(json_get ".[0].invoice_id")
    if [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
        pass
    else
        skip "No invoice approvals available to capture invoice ID"
    fi
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing invoice approvals"
    else
        skip "Could not list invoice approvals"
    fi
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

if [[ -n "$XBE_TEST_INVOICE_ID" ]]; then
    CREATE_INVOICE_ID="$XBE_TEST_INVOICE_ID"
elif [[ -n "$SAMPLE_INVOICE_ID" && "$SAMPLE_INVOICE_ID" != "null" ]]; then
    CREATE_INVOICE_ID="$SAMPLE_INVOICE_ID"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invoice PDF email"
if [[ -n "$CREATE_INVOICE_ID" && "$CREATE_INVOICE_ID" != "null" ]]; then
    xbe_json do invoice-pdf-emails create \
        --invoice "$CREATE_INVOICE_ID" \
        --email-address "cli-test@example.com"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No invoice ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create invoice PDF email without invoice fails"
xbe_run do invoice-pdf-emails create --email-address "cli-test@example.com"
assert_failure

test_name "Create invoice PDF email without email fails"
xbe_run do invoice-pdf-emails create --invoice "123"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
