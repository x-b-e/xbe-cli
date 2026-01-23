#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Invoices
#
# Tests list filters and show operations for the time_card_invoices resource.
#
# COVERAGE: List filters + show
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

TIME_CARD_INVOICE_ID=""
INVOICE_ID=""
TIME_CARD_ID=""
INVOICE_STATUS=""
INVOICE_TYPE=""
SELLER_ID=""
SKIP_ID_FILTERS=0

describe "Resource: time-card-invoices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card invoices"
xbe_json view time-card-invoices list --limit 5
assert_success

test_name "List time card invoices returns array"
xbe_json view time-card-invoices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time card invoices"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample time card invoice"
xbe_json view time-card-invoices list --limit 1
if [[ $status -eq 0 ]]; then
    TIME_CARD_INVOICE_ID=$(json_get ".[0].id")
    INVOICE_ID=$(json_get ".[0].invoice_id")
    TIME_CARD_ID=$(json_get ".[0].time_card_id")
    INVOICE_STATUS=$(json_get ".[0].invoice_status")
    INVOICE_TYPE=$(json_get ".[0].invoice_type")
    SELLER_ID=$(json_get ".[0].seller_id")
    if [[ -n "$TIME_CARD_INVOICE_ID" && "$TIME_CARD_INVOICE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No time card invoices available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list time card invoices"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time card invoice"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TIME_CARD_INVOICE_ID" && "$TIME_CARD_INVOICE_ID" != "null" ]]; then
    xbe_json view time-card-invoices show "$TIME_CARD_INVOICE_ID"
    assert_success
else
    skip "No time card invoice ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List time card invoices with --shift-starts-after filter"
xbe_json view time-card-invoices list --shift-starts-after 2000-01-01T00:00:00Z --limit 5
assert_success

test_name "List time card invoices with --invoice-status filter"
if [[ -n "$INVOICE_STATUS" && "$INVOICE_STATUS" != "null" ]]; then
    xbe_json view time-card-invoices list --invoice-status "$INVOICE_STATUS" --limit 5
    assert_success
else
    skip "No invoice status available"
fi

test_name "List time card invoices with --invoice-type filter"
if [[ -n "$INVOICE_TYPE" && "$INVOICE_TYPE" != "null" ]]; then
    xbe_json view time-card-invoices list --invoice-type "$INVOICE_TYPE" --limit 5
    assert_success
else
    skip "No invoice type available"
fi

test_name "List time card invoices with --seller filter"
if [[ -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    xbe_json view time-card-invoices list --seller "$SELLER_ID" --limit 5
    assert_success
else
    skip "No seller ID available"
fi

run_tests
