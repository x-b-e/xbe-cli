#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Invoice Batchings
#
# Tests view and create operations for organization_invoices_batch_invoice_batchings.
# Batchings mark skipped or failed batch invoices as successful.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_INVOICE_ID=""
SECOND_BATCH_INVOICE_ID=""
SAMPLE_BATCHING_ID=""
SKIP_MUTATION=0

describe "Resource: organization-invoices-batch-invoice-batchings"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List organization invoices batch invoice batchings"
xbe_json view organization-invoices-batch-invoice-batchings list --limit 1
assert_success

test_name "Capture sample organization invoices batch invoice batching (if available)"
xbe_json view organization-invoices-batch-invoice-batchings list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_BATCHING_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No organization invoices batch invoice batchings available; skipping show test."
        pass
    fi
else
    fail "Failed to list organization invoices batch invoice batchings"
fi

if [[ -n "$SAMPLE_BATCHING_ID" && "$SAMPLE_BATCHING_ID" != "null" ]]; then
    test_name "Show organization invoices batch invoice batching"
    xbe_json view organization-invoices-batch-invoice-batchings show "$SAMPLE_BATCHING_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Create organization invoices batch invoice batching requires --organization-invoices-batch-invoice"
xbe_run do organization-invoices-batch-invoice-batchings create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find failed or skipped organization invoices batch invoice"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/organization-invoices-batch-invoices" \
        --data-urlencode "page[limit]=50" \
        --data-urlencode "fields[organization-invoices-batch-invoices]=status"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        mapfile -t candidate_ids < <(jq -r '.data[] | select(.attributes.status=="failed" or .attributes.status=="skipped") | .id' "$response_file" 2>/dev/null)
        BATCH_INVOICE_ID="${candidate_ids[0]}"
        SECOND_BATCH_INVOICE_ID="${candidate_ids[1]}"
        if [[ -n "$BATCH_INVOICE_ID" && "$BATCH_INVOICE_ID" != "null" ]]; then
            pass
        else
            skip "No failed or skipped organization invoices batch invoice found"
        fi
    else
        skip "Unable to list organization invoices batch invoices (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$BATCH_INVOICE_ID" && "$BATCH_INVOICE_ID" != "null" ]]; then
    if [[ -n "$SECOND_BATCH_INVOICE_ID" && "$SECOND_BATCH_INVOICE_ID" != "null" ]]; then
        test_name "Create organization invoices batch invoice batching (minimal)"
        xbe_json do organization-invoices-batch-invoice-batchings create --organization-invoices-batch-invoice "$BATCH_INVOICE_ID"
        assert_success

        test_name "Create organization invoices batch invoice batching with comment"
        COMMENT_TEXT="$(unique_name "OrgBatchInvoiceBatching")"
        xbe_json do organization-invoices-batch-invoice-batchings create \
            --organization-invoices-batch-invoice "$SECOND_BATCH_INVOICE_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create organization invoices batch invoice batching with comment"
        fi
    else
        test_name "Create organization invoices batch invoice batching with comment"
        COMMENT_TEXT="$(unique_name "OrgBatchInvoiceBatching")"
        xbe_json do organization-invoices-batch-invoice-batchings create \
            --organization-invoices-batch-invoice "$BATCH_INVOICE_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to create organization invoices batch invoice batching"
        fi
    fi
else
    test_name "Create organization invoices batch invoice batching"
    skip "No failed or skipped organization invoices batch invoice available for batching"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization invoices batch invoice batching with invalid ID fails"
xbe_run do organization-invoices-batch-invoice-batchings create --organization-invoices-batch-invoice "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
