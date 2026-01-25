#!/bin/bash
#
# XBE CLI Integration Tests: Broker Invoices
#
# Tests list/show/create/update operations for broker-invoices.
#
# COVERAGE: List filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_STATUS=""
SAMPLE_INVOICE_DATE=""
SAMPLE_DUE_ON=""
SAMPLE_BUYER_TYPE=""
SAMPLE_BUYER_ID=""
SAMPLE_SELLER_TYPE=""
SAMPLE_SELLER_ID=""
SAMPLE_BUSINESS_UNIT_ID=""
SAMPLE_CUSTOMER_ID=""
SAMPLE_BROKER_ID=""
SAMPLE_TIME_CARD_ID=""

CREATED_ID=""

BROKER_INVOICE_ID="${XBE_TEST_BROKER_INVOICE_ID:-}"
TIME_CARD_ID="${XBE_TEST_TIME_CARD_ID:-}"
BROKER_ID="${XBE_TEST_TIME_CARD_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_TIME_CARD_CUSTOMER_ID:-}"
TICKET_NUMBER=""

fetch_time_card_details() {
    local tc_id="$1"
    if [[ -z "$tc_id" || "$tc_id" == "null" ]]; then
        return
    fi
    xbe_json view time-cards show "$tc_id"
    if [[ $status -ne 0 ]]; then
        return
    fi
    if [[ -z "$BROKER_ID" || "$BROKER_ID" == "null" ]]; then
        BROKER_ID=$(json_get ".broker_id")
    fi
    if [[ -z "$CUSTOMER_ID" || "$CUSTOMER_ID" == "null" ]]; then
        CUSTOMER_ID=$(json_get ".customer_id")
    fi
    if [[ -z "$TICKET_NUMBER" || "$TICKET_NUMBER" == "null" ]]; then
        TICKET_NUMBER=$(json_get ".ticket_number")
    fi
}

describe "Resource: broker-invoices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List broker invoices"
xbe_json view broker-invoices list --limit 5
assert_success

test_name "List broker invoices returns array"
xbe_json view broker-invoices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_STATUS=$(json_get ".[0].status")
    SAMPLE_INVOICE_DATE=$(json_get ".[0].invoice_date")
    SAMPLE_DUE_ON=$(json_get ".[0].due_on")
    SAMPLE_BUYER_TYPE=$(json_get ".[0].buyer_type")
    SAMPLE_BUYER_ID=$(json_get ".[0].buyer_id")
    SAMPLE_SELLER_TYPE=$(json_get ".[0].seller_type")
    SAMPLE_SELLER_ID=$(json_get ".[0].seller_id")
else
    fail "Failed to list broker invoices"
fi

# ============================================================================
# SHOW Tests (capture additional sample data)
# ============================================================================

test_name "Show broker invoice"
DETAIL_ID="${BROKER_INVOICE_ID:-$SAMPLE_ID}"
if [[ -n "$DETAIL_ID" && "$DETAIL_ID" != "null" ]]; then
    xbe_json view broker-invoices show "$DETAIL_ID"
    if [[ $status -eq 0 ]]; then
        SAMPLE_BUSINESS_UNIT_ID=$(json_get ".business_unit_ids[0]")
        SAMPLE_CUSTOMER_ID=$(json_get ".customer_ids[0]")
        SAMPLE_TIME_CARD_ID=$(json_get ".time_card_ids[0]")
        if [[ -z "$SAMPLE_BUYER_ID" || "$SAMPLE_BUYER_ID" == "null" ]]; then
            SAMPLE_BUYER_ID=$(json_get ".buyer_id")
            SAMPLE_BUYER_TYPE=$(json_get ".buyer_type")
        fi
        if [[ -z "$SAMPLE_SELLER_ID" || "$SAMPLE_SELLER_ID" == "null" ]]; then
            SAMPLE_SELLER_ID=$(json_get ".seller_id")
            SAMPLE_SELLER_TYPE=$(json_get ".seller_type")
        fi
        pass
    else
        fail "Failed to show broker invoice"
    fi
else
    skip "No broker invoice ID available for show tests"
fi

if [[ -z "$TIME_CARD_ID" || "$TIME_CARD_ID" == "null" ]]; then
    TIME_CARD_ID="$SAMPLE_TIME_CARD_ID"
fi

