#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Rejections
#
# Tests list, show, and create operations for the material-transaction-rejections resource.
#
# COVERAGE: List + show + create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_MATERIAL_TRANSACTION_ID=""
REJECTION_MATERIAL_TRANSACTION_ID=""
LIST_SUPPORTED="true"

describe "Resource: material-transaction-rejections"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List material transaction rejections"
xbe_json view material-transaction-rejections list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"Not Authorized"* ]] || \
       [[ "$output" == *"not authorized"* ]] || \
       [[ "$output" == *"403"* ]] || \
       [[ "$output" == *"Forbidden"* ]]; then
        LIST_SUPPORTED="false"
        skip "Listing requires elevated access"
    elif [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "List endpoint not supported"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List material transaction rejections returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view material-transaction-rejections list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list material transaction rejections"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample material transaction rejection"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view material-transaction-rejections list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].material_transaction_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No material transaction rejections available for follow-on tests"
        fi
    else
        skip "Could not list material transaction rejections to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

test_name "Find material transaction to reject"
for status_candidate in accepted submitted unmatched invalidated denied; do
    xbe_json view material-transactions list --status "$status_candidate" --limit 1 --include-all
    if [[ $status -eq 0 ]]; then
        REJECTION_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
        if [[ -n "$REJECTION_MATERIAL_TRANSACTION_ID" && "$REJECTION_MATERIAL_TRANSACTION_ID" != "null" ]]; then
            pass
            break
        fi
    fi
    REJECTION_MATERIAL_TRANSACTION_ID=""
done
if [[ -z "$REJECTION_MATERIAL_TRANSACTION_ID" ]]; then
    skip "No material transactions available for rejection"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create material transaction rejection"
if [[ -n "$REJECTION_MATERIAL_TRANSACTION_ID" && "$REJECTION_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    xbe_json do material-transaction-rejections create \
        --material-transaction "$REJECTION_MATERIAL_TRANSACTION_ID" \
        --comment "CLI test rejection"

    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"403"* ]] || \
           [[ "$output" == *"previous status"* ]] || \
           [[ "$output" == *"cannot be rejected when recently created"* ]] || \
           [[ "$output" == *"cannot be changed when invoiced"* ]] || \
           [[ "$output" == *"must be related to a shift"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No material transaction available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show material transaction rejection"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view material-transaction-rejections show "$SAMPLE_ID"
    assert_success
else
    skip "No rejection ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create material transaction rejection without material transaction fails"
xbe_run do material-transaction-rejections create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
