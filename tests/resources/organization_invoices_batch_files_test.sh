#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Files
#
# Tests create, show, and list filtering for organization invoices batch files.
#
# COVERAGE: Create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BATCH_FILE_ID=""
LIST_BATCH_FILE_ID=""
LIST_BATCH_ID=""

ENV_BATCH_ID="${XBE_TEST_ORG_INVOICES_BATCH_ID:-}"
ENV_FORMATTER_ID="${XBE_TEST_ORGANIZATION_FORMATTER_ID:-}"
ENV_ORGANIZATION="${XBE_TEST_ORGANIZATION:-}"

describe "Resource: organization_invoices_batch_files"

# ============================================================================
# LIST Tests (smoke)
# ============================================================================

test_name "List organization invoices batch files"
xbe_json view organization-invoices-batch-files list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_BATCH_FILE_ID=$(json_get ".[0].id")
    LIST_BATCH_ID=$(json_get ".[0].organization_invoices_batch_id")
    if [[ -n "$LIST_BATCH_FILE_ID" && "$LIST_BATCH_FILE_ID" != "null" ]]; then
        pass
    else
        skip "No organization invoices batch files returned"
    fi
else
    fail "Failed to list organization invoices batch files"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization invoices batch file without required flags fails"
xbe_run do organization-invoices-batch-files create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization invoices batch file"
if [[ -n "$ENV_BATCH_ID" && -n "$ENV_FORMATTER_ID" ]]; then
    if [[ -n "$ENV_ORGANIZATION" ]]; then
        xbe_json do organization-invoices-batch-files create \
            --organization-invoices-batch "$ENV_BATCH_ID" \
            --organization-formatter "$ENV_FORMATTER_ID" \
            --organization "$ENV_ORGANIZATION" \
            --body "CLI batch file" \
            --mime-type "text/plain" \
            --refresh-invoice-revisions false
    else
        xbe_json do organization-invoices-batch-files create \
            --organization-invoices-batch "$ENV_BATCH_ID" \
            --organization-formatter "$ENV_FORMATTER_ID" \
            --body "CLI batch file" \
            --mime-type "text/plain" \
            --refresh-invoice-revisions false
    fi

    if [[ $status -eq 0 ]]; then
        CREATED_BATCH_FILE_ID=$(json_get ".id")
        if [[ -n "$CREATED_BATCH_FILE_ID" && "$CREATED_BATCH_FILE_ID" != "null" ]]; then
            pass
        else
            fail "Created organization invoices batch file but no ID returned"
        fi
    else
        fail "Failed to create organization invoices batch file"
    fi
else
    skip "Set XBE_TEST_ORG_INVOICES_BATCH_ID and XBE_TEST_ORGANIZATION_FORMATTER_ID to run"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch file details"
SHOW_ID="$CREATED_BATCH_FILE_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$LIST_BATCH_FILE_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-files show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        actual_id=$(json_get ".id")
        if [[ "$actual_id" == "$SHOW_ID" ]]; then
            pass
        else
            fail "Expected show id $SHOW_ID, got $actual_id"
        fi
    else
        fail "Failed to show organization invoices batch file"
    fi
else
    skip "No organization invoices batch file ID available"
fi

# ============================================================================
# LIST Filter Tests
# ============================================================================

test_name "List organization invoices batch files filtered by batch"
FILTER_BATCH_ID="$ENV_BATCH_ID"
FILTER_FILE_ID="$CREATED_BATCH_FILE_ID"
if [[ -z "$FILTER_BATCH_ID" || "$FILTER_BATCH_ID" == "null" ]]; then
    FILTER_BATCH_ID="$LIST_BATCH_ID"
    FILTER_FILE_ID="$LIST_BATCH_FILE_ID"
fi

if [[ -n "$FILTER_BATCH_ID" && "$FILTER_BATCH_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-files list --organization-invoices-batch "$FILTER_BATCH_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_FILE_ID" && "$FILTER_FILE_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_FILE_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch file in filtered results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batch files"
    fi
else
    skip "No batch ID available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
