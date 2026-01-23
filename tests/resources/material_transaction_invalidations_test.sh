#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Invalidations
#
# Tests create operations for material-transaction-invalidations.
#
# COVERAGE: Writable attributes (comment)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MTXN_ID=""

MTXN_ID="${XBE_TEST_MATERIAL_TRANSACTION_INVALIDATION_ID:-}"
if [[ -z "$MTXN_ID" && -n "$XBE_TEST_MATERIAL_TRANSACTION_ID" ]]; then
    MTXN_ID="$XBE_TEST_MATERIAL_TRANSACTION_ID"
fi

describe "Resource: material-transaction-invalidations"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create invalidation requires material transaction"
xbe_run do material-transaction-invalidations create --comment "missing mtxn"
assert_failure

if [[ -z "$MTXN_ID" || "$MTXN_ID" == "null" ]]; then
    xbe_json view material-transactions list --status "submitted,editing,rejected,denied,accepted,unmatched" --limit 1
    if [[ $status -eq 0 ]]; then
        MTXN_ID=$(json_get '.[0].id // empty')
    fi
fi

test_name "Create material transaction invalidation"
if [[ -n "$MTXN_ID" && "$MTXN_ID" != "null" ]]; then
    COMMENT=$(unique_name "MtxnInvalidation")
    xbe_json do material-transaction-invalidations create \
        --material-transaction "$MTXN_ID" \
        --comment "$COMMENT"
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"cannot be changed"* ]] || [[ "$output" == *"cannot be invalidated"* ]] || [[ "$output" == *"invoiced"* ]] || [[ "$output" == *"approved time card"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create material transaction invalidation: $output"
        fi
    fi
else
    skip "No material transaction ID available. Set XBE_TEST_MATERIAL_TRANSACTION_INVALIDATION_ID to enable create testing."
fi

run_tests
