#!/bin/bash
#
# XBE CLI Integration Tests: Integration Invoices Batch Exports
#
# Tests view operations for integration invoices batch exports.
#
# COVERAGE: List + show + filters (view-only resource)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

EXPORT_ID=""
BATCH_ID=""
BATCH_FILE_ID=""
INTEGRATION_EXPORT_ID=""
SKIP_ID_TESTS=0

describe "Resource: integration-invoices-batch-exports"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List integration invoices batch exports"
xbe_json view integration-invoices-batch-exports list --limit 5
assert_success

test_name "List integration invoices batch exports returns array"
xbe_json view integration-invoices-batch-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list integration invoices batch exports"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample integration invoices batch export"
xbe_json view integration-invoices-batch-exports list --limit 1
if [[ $status -eq 0 ]]; then
    EXPORT_ID=$(json_get ".[0].id")
    BATCH_ID=$(json_get ".[0].organization_invoices_batch_id")
    BATCH_FILE_ID=$(json_get ".[0].organization_invoices_batch_file_id")
    INTEGRATION_EXPORT_ID=$(json_get ".[0].integration_export_id")
    if [[ -n "$EXPORT_ID" && "$EXPORT_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_TESTS=1
        skip "No integration invoices batch exports available"
    fi
else
    SKIP_ID_TESTS=1
    fail "Failed to list integration invoices batch exports"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show integration invoices batch export"
if [[ $SKIP_ID_TESTS -eq 0 && -n "$EXPORT_ID" && "$EXPORT_ID" != "null" ]]; then
    xbe_json view integration-invoices-batch-exports show "$EXPORT_ID"
    assert_success
else
    skip "No integration invoices batch export ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List integration invoices batch exports with --organization-invoices-batch filter"
if [[ -n "$BATCH_ID" && "$BATCH_ID" != "null" ]]; then
    xbe_json view integration-invoices-batch-exports list --organization-invoices-batch "$BATCH_ID" --limit 5
    assert_success
else
    skip "No organization invoices batch ID available"
fi

test_name "List integration invoices batch exports with --organization-invoices-batch-file filter"
if [[ -n "$BATCH_FILE_ID" && "$BATCH_FILE_ID" != "null" ]]; then
    xbe_json view integration-invoices-batch-exports list --organization-invoices-batch-file "$BATCH_FILE_ID" --limit 5
    assert_success
else
    skip "No organization invoices batch file ID available"
fi

test_name "List integration invoices batch exports with --integration-export filter"
if [[ -n "$INTEGRATION_EXPORT_ID" && "$INTEGRATION_EXPORT_ID" != "null" ]]; then
    xbe_json view integration-invoices-batch-exports list --integration-export "$INTEGRATION_EXPORT_ID" --limit 5
    assert_success
else
    skip "No integration export ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