fetch_time_card_details "$TIME_CARD_ID"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create broker invoice"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" && -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    INVOICE_DATE=$(date -u +%Y-%m-%d)
    DUE_DATE=$(date -u -v+30d +%Y-%m-%d 2>/dev/null || date -u -d '+30 days' +%Y-%m-%d)
    if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
        xbe_json do broker-invoices create \
            --customer "$CUSTOMER_ID" \
            --broker "$BROKER_ID" \
            --time-card-ids "$TIME_CARD_ID" \
            --invoice-date "$INVOICE_DATE" \
            --due-on "$DUE_DATE" \
            --adjustment-amount "0" \
            --currency-code "USD" \
            --notes "CLI broker invoice test"
    else
        xbe_json do broker-invoices create \
            --customer "$CUSTOMER_ID" \
            --broker "$BROKER_ID" \
            --invoice-date "$INVOICE_DATE" \
            --due-on "$DUE_DATE" \
            --adjustment-amount "0" \
            --currency-code "USD" \
            --notes "CLI broker invoice test"
    fi

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "broker-invoices" "$CREATED_ID"
            pass
        else
            fail "Created broker invoice but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create broker invoice: $output"
        fi
    fi
else
    skip "No broker/customer IDs available (set XBE_TEST_TIME_CARD_BROKER_ID and XBE_TEST_TIME_CARD_CUSTOMER_ID or XBE_TEST_TIME_CARD_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

UPDATE_ID="${CREATED_ID:-${BROKER_INVOICE_ID:-$SAMPLE_ID}}"

update_broker_invoice() {
    local description="$1"
    shift
    test_name "$description"
    xbe_json do broker-invoices update "$UPDATE_ID" "$@"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]]; then
            skip "Update blocked by server policy/validation"
        else
            fail "Update failed: $output"
        fi
    fi
}

if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    update_broker_invoice "Update notes" --notes "Updated broker invoice notes"
    update_broker_invoice "Update invoice date" --invoice-date "$(date -u +%Y-%m-%d)"
    update_broker_invoice "Update due on" --due-on "$(date -u -v+15d +%Y-%m-%d 2>/dev/null || date -u -d '+15 days' +%Y-%m-%d)"
    update_broker_invoice "Update adjustment amount" --adjustment-amount "5.00"
    update_broker_invoice "Update currency code" --currency-code "USD"
    update_broker_invoice "Update explicit buyer name" --explicit-buyer-name "Invoice Buyer Override"
    update_broker_invoice "Update explicit buyer address" --explicit-buyer-address "123 Billing St"
else
    skip "No broker invoice ID available for update tests"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List broker invoices with --buyer filter"
if [[ -n "$SAMPLE_BUYER_ID" && "$SAMPLE_BUYER_ID" != "null" ]]; then
    BUYER_TYPE="$SAMPLE_BUYER_TYPE"
    if [[ -z "$BUYER_TYPE" || "$BUYER_TYPE" == "null" ]]; then
        BUYER_TYPE="customers"
    fi
    xbe_json view broker-invoices list --buyer "${BUYER_TYPE}|${SAMPLE_BUYER_ID}" --limit 5
    assert_success
else
    skip "No buyer available for filter"
fi

test_name "List broker invoices with --seller filter"
if [[ -n "$SAMPLE_SELLER_ID" && "$SAMPLE_SELLER_ID" != "null" ]]; then
    SELLER_TYPE="$SAMPLE_SELLER_TYPE"
    if [[ -z "$SELLER_TYPE" || "$SELLER_TYPE" == "null" ]]; then
        SELLER_TYPE="brokers"
    fi
    xbe_json view broker-invoices list --seller "${SELLER_TYPE}|${SAMPLE_SELLER_ID}" --limit 5
    assert_success
else
    skip "No seller available for filter"
fi

test_name "List broker invoices with --status filter"
if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    xbe_json view broker-invoices list --status "$SAMPLE_STATUS" --limit 5
    assert_success
else
    xbe_json view broker-invoices list --status approved --limit 5
    assert_success
fi

INVOICE_DATE_FILTER="$SAMPLE_INVOICE_DATE"
if [[ -z "$INVOICE_DATE_FILTER" || "$INVOICE_DATE_FILTER" == "null" ]]; then
    INVOICE_DATE_FILTER=$(date -u +%Y-%m-%d)
fi

DUE_DATE_FILTER="$SAMPLE_DUE_ON"
if [[ -z "$DUE_DATE_FILTER" || "$DUE_DATE_FILTER" == "null" ]]; then
    DUE_DATE_FILTER=$(date -u +%Y-%m-%d)
fi

test_name "List broker invoices with --invoice-date filter"
xbe_json view broker-invoices list --invoice-date "$INVOICE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --invoice-date-min filter"
xbe_json view broker-invoices list --invoice-date-min "$INVOICE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --invoice-date-max filter"
xbe_json view broker-invoices list --invoice-date-max "$INVOICE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --has-invoice-date=true"
xbe_json view broker-invoices list --has-invoice-date true --limit 5
assert_success

