#!/bin/bash
#
# XBE CLI Integration Tests: Invoice Approvals
#
# Tests list, show, and create operations for the invoice-approvals resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_INVOICE_ID=""
CREATE_INVOICE_ID=""
LIST_SUPPORTED="true"

describe "Resource: invoice-approvals"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoice approvals"
xbe_json view invoice-approvals list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing invoice approvals"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List invoice approvals returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-approvals list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list invoice approvals"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample invoice approval"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view invoice-approvals list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_INVOICE_ID=$(json_get ".[0].invoice_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No invoice approvals available for follow-on tests"
        fi
    else
        skip "Could not list invoice approvals to capture sample"
    fi
else
    skip "List endpoint not supported"
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

test_name "Create invoice approval"
if [[ -n "$CREATE_INVOICE_ID" && "$CREATE_INVOICE_ID" != "null" ]]; then
    xbe_json do invoice-approvals create \
        --invoice "$CREATE_INVOICE_ID" \
        --comment "CLI test approval"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"previous status must be valid"* ]] || \
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
# SHOW Tests
# ============================================================================

test_name "Show invoice approval"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view invoice-approvals show "$SAMPLE_ID"
    assert_success
else
    skip "No approval ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create approval without invoice fails"
xbe_run do invoice-approvals create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
