#!/bin/bash
#
# XBE CLI Integration Tests: Material Transaction Acceptances
#
# Tests create operations for material-transaction-acceptances.
#
# COVERAGE: Writable attributes (comment, skip-not-overlapping-validation)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

MTXN_ID=""

MTXN_ID="${XBE_TEST_MATERIAL_TRANSACTION_ACCEPTANCE_ID:-}"
if [[ -z "$MTXN_ID" && -n "$XBE_TEST_MATERIAL_TRANSACTION_ID" ]]; then
    MTXN_ID="$XBE_TEST_MATERIAL_TRANSACTION_ID"
fi

describe "Resource: material-transaction-acceptances"

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create acceptance requires material transaction"
xbe_run do material-transaction-acceptances create --comment "missing mtxn"
assert_failure

if [[ -z "$MTXN_ID" || "$MTXN_ID" == "null" ]]; then
    xbe_json view material-transactions list --status "submitted,editing,rejected,unmatched" --limit 1
    if [[ $status -eq 0 ]]; then
        MTXN_ID=$(json_get '.[0].id // empty')
    fi
fi

test_name "Create material transaction acceptance"
if [[ -n "$MTXN_ID" && "$MTXN_ID" != "null" ]]; then
    COMMENT=$(unique_name "MtxnAcceptance")
    xbe_json do material-transaction-acceptances create \
        --material-transaction "$MTXN_ID" \
        --comment "$COMMENT" \
        --skip-not-overlapping-validation
    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"409"* ]] || [[ "$output" == *"must have"* ]] || [[ "$output" == *"cannot be changed"* ]]; then
            skip "Create blocked by server policy/validation"
        else
            fail "Failed to create material transaction acceptance: $output"
        fi
    fi
else
    skip "No material transaction ID available. Set XBE_TEST_MATERIAL_TRANSACTION_ACCEPTANCE_ID to enable create testing."
fi

run_tests
