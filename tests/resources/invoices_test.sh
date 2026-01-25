#!/bin/bash
#
# XBE CLI Integration Tests: Invoices
#
# Tests list filters and show operations for the invoices resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

INVOICE_ID=""
STATUS=""
INVOICE_DATE=""
DUE_ON=""
BUYER_ID=""
BUYER_TYPE=""
SELLER_ID=""
SELLER_TYPE=""
BUSINESS_UNIT_ID=""
CUSTOMER_ID=""
BROKER_ID=""
IS_MANAGEMENT_SERVICE_TYPE=""
BATCH_ORG_TYPE=""
BATCH_ORG_ID=""
SKIP_ID_FILTERS=0

describe "Resource: invoices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List invoices"
xbe_json view invoices list --limit 5
assert_success

test_name "List invoices returns array"
xbe_json view invoices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list invoices"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample invoice"
xbe_json view invoices list --limit 1
if [[ $status -eq 0 ]]; then
    INVOICE_ID=$(json_get ".[0].id")
    STATUS=$(json_get ".[0].status")
    INVOICE_DATE=$(json_get ".[0].invoice_date")
    DUE_ON=$(json_get ".[0].due_on")
    BUYER_ID=$(json_get ".[0].buyer_id")
    BUYER_TYPE=$(json_get ".[0].buyer_type")
    SELLER_ID=$(json_get ".[0].seller_id")
    SELLER_TYPE=$(json_get ".[0].seller_type")

    if [[ -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No invoices available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list invoices"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show invoice"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$INVOICE_ID" && "$INVOICE_ID" != "null" ]]; then
    xbe_json view invoices show "$INVOICE_ID"
    if [[ $status -eq 0 ]]; then
        BUSINESS_UNIT_ID=$(json_get ".business_unit_ids[0]")
        CUSTOMER_ID=$(json_get ".customer_ids[0]")
        IS_MANAGEMENT_SERVICE_TYPE=$(json_get ".is_management_service_type")
        pass
    else
        fail "Failed to show invoice"
    fi
else
    skip "No invoice ID available"
fi

if [[ "$BUYER_TYPE" == "brokers" && -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
    BROKER_ID="$BUYER_ID"
elif [[ "$SELLER_TYPE" == "brokers" && -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    BROKER_ID="$SELLER_ID"
fi

if [[ -z "$CUSTOMER_ID" || "$CUSTOMER_ID" == "null" ]]; then
    if [[ "$BUYER_TYPE" == "customers" && -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
        CUSTOMER_ID="$BUYER_ID"
    fi
fi

if [[ "$BUYER_TYPE" =~ ^(customers|brokers|truckers)$ && -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
    BATCH_ORG_TYPE="$BUYER_TYPE"
    BATCH_ORG_ID="$BUYER_ID"
elif [[ "$SELLER_TYPE" =~ ^(customers|brokers|truckers)$ && -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    BATCH_ORG_TYPE="$SELLER_TYPE"
    BATCH_ORG_ID="$SELLER_ID"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List invoices with --status filter"
if [[ -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view invoices list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List invoices with --invoice-date filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view invoices list --invoice-date "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List invoices with --invoice-date-min filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view invoices list --invoice-date-min "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List invoices with --invoice-date-max filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view invoices list --invoice-date-max "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List invoices with --has-invoice-date filter"
xbe_json view invoices list --has-invoice-date true --limit 5
assert_success

test_name "List invoices with --due-on filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view invoices list --due-on "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List invoices with --due-on-min filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view invoices list --due-on-min "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List invoices with --due-on-max filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view invoices list --due-on-max "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List invoices with --has-due-on filter"
xbe_json view invoices list --has-due-on true --limit 5
assert_success

test_name "List invoices with --buyer filter"
if [[ -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
    xbe_json view invoices list --buyer "$BUYER_ID" --limit 5
    assert_success
else
    skip "No buyer ID available"
fi

test_name "List invoices with --seller filter"
if [[ -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    xbe_json view invoices list --seller "$SELLER_ID" --limit 5
    assert_success
else
    skip "No seller ID available"
fi

test_name "List invoices with --ticket-number filter"
xbe_json view invoices list --ticket-number "TEST-123" --limit 5
assert_success

test_name "List invoices with --material-transaction-ticket-numbers filter"
xbe_json view invoices list --material-transaction-ticket-numbers "MT-123" --limit 5
assert_success

test_name "List invoices with --tender filter"
xbe_json view invoices list --tender "1" --limit 5
assert_success

test_name "List invoices with --is-management-service-type filter"
if [[ -n "$IS_MANAGEMENT_SERVICE_TYPE" && "$IS_MANAGEMENT_SERVICE_TYPE" != "null" ]]; then
    xbe_json view invoices list --is-management-service-type "$IS_MANAGEMENT_SERVICE_TYPE" --limit 5
    assert_success
else
    xbe_json view invoices list --is-management-service-type false --limit 5
    assert_success
fi

test_name "List invoices with --business-unit filter"
if [[ -n "$BUSINESS_UNIT_ID" && "$BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view invoices list --business-unit "$BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List invoices with --not-business-unit filter"
if [[ -n "$BUSINESS_UNIT_ID" && "$BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view invoices list --not-business-unit "$BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List invoices with --customer filter"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
    xbe_json view invoices list --customer "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "List invoices with --material-transaction-cost-codes filter"
xbe_json view invoices list --material-transaction-cost-codes "123.45" --limit 5
assert_success

test_name "List invoices with --allocated-cost-codes filter"
xbe_json view invoices list --allocated-cost-codes "123-456" --limit 5
assert_success

test_name "List invoices with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view invoices list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List invoices with --has-ticket-report filter"
xbe_json view invoices list --has-ticket-report true --limit 5
assert_success

test_name "List invoices with --batch-status filter"
if [[ -n "$BATCH_ORG_TYPE" && -n "$BATCH_ORG_ID" && "$BATCH_ORG_ID" != "null" ]]; then
    xbe_json view invoices list --batch-status "${BATCH_ORG_TYPE}|${BATCH_ORG_ID}|never_processed" --limit 5
    assert_success
else
    skip "No batch-status org info available"
fi

test_name "List invoices with --having-plans-with-job-number-like filter"
xbe_json view invoices list --having-plans-with-job-number-like "JOB" --limit 5
assert_success

run_tests
