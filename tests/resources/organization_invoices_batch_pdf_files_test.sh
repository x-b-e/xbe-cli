#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch PDF Files
#
# Tests view and download operations for organization-invoices-batch-pdf-files.
#
# COVERAGE: List filters + download
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_PDF_GENERATION_ID=""
SAMPLE_INVOICE_REVISION_ID=""
SAMPLE_STATUS=""
SAMPLE_CREATED_AT=""
DOWNLOAD_ID=""


describe "Resource: organization-invoices-batch-pdf-files"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List organization invoices batch PDF files"
xbe_json view organization-invoices-batch-pdf-files list --limit 5
assert_success

test_name "Capture sample organization invoices batch PDF file (if available)"
xbe_json view organization-invoices-batch-pdf-files list --limit 10
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_PDF_GENERATION_ID=$(json_get ".[0].pdf_generation_id")
        SAMPLE_INVOICE_REVISION_ID=$(json_get ".[0].invoice_revision_id")
        SAMPLE_STATUS=$(json_get ".[0].status")
        SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
        DOWNLOAD_ID=$(echo "$output" | jq -r '.[] | select(.status=="completed") | .id' 2>/dev/null | head -n1)
        pass
    else
        echo "    No organization invoices batch PDF files available; skipping show test."
        pass
    fi
else
    fail "Failed to list organization invoices batch PDF files"
fi

if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    test_name "Show organization invoices batch PDF file"
    xbe_json view organization-invoices-batch-pdf-files show "$SAMPLE_ID"
    assert_success
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List organization invoices batch PDF files with --pdf-generation filter"
if [[ -n "$SAMPLE_PDF_GENERATION_ID" && "$SAMPLE_PDF_GENERATION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-files list --pdf-generation "$SAMPLE_PDF_GENERATION_ID" --limit 5
    assert_success
else
    skip "No PDF generation ID available for filter test"
fi


test_name "List organization invoices batch PDF files with --invoice-revision filter"
if [[ -n "$SAMPLE_INVOICE_REVISION_ID" && "$SAMPLE_INVOICE_REVISION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-files list --invoice-revision "$SAMPLE_INVOICE_REVISION_ID" --limit 5
    assert_success
else
    skip "No invoice revision ID available for filter test"
fi


test_name "List organization invoices batch PDF files with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-files list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    skip "No status available for filter test"
fi


test_name "List organization invoices batch PDF files with --created-at-min filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-files list --created-at-min "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at timestamp available for filter test"
fi


test_name "List organization invoices batch PDF files with --created-at-max filter"
if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-files list --created-at-max "$SAMPLE_CREATED_AT" --limit 5
    assert_success
else
    skip "No created-at timestamp available for filter test"
fi

# ============================================================================
# DOWNLOAD Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping download tests)"
else
    test_name "Download organization invoices batch PDF file"
    if [[ -n "$DOWNLOAD_ID" && "$DOWNLOAD_ID" != "null" ]]; then
        tmpdir=$(mktemp -d -t xbe-pdf-XXXXXX)
        output_file="$tmpdir/organization-invoices-batch-pdf-file.pdf"
        xbe_run do organization-invoices-batch-pdf-files download "$DOWNLOAD_ID" --output "$output_file"
        if [[ $status -eq 0 ]]; then
            if [[ -s "$output_file" ]]; then
                pass
            else
                fail "Downloaded PDF file is empty"
            fi
        else
            fail "Failed to download PDF file"
        fi
        rm -rf "$tmpdir"
    else
        skip "No completed PDF file available for download test"
    fi

    test_name "Download organization invoices batch PDF file with invalid ID fails"
    tmpdir=$(mktemp -d -t xbe-pdf-XXXXXX)
    output_file="$tmpdir/invalid.pdf"
    xbe_run do organization-invoices-batch-pdf-files download "999999999" --output "$output_file"
    assert_failure
    rm -rf "$tmpdir"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
