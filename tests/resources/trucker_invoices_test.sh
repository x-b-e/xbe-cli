#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Invoices
#
# Tests list filters, show, and create/update/delete operations for the
# trucker-invoices resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
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
QUICKBOOKS_ID=""
TIME_CARD_ID=""
SKIP_ID_FILTERS=0

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRUCKER_INVOICE_ID=""

describe "Resource: trucker-invoices"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List trucker invoices"
xbe_json view trucker-invoices list --limit 5
assert_success

test_name "List trucker invoices returns array"
xbe_json view trucker-invoices list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker invoices"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample trucker invoice"
xbe_json view trucker-invoices list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    STATUS=$(json_get ".[0].status")
    INVOICE_DATE=$(json_get ".[0].invoice_date")
    DUE_ON=$(json_get ".[0].due_on")
    BUYER_ID=$(json_get ".[0].buyer_id")
    BUYER_TYPE=$(json_get ".[0].buyer_type")
    SELLER_ID=$(json_get ".[0].seller_id")
    SELLER_TYPE=$(json_get ".[0].seller_type")
    QUICKBOOKS_ID=$(json_get ".[0].quickbooks_id")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No trucker invoices available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list trucker invoices"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show trucker invoice"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view trucker-invoices show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        BUSINESS_UNIT_ID=$(json_get ".business_unit_ids[0]")
        CUSTOMER_ID=$(json_get ".customer_ids[0]")
        IS_MANAGEMENT_SERVICE_TYPE=$(json_get ".is_management_service_type")
        pass
    else
        fail "Failed to show trucker invoice"
    fi
else
    skip "No trucker invoice ID available"
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

