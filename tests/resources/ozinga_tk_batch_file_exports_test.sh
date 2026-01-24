#!/bin/bash
#
# XBE CLI Integration Tests: Ozinga TK Batch File Exports
#
# Tests create operations for the ozinga-tk-batch-file-exports resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_FILE_ID=""

if [[ -n "$XBE_TEST_OZINGA_TK_BATCH_FILE_ID" ]]; then
    BATCH_FILE_ID="$XBE_TEST_OZINGA_TK_BATCH_FILE_ID"
elif [[ -n "$XBE_TOKEN" ]]; then
    base_url="${XBE_BASE_URL%/}"
    files_json=$(curl -s \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        "$base_url/v1/organization-invoices-batch-files?page[limit]=10&fields[organization-invoices-batch-files]=status" || true)

    BATCH_FILE_ID=$(echo "$files_json" | jq -r '.data[] | select(.attributes.status=="processed") | .id' 2>/dev/null | head -n 1)
fi

describe "Resource: ozinga-tk-batch-file-exports"

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create export without required fields fails"
xbe_run do ozinga-tk-batch-file-exports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create Ozinga TK batch file export"
if [[ -n "$BATCH_FILE_ID" && "$BATCH_FILE_ID" != "null" ]]; then
    xbe_json do ozinga-tk-batch-file-exports create \
        --organization-invoices-batch-file "$BATCH_FILE_ID" \
        --dry-run

    if [[ $status -eq 0 ]]; then
        assert_json_equals ".organization_invoices_batch_file_id" "$BATCH_FILE_ID"
        assert_json_bool ".dry_run" "true"
    else
        if [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to create export"
        elif [[ "$output" == *"is not a tk batch export file"* ]]; then
            skip "Batch file is not a TK batch export file"
        elif [[ "$output" == *"cannot be exported unless formatting is completed"* ]]; then
            skip "Batch file not processed"
        elif [[ "$output" == *"cannot be exported when formatting failed"* ]]; then
            skip "Batch file formatting failed"
        elif [[ "$output" == *"must be for a valid organization"* ]]; then
            skip "Batch file not for Ozinga organization"
        elif [[ "$output" == *"could not connect to exporter integration"* ]]; then
            skip "Exporter integration unavailable"
        else
            fail "Failed to create export"
        fi
    fi
else
    skip "Missing processed organization invoices batch file ID"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
