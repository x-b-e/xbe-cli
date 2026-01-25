#!/bin/bash
#
# XBE CLI Integration Tests: Haskell Lemon Outbound Material Transaction Exports
#
# Tests list, show, and create operations for the haskell-lemon-outbound-material-transaction-exports resource.
#
# COVERAGE: List filters + show + create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
TRANSACTION_DATE=""
CREATED_BY_ID=""
SKIP_ID_FILTERS=0
CREATED_EXPORT_ID=""

describe "Resource: haskell-lemon-outbound-material-transaction-exports"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Haskell Lemon outbound material transaction exports"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --limit 5
assert_success

test_name "List outbound material transaction exports returns array"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list exports"
fi

# ============================================================================
# Sample Data
# ============================================================================

test_name "Find sample outbound material transaction export"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    TRANSACTION_DATE=$(json_get ".[0].transaction_date")
    CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No outbound material transaction exports available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list exports"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List exports with --transaction-date filter"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$TRANSACTION_DATE" && "$TRANSACTION_DATE" != "null" ]]; then
    xbe_json view haskell-lemon-outbound-material-transaction-exports list --transaction-date "$TRANSACTION_DATE" --limit 5
    assert_success
else
    skip "No transaction date available"
fi

test_name "List exports with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view haskell-lemon-outbound-material-transaction-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

test_name "List exports with --transaction-date-min filter"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --transaction-date-min "2020-01-01" --limit 5
assert_success

test_name "List exports with --transaction-date-max filter"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --transaction-date-max "2030-01-01" --limit 5
assert_success

test_name "List exports with --has-transaction-date filter"
xbe_json view haskell-lemon-outbound-material-transaction-exports list --has-transaction-date true --limit 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show outbound material transaction export"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view haskell-lemon-outbound-material-transaction-exports show "$SAMPLE_ID"
    assert_success
else
    skip "No export ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create export without required fields fails"
xbe_run do haskell-lemon-outbound-material-transaction-exports create
assert_failure

# ============================================================================
# CREATE Tests
# ============================================================================

TEST_DATE="$(date -u +%Y-%m-%d)"
TO_ADDRESS="${XBE_TEST_HASKELL_LEMON_OUTBOUND_EXPORT_TO:-$(unique_email)}"
CC_ADDRESS="${XBE_TEST_HASKELL_LEMON_OUTBOUND_EXPORT_CC:-$(unique_email)}"

test_name "Create test outbound material transaction export"
xbe_json do haskell-lemon-outbound-material-transaction-exports create \
    --transaction-date "$TEST_DATE" \
    --is-test \
    --to-addresses "$TO_ADDRESS" \
    --cc-addresses "$CC_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_EXPORT_ID=$(json_get ".id")
    if [[ -n "$CREATED_EXPORT_ID" && "$CREATED_EXPORT_ID" != "null" ]]; then
        pass
    else
        fail "Created export but no ID returned"
    fi
else
    if [[ "$output" == *"It should be id 85, but the broker doesn't exist"* ]] || [[ "$output" == *"broker doesn't exist"* ]]; then
        skip "Broker for Haskell Lemon export not configured"
    elif [[ "$output" == *"403"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
        skip "Not authorized to create export"
    else
        fail "Failed to create export"
    fi
fi

test_name "Show created export includes requested attributes"
if [[ -n "$CREATED_EXPORT_ID" && "$CREATED_EXPORT_ID" != "null" ]]; then
    xbe_json view haskell-lemon-outbound-material-transaction-exports show "$CREATED_EXPORT_ID"
    if [[ $status -eq 0 ]]; then
        assert_json_equals ".transaction_date" "$TEST_DATE"
        assert_json_bool ".is_test" "true"
        assert_json_equals ".to_addresses[0]" "$TO_ADDRESS"
        assert_json_equals ".cc_addresses[0]" "$CC_ADDRESS"
    else
        fail "Failed to show created export"
    fi
else
    skip "No created export ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