test_name "List broker invoices with --has-invoice-date=false"
xbe_json view broker-invoices list --has-invoice-date false --limit 5
assert_success

test_name "List broker invoices with --due-on filter"
xbe_json view broker-invoices list --due-on "$DUE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --due-on-min filter"
xbe_json view broker-invoices list --due-on-min "$DUE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --due-on-max filter"
xbe_json view broker-invoices list --due-on-max "$DUE_DATE_FILTER" --limit 5
assert_success

test_name "List broker invoices with --has-due-on=true"
xbe_json view broker-invoices list --has-due-on true --limit 5
assert_success

test_name "List broker invoices with --has-due-on=false"
xbe_json view broker-invoices list --has-due-on false --limit 5
assert_success

test_name "List broker invoices with --ticket-number filter"
if [[ -n "$TICKET_NUMBER" && "$TICKET_NUMBER" != "null" ]]; then
    xbe_json view broker-invoices list --ticket-number "$TICKET_NUMBER" --limit 5
    assert_success
else
    xbe_json view broker-invoices list --ticket-number "TEST-TICKET" --limit 5
    assert_success
fi

test_name "List broker invoices with --material-transaction-ticket-numbers filter"
xbe_json view broker-invoices list --material-transaction-ticket-numbers "12345" --limit 5
assert_success

test_name "List broker invoices with --tender filter"
xbe_json view broker-invoices list --tender "1" --limit 5
assert_success

test_name "List broker invoices with --is-management-service-type=true"
xbe_json view broker-invoices list --is-management-service-type true --limit 5
assert_success

test_name "List broker invoices with --is-management-service-type=false"
xbe_json view broker-invoices list --is-management-service-type false --limit 5
assert_success

test_name "List broker invoices with --business-unit filter"
if [[ -n "$SAMPLE_BUSINESS_UNIT_ID" && "$SAMPLE_BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view broker-invoices list --business-unit "$SAMPLE_BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List broker invoices with --not-business-unit filter"
if [[ -n "$SAMPLE_BUSINESS_UNIT_ID" && "$SAMPLE_BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view broker-invoices list --not-business-unit "$SAMPLE_BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List broker invoices with --customer filter"
FILTER_CUSTOMER_ID="$SAMPLE_CUSTOMER_ID"
if [[ -z "$FILTER_CUSTOMER_ID" || "$FILTER_CUSTOMER_ID" == "null" ]]; then
    FILTER_CUSTOMER_ID="$CUSTOMER_ID"
fi
if [[ -z "$FILTER_CUSTOMER_ID" || "$FILTER_CUSTOMER_ID" == "null" ]]; then
    if [[ "$SAMPLE_BUYER_TYPE" == "customers" ]]; then
        FILTER_CUSTOMER_ID="$SAMPLE_BUYER_ID"
    fi
fi
if [[ -n "$FILTER_CUSTOMER_ID" && "$FILTER_CUSTOMER_ID" != "null" ]]; then
    xbe_json view broker-invoices list --customer "$FILTER_CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "List broker invoices with --material-transaction-cost-codes filter"
xbe_json view broker-invoices list --material-transaction-cost-codes "COST123" --limit 5
assert_success

test_name "List broker invoices with --allocated-cost-codes filter"
xbe_json view broker-invoices list --allocated-cost-codes "ALLOC123" --limit 5
assert_success

test_name "List broker invoices with --broker filter"
FILTER_BROKER_ID="$BROKER_ID"
if [[ -z "$FILTER_BROKER_ID" || "$FILTER_BROKER_ID" == "null" ]]; then
    if [[ "$SAMPLE_SELLER_TYPE" == "brokers" ]]; then
        FILTER_BROKER_ID="$SAMPLE_SELLER_ID"
    fi
fi
if [[ -n "$FILTER_BROKER_ID" && "$FILTER_BROKER_ID" != "null" ]]; then
    xbe_json view broker-invoices list --broker "$FILTER_BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List broker invoices with --has-ticket-report=true"
xbe_json view broker-invoices list --has-ticket-report true --limit 5
assert_success

test_name "List broker invoices with --has-ticket-report=false"
xbe_json view broker-invoices list --has-ticket-report false --limit 5
assert_success

test_name "List broker invoices with --batch-status filter"
if [[ -n "$FILTER_CUSTOMER_ID" && "$FILTER_CUSTOMER_ID" != "null" ]]; then
    xbe_json view broker-invoices list --batch-status "customers|${FILTER_CUSTOMER_ID}|never_processed" --limit 5
    assert_success
else
    skip "No customer ID available for batch status filter"
fi

test_name "List broker invoices with --having-plans-with-job-number-like filter"
xbe_json view broker-invoices list --having-plans-with-job-number-like "TEST" --limit 5
assert_success

run_tests
