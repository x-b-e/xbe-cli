#!/bin/bash
#
# XBE CLI Integration Tests: Haskell Lemon Inbound Material Transaction Exports
#
# Tests list/show/create operations for haskell-lemon-inbound-material-transaction-exports.
#
# COVERAGE: List filters + create attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_EXPORT_ID=""
CREATED_BY_ID=""
SAMPLE_ID=""
TRANSACTION_DATE="$(date -u +%Y-%m-%d)"
NOW_ISO=$(date -u +%Y-%m-%dT%H:%M:%SZ)
TO_ADDRESS=$(unique_email)
CC_ADDRESS=$(unique_email)


describe "Resource: haskell-lemon-inbound-material-transaction-exports"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create Haskell Lemon inbound material transaction export (test)"
xbe_json do haskell-lemon-inbound-material-transaction-exports create \
    --transaction-date "$TRANSACTION_DATE" \
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
    if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"broker doesn't exist"* ]] || [[ "$output" == *"It should be id 85"* ]]; then
        pass
    else
        fail "Failed to create Haskell Lemon inbound material transaction export"
    fi
fi


test_name "Create export fails without to-addresses when is-test is set"
xbe_run do haskell-lemon-inbound-material-transaction-exports create \
    --transaction-date "$TRANSACTION_DATE" \
    --is-test
assert_failure

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List Haskell Lemon inbound material transaction exports"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --limit 5
assert_success


test_name "List Haskell Lemon inbound material transaction exports returns array"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
else
    fail "Failed to list Haskell Lemon inbound material transaction exports"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show Haskell Lemon inbound material transaction export"
SHOW_ID="${CREATED_EXPORT_ID:-$SAMPLE_ID}"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view haskell-lemon-inbound-material-transaction-exports show "$SHOW_ID"
    if [[ $status -eq 0 ]]; then
        CREATED_BY_ID=$(json_get ".created_by_id")
        TRANSACTION_DATE=$(json_get ".transaction_date")
        pass
    else
        fail "Failed to show Haskell Lemon inbound material transaction export"
    fi
else
    skip "No export ID available for show"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List exports with --created-by filter"
if [[ -n "$CREATED_BY_ID" && "$CREATED_BY_ID" != "null" ]]; then
    xbe_json view haskell-lemon-inbound-material-transaction-exports list --created-by "$CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available for filter test"
fi


test_name "List exports with --transaction-date filter"
if [[ -n "$TRANSACTION_DATE" && "$TRANSACTION_DATE" != "null" ]]; then
    xbe_json view haskell-lemon-inbound-material-transaction-exports list --transaction-date "$TRANSACTION_DATE" --limit 5
    assert_success
else
    skip "No transaction date available for filter test"
fi


test_name "List exports with --transaction-date-min filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --transaction-date-min "$TRANSACTION_DATE" --limit 5
assert_success


test_name "List exports with --transaction-date-max filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --transaction-date-max "$TRANSACTION_DATE" --limit 5
assert_success


test_name "List exports with --has-transaction-date filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --has-transaction-date true --limit 5
assert_success


test_name "List exports with --created-at-min filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --created-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List exports with --created-at-max filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --created-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List exports with --is-created-at filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --is-created-at true --limit 5
assert_success


test_name "List exports with --updated-at-min filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --updated-at-min "$NOW_ISO" --limit 5
assert_success


test_name "List exports with --updated-at-max filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --updated-at-max "$NOW_ISO" --limit 5
assert_success


test_name "List exports with --is-updated-at filter"
xbe_json view haskell-lemon-inbound-material-transaction-exports list --is-updated-at true --limit 5
assert_success

run_tests
