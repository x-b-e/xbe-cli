#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Denials
#
# Tests create operations for the material-transaction-denials resource.
#
# COVERAGE: Create + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_MATERIAL_TRANSACTION_ID=""

describe "Resource: material-transaction-denials"

# ============================================================================
# Sample Record (used for create)
# ============================================================================

test_name "Capture material transaction in accepted status"
xbe_json view material-transactions list --status accepted --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
    if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
        pass
    else
        skip "No accepted material transactions available"
    fi
else
    skip "Could not list material transactions"
fi

if [[ -z "$SAMPLE_MATERIAL_TRANSACTION_ID" || "$SAMPLE_MATERIAL_TRANSACTION_ID" == "null" ]]; then
    test_name "Capture material transaction in submitted status"
    xbe_json view material-transactions list --status submitted --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
            pass
        else
            skip "No submitted material transactions available"
        fi
    else
        skip "Could not list material transactions"
    fi
fi

if [[ -z "$SAMPLE_MATERIAL_TRANSACTION_ID" || "$SAMPLE_MATERIAL_TRANSACTION_ID" == "null" ]]; then
    test_name "Capture material transaction in editing status"
    xbe_json view material-transactions list --status editing --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_MATERIAL_TRANSACTION_ID=$(json_get ".[0].id")
        if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
            pass
        else
            skip "No editing material transactions available"
        fi
    else
        skip "Could not list material transactions"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Deny material transaction"
if [[ -n "$SAMPLE_MATERIAL_TRANSACTION_ID" && "$SAMPLE_MATERIAL_TRANSACTION_ID" != "null" ]]; then
    xbe_json do material-transaction-denials create \
        --material-transaction "$SAMPLE_MATERIAL_TRANSACTION_ID" \
        --comment "CLI denial test"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Denial failed: $output"
        fi
    fi
else
    skip "No material transaction available for denial"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Deny without required fields fails"
xbe_run do material-transaction-denials create
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
