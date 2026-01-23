#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Status Changes
#
# Tests list/show operations for the material-transaction-status-changes resource.
#
# COVERAGE: All filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_STATUS_CHANGE_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID=""
SAMPLE_STATUS=""
SAMPLE_CREATED_AT=""
SAMPLE_UPDATED_AT=""

describe "Resource: material-transaction-status-changes"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction status changes"
xbe_json view material-transaction-status-changes list --limit 5
assert_success

test_name "List material transaction status changes returns array"
xbe_json view material-transaction-status-changes list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list material transaction status changes"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate material transaction status change for filters"
xbe_json view material-transaction-status-changes list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_STATUS_CHANGE_ID=$(json_get ".[0].id")
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].material_transaction_id")
        SAMPLE_STATUS=$(json_get ".[0].status")
        SAMPLE_CREATED_AT=$(json_get ".[0].created_at")
        SAMPLE_UPDATED_AT=$(json_get ".[0].updated_at")
        if [[ -z "$SAMPLE_MATERIAL_TRANSACTION_ID" || "$SAMPLE_MATERIAL_TRANSACTION_ID" == "null" ]]; then
            xbe_json view material-transaction-status-changes show "$SAMPLE_STATUS_CHANGE_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".material_transaction_id")
                SAMPLE_STATUS=$(json_get ".status")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
            fi
        fi
        pass
    else
        if [[ -n "$XBE_TEST_MATERIAL_TRANSACTION_STATUS_CHANGE_ID" ]]; then
            xbe_json view material-transaction-status-changes show "$XBE_TEST_MATERIAL_TRANSACTION_STATUS_CHANGE_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_STATUS_CHANGE_ID=$(json_get ".id")
                SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".material_transaction_id")
                SAMPLE_STATUS=$(json_get ".status")
                SAMPLE_CREATED_AT=$(json_get ".created_at")
                SAMPLE_UPDATED_AT=$(json_get ".updated_at")
                pass
            else
                skip "Failed to load XBE_TEST_MATERIAL_TRANSACTION_STATUS_CHANGE_ID"
            fi
        else
            skip "No status changes found. Set XBE_TEST_MATERIAL_TRANSACTION_STATUS_CHANGE_ID for filter tests."
        fi
    fi
else
    fail "Failed to list material transaction status changes for filters"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

if [[ -n "$SAMPLE_STATUS_CHANGE_ID" && "$SAMPLE_STATUS_CHANGE_ID" != "null" ]]; then
    test_name "Show material transaction status change"
    xbe_json view material-transaction-status-changes show "$SAMPLE_STATUS_CHANGE_ID"
    assert_success
else
    test_name "Show material transaction status change"
    skip "No sample status change available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$XBE_TEST_MATERIAL_TRANSACTION_ID" ]]; then
    SAMPLE_MATERIAL_TRANSACTION_ID="$XBE_TEST_MATERIAL_TRANSACTION_ID"
fi

if [[ -z "$SAMPLE_MATERIAL_TRANSACTION_ID" || "$SAMPLE_MATERIAL_TRANSACTION_ID" == "null" ]]; then
    xbe_json view material-transactions list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
    fi
fi

if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    test_name "Filter by material transaction"
    xbe_json view material-transaction-status-changes list --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID"
    assert_success
else
    test_name "Filter by material transaction"
    skip "No material transaction ID available"
fi

if [[ -n "$SAMPLE_STATUS" && "$SAMPLE_STATUS" != "null" ]]; then
    test_name "Filter by status"
    xbe_json view material-transaction-status-changes list --status "$SAMPLE_STATUS"
    assert_success
else
    test_name "Filter by status"
    skip "No status available"
fi

if [[ -n "$SAMPLE_CREATED_AT" && "$SAMPLE_CREATED_AT" != "null" ]]; then
    test_name "Filter by created-at min/max"
    xbe_json view material-transaction-status-changes list \
        --created-at-min "$SAMPLE_CREATED_AT" \
        --created-at-max "$SAMPLE_CREATED_AT"
    assert_success

    test_name "Filter by is-created-at"
    xbe_json view material-transaction-status-changes list --is-created-at true --limit 5
    assert_success
else
    test_name "Filter by created-at min/max"
    skip "No created-at available"
    test_name "Filter by is-created-at"
    skip "No created-at available"
fi

if [[ -n "$SAMPLE_UPDATED_AT" && "$SAMPLE_UPDATED_AT" != "null" ]]; then
    test_name "Filter by updated-at min/max"
    xbe_json view material-transaction-status-changes list \
        --updated-at-min "$SAMPLE_UPDATED_AT" \
        --updated-at-max "$SAMPLE_UPDATED_AT"
    assert_success

    test_name "Filter by is-updated-at"
    xbe_json view material-transaction-status-changes list --is-updated-at true --limit 5
    assert_success
else
    test_name "Filter by updated-at min/max"
    skip "No updated-at available"
    test_name "Filter by is-updated-at"
    skip "No updated-at available"
fi

test_name "List material transaction status changes with --offset"
xbe_json view material-transaction-status-changes list --limit 3 --offset 1
assert_success

test_name "List material transaction status changes with --sort"
xbe_json view material-transaction-status-changes list --sort created-at --limit 5
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
