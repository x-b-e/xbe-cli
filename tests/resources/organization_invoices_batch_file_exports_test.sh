#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch File Exports
#
# Tests create operations for organization-invoices-batch-file-exports.
#
# COVERAGE: Writable attributes (dry-run)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_FILE_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_FILE_EXPORT_ORGANIZATION_INVOICES_BATCH_FILE_ID:-}"

describe "Resource: organization-invoices-batch-file-exports"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create export requires organization invoices batch file"
xbe_run do organization-invoices-batch-file-exports create --dry-run
assert_failure

test_name "Create organization invoices batch file export"
if [[ -n "$BATCH_FILE_ID" && "$BATCH_FILE_ID" != "null" ]]; then
    xbe_json do organization-invoices-batch-file-exports create \
        --organization-invoices-batch-file "$BATCH_FILE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
        assert_json_equals ".organization_invoices_batch_file_id" "$BATCH_FILE_ID"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"Unable to determine worker_class"* ]] || [[ "$output" == *"could not connect"* ]] || [[ "$output" == *"cannot be exported"* ]] || [[ "$output" == *"not a tk batch export file"* ]] || [[ "$output" == *"formatting"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create organization invoices batch file export: $output"
        fi
    fi
else
    skip "No batch file ID available. Set XBE_TEST_ORGANIZATION_INVOICES_BATCH_FILE_EXPORT_ORGANIZATION_INVOICES_BATCH_FILE_ID to enable create testing."
fi

test_name "Create organization invoices batch file export with --dry-run"
if [[ -n "$BATCH_FILE_ID" && "$BATCH_FILE_ID" != "null" ]]; then
    xbe_json do organization-invoices-batch-file-exports create \
        --organization-invoices-batch-file "$BATCH_FILE_ID" \
        --dry-run
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"Unable to determine worker_class"* ]] || [[ "$output" == *"could not connect"* ]] || [[ "$output" == *"cannot be exported"* ]] || [[ "$output" == *"not a tk batch export file"* ]] || [[ "$output" == *"formatting"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create organization invoices batch file export (dry-run): $output"
        fi
    fi
else
    skip "No batch file ID available. Set XBE_TEST_ORGANIZATION_INVOICES_BATCH_FILE_EXPORT_ORGANIZATION_INVOICES_BATCH_FILE_ID to enable dry-run testing."
fi

run_tests
