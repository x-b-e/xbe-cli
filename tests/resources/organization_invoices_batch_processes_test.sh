#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Processes
#
# Tests view and create operations for organization_invoices_batch_processes.
# Processes transition batches from not processed to processed.
#
# COVERAGE: Create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_ID=""
SECOND_BATCH_ID=""
SAMPLE_PROCESS_ID=""
SKIP_MUTATION=0

describe "Resource: organization-invoices-batch-processes"

# ============================================================================
# LIST / SHOW Tests
# ============================================================================

test_name "List organization invoices batch processes"
xbe_json view organization-invoices-batch-processes list --limit 1
assert_success

test_name "Capture sample organization invoices batch process (if available)"
xbe_json view organization-invoices-batch-processes list --limit 1
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_PROCESS_ID=$(json_get ".[0].id")
        pass
    else
        echo "    No organization invoices batch processes available; skipping show test."
        pass
    fi
else
    fail "Failed to list organization invoices batch processes"
fi

if [[ -n "$SAMPLE_PROCESS_ID" && "$SAMPLE_PROCESS_ID" != "null" ]]; then
    test_name "Show organization invoices batch process"
    xbe_json view organization-invoices-batch-processes show "$SAMPLE_PROCESS_ID"
    assert_success
fi

# ============================================================================
# CREATE Error Tests
# ============================================================================

test_name "Process organization invoices batch requires --organization-invoices-batch"
xbe_run do organization-invoices-batch-processes create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

if [[ -z "$XBE_TOKEN" ]]; then
    echo "    (XBE_TOKEN not set; skipping mutation tests)"
    SKIP_MUTATION=1
fi

if [[ $SKIP_MUTATION -eq 0 ]]; then
    test_name "Find not processed organization invoices batch"
    response_file=$(mktemp)
    run curl -s -o "$response_file" -w "%{http_code}" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Accept: application/vnd.api+json" \
        -G "$XBE_BASE_URL/v1/organization-invoices-batches" \
        --data-urlencode "page[limit]=50" \
        --data-urlencode "filter[processed]=false" \
        --data-urlencode "fields[organization-invoices-batches]=status"

    http_code="$output"
    if [[ $status -eq 0 && "$http_code" == 2* ]]; then
        mapfile -t candidate_ids < <(jq -r '.data[] | select(.attributes.status=="not_processed") | .id' "$response_file" 2>/dev/null)
        BATCH_ID="${candidate_ids[0]}"
        SECOND_BATCH_ID="${candidate_ids[1]}"
        if [[ -n "$BATCH_ID" && "$BATCH_ID" != "null" ]]; then
            pass
        else
            skip "No not processed organization invoices batch found"
        fi
    else
        skip "Unable to list organization invoices batches (HTTP ${http_code})"
    fi
    rm -f "$response_file"
fi

if [[ -n "$BATCH_ID" && "$BATCH_ID" != "null" ]]; then
    if [[ -n "$SECOND_BATCH_ID" && "$SECOND_BATCH_ID" != "null" ]]; then
        test_name "Process organization invoices batch (minimal)"
        xbe_json do organization-invoices-batch-processes create --organization-invoices-batch "$BATCH_ID"
        assert_success

        test_name "Process organization invoices batch with comment"
        COMMENT_TEXT="$(unique_name "OrgInvoicesBatchProcess")"
        xbe_json do organization-invoices-batch-processes create \
            --organization-invoices-batch "$SECOND_BATCH_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to process organization invoices batch with comment"
        fi
    else
        test_name "Process organization invoices batch with comment"
        COMMENT_TEXT="$(unique_name "OrgInvoicesBatchProcess")"
        xbe_json do organization-invoices-batch-processes create \
            --organization-invoices-batch "$BATCH_ID" \
            --comment "$COMMENT_TEXT"

        if [[ $status -eq 0 ]]; then
            assert_json_equals ".comment" "$COMMENT_TEXT"
        else
            fail "Failed to process organization invoices batch"
        fi
    fi
else
    test_name "Process organization invoices batch"
    skip "No not processed organization invoices batch available for processing"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Process organization invoices batch with invalid ID fails"
xbe_run do organization-invoices-batch-processes create --organization-invoices-batch "999999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
