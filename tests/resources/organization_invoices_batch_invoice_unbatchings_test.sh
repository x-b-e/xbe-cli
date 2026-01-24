#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Invoice Unbatchings
#
# Tests list, show, and create operations for the organization-invoices-batch-invoice-unbatchings resource.
#
# COVERAGE: Create + list + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
LIST_SUPPORTED="false"
CREATED_UNBATCHING_ID=""
ENV_BATCH_INVOICE_ID="${XBE_TEST_ORG_INVOICES_BATCH_INVOICE_ID:-}"

describe "Resource: organization_invoices_batch_invoice_unbatchings"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List organization invoices batch invoice unbatchings"
xbe_json view organization-invoices-batch-invoice-unbatchings list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_SUPPORTED="true"
    pass
else
    if [[ "$output" == *"404"* ]]; then
        skip "Organization invoices batch invoice unbatchings list endpoint not available"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List organization invoices batch invoice unbatchings returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view organization-invoices-batch-invoice-unbatchings list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list organization invoices batch invoice unbatchings"
    fi
else
    skip "Organization invoices batch invoice unbatchings list endpoint not available"
fi

# ==========================================================================
# Sample Record (used for show)
# ==========================================================================

test_name "Capture sample organization invoices batch invoice unbatching"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view organization-invoices-batch-invoice-unbatchings list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No organization invoices batch invoice unbatchings available for show"
        fi
    else
        skip "Could not list organization invoices batch invoice unbatchings to capture sample"
    fi
else
    skip "Organization invoices batch invoice unbatchings list endpoint not available"
fi

# ============================================================================
# SHOW Tests (sample)
# ============================================================================

test_name "Show organization invoices batch invoice unbatching (sample)"
if [[ "$LIST_SUPPORTED" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoice-unbatchings show "$SAMPLE_ID"
    assert_success
else
    skip "No organization invoices batch invoice unbatching ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization invoices batch invoice unbatching without required batch invoice fails"
xbe_run do organization-invoices-batch-invoice-unbatchings create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization invoices batch invoice unbatching"
if [[ -n "$ENV_BATCH_INVOICE_ID" && "$ENV_BATCH_INVOICE_ID" != "null" ]]; then
    COMMENT="Unbatching batch invoice for test"

    xbe_json do organization-invoices-batch-invoice-unbatchings create \
        --organization-invoices-batch-invoice "$ENV_BATCH_INVOICE_ID" \
        --comment "$COMMENT"

    if [[ $status -eq 0 ]]; then
        CREATED_UNBATCHING_ID=$(json_get ".id")
        assert_json_equals ".organization_invoices_batch_invoice_id" "$ENV_BATCH_INVOICE_ID"
        assert_json_equals ".comment" "$COMMENT"
    else
        fail "Failed to create organization invoices batch invoice unbatching"
    fi
else
    skip "Set XBE_TEST_ORG_INVOICES_BATCH_INVOICE_ID to a successful or failed batch invoice to run"
fi

# ============================================================================
# SHOW Tests (created)
# ============================================================================

test_name "Show organization invoices batch invoice unbatching (created)"
SHOW_ID="$CREATED_UNBATCHING_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$SAMPLE_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoice-unbatchings show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        actual_id=$(json_get ".id")
        if [[ "$actual_id" == "$SHOW_ID" ]]; then
            pass
        else
            fail "Expected show id $SHOW_ID, got $actual_id"
        fi
    else
        fail "Failed to show organization invoices batch invoice unbatching"
    fi
else
    skip "No organization invoices batch invoice unbatching ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
