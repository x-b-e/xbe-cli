#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch PDF Generations
#
# Tests create, show, list filtering, and download-all for organization invoices batch PDF generations.
#
# COVERAGE: Create attributes + list filters + download-all
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_GENERATION_ID=""
LIST_GENERATION_ID=""
LIST_BATCH_ID=""
LIST_STATUS=""
LIST_CREATED_BY_ID=""
SHOW_CREATED_AT=""

ENV_BATCH_ID="${XBE_TEST_ORG_INVOICES_BATCH_ID:-}"
ENV_TEMPLATE_ID="${XBE_TEST_ORG_INVOICES_BATCH_PDF_TEMPLATE_ID:-}"
ENV_DOWNLOAD_ID="${XBE_TEST_ORG_INVOICES_BATCH_PDF_GENERATION_ID:-}"

describe "Resource: organization_invoices_batch_pdf_generations"

# ============================================================================
# LIST Tests (smoke)
# ============================================================================

test_name "List organization invoices batch PDF generations"
xbe_json view organization-invoices-batch-pdf-generations list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_GENERATION_ID=$(json_get ".[0].id")
    LIST_BATCH_ID=$(json_get ".[0].organization_invoices_batch_id")
    LIST_STATUS=$(json_get ".[0].status")
    LIST_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$LIST_GENERATION_ID" && "$LIST_GENERATION_ID" != "null" ]]; then
        pass
    else
        skip "No organization invoices batch PDF generations returned"
    fi
else
    fail "Failed to list organization invoices batch PDF generations"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization invoices batch PDF generation without required flags fails"
xbe_run do organization-invoices-batch-pdf-generations create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization invoices batch PDF generation"
if [[ -n "$ENV_BATCH_ID" && -n "$ENV_TEMPLATE_ID" ]]; then
    xbe_json do organization-invoices-batch-pdf-generations create \
        --organization-invoices-batch "$ENV_BATCH_ID" \
        --organization-pdf-template "$ENV_TEMPLATE_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_GENERATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_GENERATION_ID" && "$CREATED_GENERATION_ID" != "null" ]]; then
            pass
        else
            fail "Created organization invoices batch PDF generation but no ID returned"
        fi
    else
        fail "Failed to create organization invoices batch PDF generation"
    fi
else
    skip "Set XBE_TEST_ORG_INVOICES_BATCH_ID and XBE_TEST_ORG_INVOICES_BATCH_PDF_TEMPLATE_ID to run"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch PDF generation details"
SHOW_ID="$CREATED_GENERATION_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$LIST_GENERATION_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        actual_id=$(json_get ".id")
        SHOW_CREATED_AT=$(json_get ".created_at")
        if [[ "$actual_id" == "$SHOW_ID" ]]; then
            pass
        else
            fail "Expected show id $SHOW_ID, got $actual_id"
        fi
    else
        fail "Failed to show organization invoices batch PDF generation"
    fi
else
    skip "No organization invoices batch PDF generation ID available"
fi

# ============================================================================
# LIST Filter Tests
# ============================================================================

test_name "List organization invoices batch PDF generations filtered by batch"
FILTER_BATCH_ID="$ENV_BATCH_ID"
FILTER_GENERATION_ID="$CREATED_GENERATION_ID"
if [[ -z "$FILTER_BATCH_ID" || "$FILTER_BATCH_ID" == "null" ]]; then
    FILTER_BATCH_ID="$LIST_BATCH_ID"
    FILTER_GENERATION_ID="$LIST_GENERATION_ID"
fi

if [[ -n "$FILTER_BATCH_ID" && "$FILTER_BATCH_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations list --organization-invoices-batch "$FILTER_BATCH_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_GENERATION_ID" && "$FILTER_GENERATION_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_GENERATION_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected PDF generation in filtered results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batch PDF generations"
    fi
else
    skip "No batch ID available for filter test"
fi

test_name "List organization invoices batch PDF generations filtered by status"
FILTER_STATUS="$LIST_STATUS"
FILTER_STATUS_ID="$LIST_GENERATION_ID"
if [[ -z "$FILTER_STATUS" || "$FILTER_STATUS" == "null" ]]; then
    FILTER_STATUS=""
fi

if [[ -n "$FILTER_STATUS" && "$FILTER_STATUS_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations list --status "$FILTER_STATUS"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$FILTER_STATUS_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
            pass
        else
            fail "Expected PDF generation in status filtered results"
        fi
    else
        fail "Failed to list organization invoices batch PDF generations by status"
    fi
else
    skip "No status available for filter test"
fi

test_name "List organization invoices batch PDF generations filtered by created-by"
if [[ -n "$LIST_CREATED_BY_ID" && "$LIST_CREATED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations list --created-by "$LIST_CREATED_BY_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$LIST_GENERATION_ID" && "$LIST_GENERATION_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$LIST_GENERATION_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected PDF generation in created-by filtered results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batch PDF generations by created-by"
    fi
else
    skip "No created-by ID available for filter test"
fi

test_name "List organization invoices batch PDF generations filtered by created-at-min"
if [[ -n "$SHOW_CREATED_AT" && "$SHOW_CREATED_AT" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations list --created-at-min "$SHOW_CREATED_AT"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$SHOW_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected PDF generation in created-at-min filtered results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batch PDF generations by created-at-min"
    fi
else
    skip "No created-at value available for filter test"
fi

test_name "List organization invoices batch PDF generations filtered by created-at-max"
if [[ -n "$SHOW_CREATED_AT" && "$SHOW_CREATED_AT" != "null" ]]; then
    xbe_json view organization-invoices-batch-pdf-generations list --created-at-max "$SHOW_CREATED_AT"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$SHOW_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected PDF generation in created-at-max filtered results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batch PDF generations by created-at-max"
    fi
else
    skip "No created-at value available for filter test"
fi

# ============================================================================
# DOWNLOAD Tests
# ============================================================================

test_name "Download all PDFs for a generation"
if [[ -n "$ENV_DOWNLOAD_ID" && "$ENV_DOWNLOAD_ID" != "null" ]]; then
    tmpfile=$(mktemp)
    xbe_run view organization-invoices-batch-pdf-generations download-all "$ENV_DOWNLOAD_ID" --output "$tmpfile" --overwrite
    if [[ $status -eq 0 ]]; then
        if [[ -s "$tmpfile" ]]; then
            pass
        else
            fail "Download succeeded but file is empty"
        fi
    else
        fail "Failed to download PDF archive"
    fi
    rm -f "$tmpfile"
else
    skip "Set XBE_TEST_ORG_INVOICES_BATCH_PDF_GENERATION_ID with completed PDFs to run"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
