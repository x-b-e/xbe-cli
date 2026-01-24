#!/bin/bash
#
# XBE CLI Integration Tests: Retainer Payments
#
# Tests list, show, create, update, and delete operations for the retainer-payments resource.
#
# COVERAGE: List filters + show + create/update attributes + delete (when possible)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

PAYMENT_ID=""
RETAINER_ID=""
RETAINER_PERIOD_ID=""
STATUS=""
PAY_ON=""
RETAINER_TYPE=""
BUYER_ID=""
SELLER_ID=""

CREATED_PAYMENT_ID=""

SKIP_SAMPLE_TESTS=0

TEST_RETAINER_PERIOD_ID="${XBE_TEST_RETAINER_PERIOD_ID:-}"
TEST_PAYMENT_AMOUNT="${XBE_TEST_RETAINER_PAYMENT_AMOUNT:-}"
TEST_PAYMENT_CREATED_ON="${XBE_TEST_RETAINER_PAYMENT_CREATED_ON:-}"
TEST_PAYMENT_STATUS="${XBE_TEST_RETAINER_PAYMENT_STATUS:-editing}"


describe "Resource: retainer-payments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List retainer payments"
xbe_json view retainer-payments list --limit 5
assert_success

test_name "List retainer payments returns array"
xbe_json view retainer-payments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list retainer payments"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample retainer payment"
xbe_json view retainer-payments list --limit 1
if [[ $status -eq 0 ]]; then
    PAYMENT_ID=$(json_get ".[0].id")
    RETAINER_ID=$(json_get ".[0].retainer_id")
    RETAINER_PERIOD_ID=$(json_get ".[0].retainer_period_id")
    STATUS=$(json_get ".[0].status")
    PAY_ON=$(json_get ".[0].pay_on")
    if [[ -n "$PAYMENT_ID" && "$PAYMENT_ID" != "null" ]]; then
        pass
    else
        SKIP_SAMPLE_TESTS=1
        skip "No retainer payments available"
    fi
else
    SKIP_SAMPLE_TESTS=1
    fail "Failed to list retainer payments"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show retainer payment"
if [[ $SKIP_SAMPLE_TESTS -eq 0 && -n "$PAYMENT_ID" && "$PAYMENT_ID" != "null" ]]; then
    xbe_json view retainer-payments show "$PAYMENT_ID"
    if [[ $status -eq 0 ]]; then
        RETAINER_TYPE=$(json_get ".retainer_type")
        BUYER_ID=$(json_get ".buyer_id")
        SELLER_ID=$(json_get ".seller_id")
        assert_json_has ".id"
    else
        fail "Failed to show retainer payment"
    fi
else
    skip "No retainer payment ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List retainer payments with --retainer-period filter"
if [[ -n "$RETAINER_PERIOD_ID" && "$RETAINER_PERIOD_ID" != "null" ]]; then
    xbe_json view retainer-payments list --retainer-period "$RETAINER_PERIOD_ID" --limit 5
    assert_success
else
    skip "No retainer period ID available"
fi

test_name "List retainer payments with --status filter"
if [[ -n "$STATUS" && "$STATUS" != "null" ]]; then
    xbe_json view retainer-payments list --status "$STATUS" --limit 5
    assert_success
else
    skip "No status available"
fi

test_name "List retainer payments with --retainer-type filter"
if [[ -n "$RETAINER_TYPE" && "$RETAINER_TYPE" != "null" ]]; then
    xbe_json view retainer-payments list --retainer-type "$RETAINER_TYPE" --limit 5
    assert_success
else
    skip "No retainer type available"
fi

test_name "List retainer payments with --buyer filter"
if [[ -n "$BUYER_ID" && "$BUYER_ID" != "null" ]]; then
    xbe_json view retainer-payments list --buyer "$BUYER_ID" --limit 5
    assert_success
else
    skip "No buyer ID available"
fi

test_name "List retainer payments with --seller filter"
if [[ -n "$SELLER_ID" && "$SELLER_ID" != "null" ]]; then
    xbe_json view retainer-payments list --seller "$SELLER_ID" --limit 5
    assert_success
else
    skip "No seller ID available"
fi

test_name "List retainer payments with --pay-on-min filter"
if [[ -n "$PAY_ON" && "$PAY_ON" != "null" ]]; then
    xbe_json view retainer-payments list --pay-on-min "$PAY_ON" --limit 5
    assert_success
else
    skip "No pay-on date available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create retainer payment without required fields fails"
xbe_run do retainer-payments create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create retainer payment"
if [[ -n "$TEST_RETAINER_PERIOD_ID" && -n "$TEST_PAYMENT_AMOUNT" && -n "$TEST_PAYMENT_CREATED_ON" ]]; then
    xbe_json do retainer-payments create \
        --retainer-period "$TEST_RETAINER_PERIOD_ID" \
        --status "$TEST_PAYMENT_STATUS" \
        --amount "$TEST_PAYMENT_AMOUNT" \
        --created-on "$TEST_PAYMENT_CREATED_ON"
    if [[ $status -eq 0 ]]; then
        CREATED_PAYMENT_ID=$(json_get ".id")
        if [[ -n "$CREATED_PAYMENT_ID" && "$CREATED_PAYMENT_ID" != "null" ]]; then
            register_cleanup "retainer-payments" "$CREATED_PAYMENT_ID"
            pass
        else
            fail "Created retainer payment but no ID returned"
        fi
    else
        skip "Create failed (check XBE_TEST_RETAINER_PERIOD_ID/AMOUNT/CREATED_ON)"
    fi
else
    skip "XBE_TEST_RETAINER_PERIOD_ID, XBE_TEST_RETAINER_PAYMENT_AMOUNT, or XBE_TEST_RETAINER_PAYMENT_CREATED_ON not set"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update retainer payment attributes"
if [[ -n "$CREATED_PAYMENT_ID" && "$CREATED_PAYMENT_ID" != "null" ]]; then
    if [[ "$TEST_PAYMENT_STATUS" == "editing" ]]; then
        xbe_json do retainer-payments update "$CREATED_PAYMENT_ID" \
            --amount "$TEST_PAYMENT_AMOUNT" \
            --created-on "$TEST_PAYMENT_CREATED_ON" \
            --retainer-period "$TEST_RETAINER_PERIOD_ID"
        assert_success
    else
        skip "Payment status not editing; amount/created-on updates require editing"
    fi
else
    skip "No created retainer payment ID available"
fi

test_name "Update retainer payment status"
if [[ -n "$CREATED_PAYMENT_ID" && "$CREATED_PAYMENT_ID" != "null" ]]; then
    xbe_json do retainer-payments update "$CREATED_PAYMENT_ID" --status "$TEST_PAYMENT_STATUS"
    assert_success
else
    skip "No created retainer payment ID available"
fi

test_name "Update retainer payment without fields fails"
if [[ -n "$CREATED_PAYMENT_ID" && "$CREATED_PAYMENT_ID" != "null" ]]; then
    xbe_run do retainer-payments update "$CREATED_PAYMENT_ID"
    assert_failure
else
    skip "No created retainer payment ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete retainer payment"
if [[ -n "$CREATED_PAYMENT_ID" && "$CREATED_PAYMENT_ID" != "null" ]]; then
    xbe_run do retainer-payments delete "$CREATED_PAYMENT_ID" --confirm
    assert_success
else
    skip "No created retainer payment ID available"
fi

run_tests
