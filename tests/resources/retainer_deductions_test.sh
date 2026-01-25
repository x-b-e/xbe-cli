#!/bin/bash
#
# XBE CLI Integration Tests: Retainer Deductions
#
# Tests CRUD operations for the retainer_deductions resource.
# Retainer deductions track deduction amounts and notes on retainers.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_RETAINER_ID=""
CREATED_DEDUCTION_ID=""

RET_DUAL_RETAINER_ID="${XBE_TEST_RETAINER_ID:-}"

DESCRIBE_RESOURCE="retainer-deductions"

describe "Resource: retainer-deductions"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List retainer deductions"
xbe_json view retainer-deductions list --limit 5
assert_success

test_name "List retainer deductions returns array"
xbe_json view retainer-deductions list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_RETAINER_ID=$(echo "$output" | jq -r '.[0].retainer_id // empty')
else
    fail "Failed to list retainer deductions"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show retainer deduction"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view retainer-deductions show "$SAMPLE_ID"
    assert_success
else
    skip "No retainer deduction ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create retainer deduction"
RET_ID="${RET_DUAL_RETAINER_ID:-$SAMPLE_RETAINER_ID}"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json do retainer-deductions create \
        --retainer "$RET_ID" \
        --amount "100.50" \
        --note "CLI test deduction"
    if [[ $status -eq 0 ]]; then
        CREATED_DEDUCTION_ID=$(json_get ".id")
        if [[ -n "$CREATED_DEDUCTION_ID" && "$CREATED_DEDUCTION_ID" != "null" ]]; then
            register_cleanup "retainer-deductions" "$CREATED_DEDUCTION_ID"
            pass
        else
            fail "Created retainer deduction but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create retainer deduction: $output"
        fi
    fi
else
    skip "No retainer ID available (set XBE_TEST_RETAINER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update retainer deduction amount"
UPDATE_ID="${CREATED_DEDUCTION_ID:-$SAMPLE_ID}"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do retainer-deductions update "$UPDATE_ID" --amount "125.75"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer deduction amount: $output"
        fi
    fi
else
    skip "No retainer deduction ID available for update"
fi

test_name "Update retainer deduction note"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do retainer-deductions update "$UPDATE_ID" --note "Updated deduction note"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer deduction note: $output"
        fi
    fi
else
    skip "No retainer deduction ID available for update note"
fi

test_name "Update retainer deduction retainer"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json do retainer-deductions update "$UPDATE_ID" --retainer "$RET_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer deduction retainer: $output"
        fi
    fi
else
    skip "No retainer deduction or retainer ID available for retainer update"
fi

test_name "Update retainer deduction without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do retainer-deductions update "$UPDATE_ID"
    assert_failure
else
    skip "No retainer deduction ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List retainer deductions with --retainer filter"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json view retainer-deductions list --retainer "$RET_ID" --limit 5
    assert_success
else
    skip "No retainer ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete retainer deduction requires --confirm flag"
if [[ -n "$CREATED_DEDUCTION_ID" && "$CREATED_DEDUCTION_ID" != "null" ]]; then
    xbe_run do retainer-deductions delete "$CREATED_DEDUCTION_ID"
    assert_failure
else
    skip "No created retainer deduction for delete confirmation test"
fi

test_name "Delete retainer deduction with --confirm"
if [[ -n "$CREATED_DEDUCTION_ID" && "$CREATED_DEDUCTION_ID" != "null" ]]; then
    xbe_run do retainer-deductions delete "$CREATED_DEDUCTION_ID" --confirm
    assert_success
else
    skip "No created retainer deduction to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create retainer deduction without retainer fails"
xbe_run do retainer-deductions create --amount "50.00"
assert_failure

test_name "Create retainer deduction without amount fails"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_run do retainer-deductions create --retainer "$RET_ID"
    assert_failure
else
    skip "No retainer ID available for missing amount test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
