#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batches
#
# Tests create, show, and list filtering for organization invoices batches.
#
# COVERAGE: Create relationships + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BATCH_ID=""
LIST_BATCH_ID=""
LIST_STATUS=""
LIST_ORG_TYPE=""
LIST_ORG_ID=""
LIST_BROKER_ID=""
LIST_CREATED_BY_ID=""
LIST_UPDATED_BY_ID=""
LIST_CREATED_AT=""

ENV_ORGANIZATION="${XBE_TEST_ORGANIZATION:-}"
ENV_INVOICE_IDS="${XBE_TEST_ORG_INVOICES_BATCH_INVOICE_IDS:-${XBE_TEST_INVOICE_IDS:-}}"

INVOICE_ID_FOR_FILTER=""

if [[ -n "$ENV_INVOICE_IDS" ]]; then
    INVOICE_ID_FOR_FILTER=$(echo "$ENV_INVOICE_IDS" | cut -d',' -f1 | tr -d ' ')
fi

describe "Resource: organization_invoices_batches"

# ============================================================================
# LIST Tests (smoke)
# ============================================================================

test_name "List organization invoices batches"
xbe_json view organization-invoices-batches list --limit 5
if [[ $status -eq 0 ]]; then
    LIST_BATCH_ID=$(json_get ".[0].id")
    LIST_STATUS=$(json_get ".[0].status")
    LIST_ORG_TYPE=$(json_get ".[0].organization_type")
    LIST_ORG_ID=$(json_get ".[0].organization_id")
    LIST_BROKER_ID=$(json_get ".[0].broker_id")
    LIST_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    LIST_UPDATED_BY_ID=$(json_get ".[0].updated_by_id")
    if [[ -n "$LIST_BATCH_ID" && "$LIST_BATCH_ID" != "null" ]]; then
        pass
    else
        skip "No organization invoices batches returned"
    fi
else
    fail "Failed to list organization invoices batches"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create organization invoices batch without required flags fails"
xbe_run do organization-invoices-batches create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create organization invoices batch"
if [[ -n "$ENV_ORGANIZATION" && -n "$ENV_INVOICE_IDS" ]]; then
    xbe_json do organization-invoices-batches create \
        --organization "$ENV_ORGANIZATION" \
        --invoices "$ENV_INVOICE_IDS"

    if [[ $status -eq 0 ]]; then
        CREATED_BATCH_ID=$(json_get ".id")
        if [[ -n "$CREATED_BATCH_ID" && "$CREATED_BATCH_ID" != "null" ]]; then
            pass
        else
            fail "Created organization invoices batch but no ID returned"
        fi
    else
        fail "Failed to create organization invoices batch"
    fi
else
    skip "Set XBE_TEST_ORGANIZATION and XBE_TEST_ORG_INVOICES_BATCH_INVOICE_IDS (or XBE_TEST_INVOICE_IDS) to run"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch details"
SHOW_ID="$CREATED_BATCH_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$LIST_BATCH_ID"
fi

if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view organization-invoices-batches show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        actual_id=$(json_get ".id")
        LIST_CREATED_AT=$(json_get ".created_at")
        if [[ "$actual_id" == "$SHOW_ID" ]]; then
            pass
        else
            fail "Expected show id $SHOW_ID, got $actual_id"
        fi
    else
        fail "Failed to show organization invoices batch"
    fi
else
    skip "No organization invoices batch ID available"
fi

# ============================================================================
# LIST Filter Tests
# ============================================================================

test_name "List organization invoices batches filtered by processed status"
FILTER_ID="$CREATED_BATCH_ID"
if [[ -z "$FILTER_ID" || "$FILTER_ID" == "null" ]]; then
    FILTER_ID="$LIST_BATCH_ID"
fi

