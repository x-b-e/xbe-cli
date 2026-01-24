#!/bin/bash
#
# XBE CLI Integration Tests: Organization Invoices Batch Invoices
#
# Tests list/show operations for organization-invoices-batch-invoices.
#
# COVERAGE: List filters (organization-invoices-batch, invoice, invoice-id, organization, organization-type/id, created-by, changed-by, successful) + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BATCH_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_ORGANIZATION_INVOICES_BATCH_ID:-}"
INVOICE_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_INVOICE_ID:-}"
INVOICE_TYPE="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_INVOICE_TYPE:-}"
ORGANIZATION_TYPE="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_ORGANIZATION_TYPE:-}"
ORGANIZATION_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_ORGANIZATION_ID:-}"
CREATED_BY_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_CREATED_BY_ID:-}"
CHANGED_BY_ID="${XBE_TEST_ORGANIZATION_INVOICES_BATCH_INVOICE_CHANGED_BY_ID:-}"
SAMPLE_ID=""

normalize_org_type() {
    case "$1" in
        brokers|Broker|BROKER) echo "Broker" ;;
        customers|Customer|CUSTOMER) echo "Customer" ;;
        truckers|Trucker|TRUCKER) echo "Trucker" ;;
        *) echo "$1" ;;
    esac
}

describe "Resource: organization-invoices-batch-invoices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List organization invoices batch invoices"
xbe_json view organization-invoices-batch-invoices list --limit 5
assert_success

test_name "List organization invoices batch invoices returns array"
xbe_json view organization-invoices-batch-invoices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list organization invoices batch invoices"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample organization invoices batch invoice"
xbe_json view organization-invoices-batch-invoices list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    if [[ -z "$BATCH_ID" || "$BATCH_ID" == "null" ]]; then
        BATCH_ID=$(json_get ".[0].organization_invoices_batch_id")
    fi
    if [[ -z "$INVOICE_ID" || "$INVOICE_ID" == "null" ]]; then
        INVOICE_ID=$(json_get ".[0].invoice_id")
    fi
    if [[ -z "$INVOICE_TYPE" || "$INVOICE_TYPE" == "null" ]]; then
        INVOICE_TYPE=$(json_get ".[0].invoice_type")
    fi
    if [[ -z "$ORGANIZATION_TYPE" || "$ORGANIZATION_TYPE" == "null" ]]; then
        ORGANIZATION_TYPE=$(json_get ".[0].organization_type")
    fi
    if [[ -z "$ORGANIZATION_ID" || "$ORGANIZATION_ID" == "null" ]]; then
        ORGANIZATION_ID=$(json_get ".[0].organization_id")
    fi
    if [[ -z "$CREATED_BY_ID" || "$CREATED_BY_ID" == "null" ]]; then
        CREATED_BY_ID=$(json_get ".[0].created_by_id")
    fi
    if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" ]]; then
        ORGANIZATION_TYPE=$(normalize_org_type "$ORGANIZATION_TYPE")
    fi
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No organization invoices batch invoices available for show test"
    fi
else
    skip "Could not list organization invoices batch invoices to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List organization invoices batch invoices with --organization-invoices-batch filter"
if [[ -n "$BATCH_ID" && "$BATCH_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --organization-invoices-batch "$BATCH_ID" --limit 5
    assert_success
else
    skip "No batch ID available for --organization-invoices-batch filter"
fi

test_name "List organization invoices batch invoices with --invoice-id filter"
if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --invoice-id "$INVOICE_ID" --limit 5
    assert_success
else
    skip "No invoice ID available for --invoice-id filter"
fi

test_name "List organization invoices batch invoices with --invoice filter"
if [[ -n "$INVOICE_TYPE" && "$INVOICE_TYPE" != "null" && -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --invoice "${INVOICE_TYPE}|${INVOICE_ID}" --limit 5
    assert_success
else
    skip "No invoice type/ID available for --invoice filter"
fi

test_name "List organization invoices batch invoices with --organization filter"
if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --organization "${ORGANIZATION_TYPE}|${ORGANIZATION_ID}" --limit 5
    assert_success
else
    skip "No organization available for --organization filter"
fi

test_name "List organization invoices batch invoices with --organization-type and --organization-id filters"
if [[ -n "$ORGANIZATION_TYPE" && "$ORGANIZATION_TYPE" != "null" && -n "$ORGANIZATION_ID" && "$ORGANIZATION_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --organization-type "$ORGANIZATION_TYPE" --organization-id "$ORGANIZATION_ID" --limit 5
    assert_success
else
    skip "No organization available for organization-type/id filters"
fi

test_name "List organization invoices batch invoices with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for --created-by filter"
fi

test_name "List organization invoices batch invoices with --changed-by filter"
if [[ -n "$CHANGED_BY_ID" && "$CHANGED_BY_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices list --changed-by "$CHANGED_BY_ID" --limit 5
    assert_success
else
    skip "No changed-by ID available for --changed-by filter"
fi

test_name "List organization invoices batch invoices with --successful filter"
xbe_json view organization-invoices-batch-invoices list --successful true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show organization invoices batch invoice"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view organization-invoices-batch-invoices show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Not Found"* ]] || [[ "$output" == *"not found"* ]]; then
            skip "Show blocked by server policy or record not found"
        else
            fail "Failed to show organization invoices batch invoice: $output"
        fi
    fi
else
    skip "No organization invoices batch invoice ID available for show"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