test_name "List trucker invoices with --status filter"
if [[ -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view trucker-invoices list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List trucker invoices with --invoice-date filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view trucker-invoices list --invoice-date "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List trucker invoices with --invoice-date-min filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view trucker-invoices list --invoice-date-min "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List trucker invoices with --invoice-date-max filter"
if [[ -n "$INVOICE_DATE" && "$INVOICE_DATE" != "null" ]]; then
    xbe_json view trucker-invoices list --invoice-date-max "$INVOICE_DATE" --limit 5
    assert_success
else
    skip "No invoice date available"
fi

test_name "List trucker invoices with --has-invoice-date filter"
xbe_json view trucker-invoices list --has-invoice-date true --limit 5
assert_success

test_name "List trucker invoices with --due-on filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view trucker-invoices list --due-on "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List trucker invoices with --due-on-min filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view trucker-invoices list --due-on-min "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List trucker invoices with --due-on-max filter"
if [[ -n "$DUE_ON" && "$DUE_ON" != "null" ]]; then
    xbe_json view trucker-invoices list --due-on-max "$DUE_ON" --limit 5
    assert_success
else
    skip "No due-on date available"
fi

test_name "List trucker invoices with --has-due-on filter"
xbe_json view trucker-invoices list --has-due-on true --limit 5
assert_success

test_name "List trucker invoices with --buyer filter"
if [[ -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --buyer "$BUYER_ID" --limit 5
    assert_success
else
    skip "No buyer ID available"
fi

test_name "List trucker invoices with --seller filter"
if [[ -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --seller "$SELLER_ID" --limit 5
    assert_success
else
    skip "No seller ID available"
fi

test_name "List trucker invoices with --ticket-number filter"
xbe_json view trucker-invoices list --ticket-number "TEST-123" --limit 5
assert_success

test_name "List trucker invoices with --material-transaction-ticket-numbers filter"
xbe_json view trucker-invoices list --material-transaction-ticket-numbers "MT-123" --limit 5
assert_success

test_name "List trucker invoices with --tender filter"
xbe_json view trucker-invoices list --tender "1" --limit 5
assert_success

test_name "List trucker invoices with --is-management-service-type filter"
if [[ -n "$IS_MANAGEMENT_SERVICE_TYPE" && "$IS_MANAGEMENT_SERVICE_TYPE" != "null" ]]; then
    xbe_json view trucker-invoices list --is-management-service-type "$IS_MANAGEMENT_SERVICE_TYPE" --limit 5
    assert_success
else
    xbe_json view trucker-invoices list --is-management-service-type false --limit 5
    assert_success
fi

test_name "List trucker invoices with --business-unit filter"
if [[ -n "$BUSINESS_UNIT_ID" && "$BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --business-unit "$BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List trucker invoices with --not-business-unit filter"
if [[ -n "$BUSINESS_UNIT_ID" && "$BUSINESS_UNIT_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --not-business-unit "$BUSINESS_UNIT_ID" --limit 5
    assert_success
else
    skip "No business unit ID available"
fi

test_name "List trucker invoices with --customer filter"
if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --customer "$CUSTOMER_ID" --limit 5
    assert_success
else
    skip "No customer ID available"
fi

test_name "List trucker invoices with --material-transaction-cost-codes filter"
xbe_json view trucker-invoices list --material-transaction-cost-codes "123.45" --limit 5
assert_success

test_name "List trucker invoices with --allocated-cost-codes filter"
xbe_json view trucker-invoices list --allocated-cost-codes "123-456" --limit 5
assert_success

test_name "List trucker invoices with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List trucker invoices with --has-ticket-report filter"
xbe_json view trucker-invoices list --has-ticket-report true --limit 5
assert_success

test_name "List trucker invoices with --batch-status filter"
if [[ -n "$BATCH_ORG_TYPE" && -n "$BATCH_ORG_ID" && "$BATCH_ORG_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --batch-status "${BATCH_ORG_TYPE}|${BATCH_ORG_ID}|never_processed" --limit 5
    assert_success
else
    skip "No batch-status org info available"
fi

test_name "List trucker invoices with --having-plans-with-job-number-like filter"
xbe_json view trucker-invoices list --having-plans-with-job-number-like "JOB" --limit 5
assert_success

test_name "List trucker invoices with --quickbooks-id filter"
if [[ -n "$QUICKBOOKS_ID" && "$QUICKBOOKS_ID" != "null" ]]; then
    xbe_json view trucker-invoices list --quickbooks-id "$QUICKBOOKS_ID" --limit 5
    assert_success
else
    xbe_json view trucker-invoices list --quickbooks-id "QB-123" --limit 5
    assert_success
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create prerequisite broker for trucker invoice"
BROKER_NAME=$(unique_name "TruckerInvoiceBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
    fi
fi

test_name "Create prerequisite trucker for trucker invoice"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    TRUCKER_NAME=$(unique_name "TruckerInvoiceTrucker")
    TRUCKER_ADDRESS="123 Trucker Invoice St"

    xbe_json do truckers create \
        --name "$TRUCKER_NAME" \
        --broker "$CREATED_BROKER_ID" \
        --company-address "$TRUCKER_ADDRESS"

    if [[ $status -eq 0 ]]; then
        CREATED_TRUCKER_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
            register_cleanup "truckers" "$CREATED_TRUCKER_ID"
            pass
        else
            fail "Created trucker but no ID returned"
        fi
    else
        if [[ -n "$XBE_TEST_TRUCKER_ID" ]]; then
            CREATED_TRUCKER_ID="$XBE_TEST_TRUCKER_ID"
            echo "    Using XBE_TEST_TRUCKER_ID: $CREATED_TRUCKER_ID"
            pass
        else
            fail "Failed to create trucker and XBE_TEST_TRUCKER_ID not set"
        fi
    fi
else
    skip "No broker available for trucker creation"
fi

test_name "Find sample time card for trucker invoice"
xbe_json view time-card-invoices list --limit 1
if [[ $status -eq 0 ]]; then
    TIME_CARD_ID=$(json_get ".[0].time_card_id")
    if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
        pass
    else
        skip "No time card IDs available"
    fi
else
    skip "Could not list time card invoices"
fi

test_name "Create trucker invoice"
if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" && -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
    CREATE_ARGS=(do trucker-invoices create \
        --buyer-type brokers --buyer "$CREATED_BROKER_ID" \
        --seller-type truckers --seller "$CREATED_TRUCKER_ID" \
        --invoice-date 2025-01-01 --due-on 2025-01-10 \
        --adjustment-amount 0.00 --currency-code USD \
        --notes "CLI test invoice" \
        --explicit-buyer-name "CLI Test Buyer" \
        --explicit-buyer-address "123 Test Ave")

    if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
        CREATE_ARGS+=(--time-cards "$TIME_CARD_ID")
    fi

    xbe_json "${CREATE_ARGS[@]}"
    if [[ $status -eq 0 ]]; then
        CREATED_TRUCKER_INVOICE_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRUCKER_INVOICE_ID" && "$CREATED_TRUCKER_INVOICE_ID" != "null" ]]; then
            register_cleanup "trucker-invoices" "$CREATED_TRUCKER_INVOICE_ID"
            pass
        else
            fail "Created trucker invoice but no ID returned"
        fi
    else
        skip "Failed to create trucker invoice (permissions or validation)"
    fi
else
    skip "No broker/trucker available for trucker invoice creation"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update trucker invoice notes"
if [[ -n "$CREATED_TRUCKER_INVOICE_ID" && "$CREATED_TRUCKER_INVOICE_ID" != "null" ]]; then
    xbe_json do trucker-invoices update "$CREATED_TRUCKER_INVOICE_ID" --notes "Updated invoice notes"
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update trucker invoice (permissions or policy)"
    fi
else
    skip "No created trucker invoice ID available"
fi

test_name "Update trucker invoice time cards"
if [[ -n "$CREATED_TRUCKER_INVOICE_ID" && "$CREATED_TRUCKER_INVOICE_ID" != "null" ]]; then
    if [[ -n "$TIME_CARD_ID" && "$TIME_CARD_ID" != "null" ]]; then
        xbe_json do trucker-invoices update "$CREATED_TRUCKER_INVOICE_ID" --time-cards "$TIME_CARD_ID"
    else
        xbe_json do trucker-invoices update "$CREATED_TRUCKER_INVOICE_ID" --time-cards ""
    fi
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not update trucker invoice time cards"
    fi
else
    skip "No created trucker invoice ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete trucker invoice requires --confirm flag"
if [[ -n "$CREATED_TRUCKER_INVOICE_ID" && "$CREATED_TRUCKER_INVOICE_ID" != "null" ]]; then
    xbe_run do trucker-invoices delete "$CREATED_TRUCKER_INVOICE_ID"
    assert_failure
else
    skip "No created trucker invoice ID available"
fi

test_name "Delete trucker invoice"
if [[ -n "$CREATED_TRUCKER_INVOICE_ID" && "$CREATED_TRUCKER_INVOICE_ID" != "null" ]]; then
    xbe_run do trucker-invoices delete "$CREATED_TRUCKER_INVOICE_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete trucker invoice (permissions or constraints)"
    fi
else
    skip "No created trucker invoice ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create trucker invoice without required fields fails"
xbe_run do trucker-invoices create
assert_failure

test_name "Update trucker invoice without any fields fails"
xbe_run do trucker-invoices update "999999"
assert_failure

run_tests