if [[ -n "$FILTER_ID" && "$FILTER_ID" != "null" && -n "$LIST_STATUS" && "$LIST_STATUS" != "null" ]]; then
    processed_value=""
    if [[ "$LIST_STATUS" == "processed" ]]; then
        processed_value="true"
    elif [[ "$LIST_STATUS" == "not_processed" ]]; then
        processed_value="false"
    fi

    if [[ -n "$processed_value" ]]; then
        xbe_json view organization-invoices-batches list --processed "$processed_value"
        if [[ $status -eq 0 ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in processed filter results"
            fi
        else
            fail "Failed to list organization invoices batches with processed filter"
        fi
    else
        skip "Unknown processed status value: $LIST_STATUS"
    fi
else
    skip "No batch status available for processed filter test"
fi


test_name "List organization invoices batches filtered by broker"
FILTER_BROKER_ID="$LIST_BROKER_ID"
FILTER_ID="$CREATED_BATCH_ID"
if [[ -z "$FILTER_ID" || "$FILTER_ID" == "null" ]]; then
    FILTER_ID="$LIST_BATCH_ID"
fi

if [[ -n "$FILTER_BROKER_ID" && "$FILTER_BROKER_ID" != "null" ]]; then
    xbe_json view organization-invoices-batches list --broker "$FILTER_BROKER_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_ID" && "$FILTER_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in broker filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batches with broker filter"
    fi
else
    skip "No broker ID available for filter test"
fi


test_name "List organization invoices batches filtered by created-by"
FILTER_CREATED_BY_ID="$LIST_CREATED_BY_ID"
FILTER_ID="$CREATED_BATCH_ID"
if [[ -z "$FILTER_ID" || "$FILTER_ID" == "null" ]]; then
    FILTER_ID="$LIST_BATCH_ID"
fi

if [[ -n "$FILTER_CREATED_BY_ID" && "$FILTER_CREATED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batches list --created-by "$FILTER_CREATED_BY_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_ID" && "$FILTER_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in created-by filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batches with created-by filter"
    fi
else
    skip "No created-by ID available for filter test"
fi


test_name "List organization invoices batches filtered by changed-by"
FILTER_CHANGED_BY_ID="$LIST_UPDATED_BY_ID"
FILTER_ID="$CREATED_BATCH_ID"
if [[ -z "$FILTER_ID" || "$FILTER_ID" == "null" ]]; then
    FILTER_ID="$LIST_BATCH_ID"
fi

if [[ -n "$FILTER_CHANGED_BY_ID" && "$FILTER_CHANGED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batches list --changed-by "$FILTER_CHANGED_BY_ID"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$FILTER_ID" && "$FILTER_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$FILTER_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in changed-by filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batches with changed-by filter"
    fi
else
    skip "No changed-by ID available for filter test"
fi


test_name "List organization invoices batches filtered by organization"
if [[ -n "$ENV_ORGANIZATION" ]]; then
    xbe_json view organization-invoices-batches list --organization "$ENV_ORGANIZATION"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$CREATED_BATCH_ID" && "$CREATED_BATCH_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$CREATED_BATCH_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in organization filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batches with organization filter"
    fi
else
    skip "Set XBE_TEST_ORGANIZATION to run"
fi


test_name "List organization invoices batches filtered by invoices"
if [[ -n "$CREATED_BATCH_ID" && "$CREATED_BATCH_ID" != "null" && -n "$INVOICE_ID_FOR_FILTER" ]]; then
    xbe_json view organization-invoices-batches list --invoices "$INVOICE_ID_FOR_FILTER"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$CREATED_BATCH_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
            pass
        else
            fail "Expected batch in invoices filter results"
        fi
    else
        fail "Failed to list organization invoices batches with invoices filter"
    fi
else
    skip "Create a batch and set invoice IDs to run"
fi


test_name "List organization invoices batches filtered by invoices-id"
if [[ -n "$CREATED_BATCH_ID" && "$CREATED_BATCH_ID" != "null" && -n "$INVOICE_ID_FOR_FILTER" ]]; then
    xbe_json view organization-invoices-batches list --invoices-id "$INVOICE_ID_FOR_FILTER"
    if [[ $status -eq 0 ]]; then
        if echo "$output" | jq -e --arg id "$CREATED_BATCH_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
            pass
        else
            fail "Expected batch in invoices-id filter results"
        fi
    else
        fail "Failed to list organization invoices batches with invoices-id filter"
    fi
else
    skip "Create a batch and set invoice IDs to run"
fi


test_name "List organization invoices batches filtered by created-at-min"
if [[ -n "$LIST_CREATED_AT" && "$LIST_CREATED_AT" != "null" ]]; then
    xbe_json view organization-invoices-batches list --created-at-min "$LIST_CREATED_AT"
    if [[ $status -eq 0 ]]; then
        if [[ -n "$LIST_BATCH_ID" && "$LIST_BATCH_ID" != "null" ]]; then
            if echo "$output" | jq -e --arg id "$LIST_BATCH_ID" 'map(select(.id == $id)) | length > 0' > /dev/null; then
                pass
            else
                fail "Expected batch in created-at-min filter results"
            fi
        else
            pass
        fi
    else
        fail "Failed to list organization invoices batches with created-at-min filter"
    fi
else
    skip "No created-at timestamp available for filter test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
