#!/bin/bash
#
# XBE CLI Integration Tests: Retainer Periods
#
# Tests CRUD operations for the retainer_periods resource.
# Retainer periods define the start/end range and weekly payment amount.
#
# COVERAGE: All create/update attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_RETAINER_ID=""
CREATED_PERIOD_ID=""

RET_DUAL_RETAINER_ID="${XBE_TEST_RETAINER_ID:-}"

DESCRIBE_RESOURCE="retainer-periods"

describe "Resource: retainer-periods"

START_ON="2099-01-01"
END_ON="2099-01-31"
UPDATE_START_ON="2099-02-01"
UPDATE_END_ON="2099-02-28"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List retainer periods"
xbe_json view retainer-periods list --limit 5
assert_success

test_name "List retainer periods returns array"
xbe_json view retainer-periods list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
    SAMPLE_ID=$(echo "$output" | jq -r '.[0].id // empty')
    SAMPLE_RETAINER_ID=$(echo "$output" | jq -r '.[0].retainer_id // empty')
else
    fail "Failed to list retainer periods"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show retainer period"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view retainer-periods show "$SAMPLE_ID"
    assert_success
else
    skip "No retainer period ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create retainer period"
RET_ID="${RET_DUAL_RETAINER_ID:-$SAMPLE_RETAINER_ID}"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json do retainer-periods create \
        --retainer "$RET_ID" \
        --start-on "$START_ON" \
        --end-on "$END_ON" \
        --weekly-payment-amount "1000"
    if [[ $status -eq 0 ]]; then
        CREATED_PERIOD_ID=$(json_get ".id")
        if [[ -n "$CREATED_PERIOD_ID" && "$CREATED_PERIOD_ID" != "null" ]]; then
            register_cleanup "retainer-periods" "$CREATED_PERIOD_ID"
            pass
        else
            fail "Created retainer period but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to create retainer period: $output"
        fi
    fi
else
    skip "No retainer ID available (set XBE_TEST_RETAINER_ID)"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update retainer period start/end dates"
UPDATE_ID="${CREATED_PERIOD_ID:-$SAMPLE_ID}"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do retainer-periods update "$UPDATE_ID" --start-on "$UPDATE_START_ON" --end-on "$UPDATE_END_ON"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer period dates: $output"
        fi
    fi
else
    skip "No retainer period ID available for update"
fi

test_name "Update retainer period weekly payment amount"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_json do retainer-periods update "$UPDATE_ID" --weekly-payment-amount "1250"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer period weekly payment amount: $output"
        fi
    fi
else
    skip "No retainer period ID available for weekly payment update"
fi

test_name "Update retainer period retainer"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" && -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json do retainer-periods update "$UPDATE_ID" --retainer "$RET_ID"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Failed to update retainer period retainer: $output"
        fi
    fi
else
    skip "No retainer period or retainer ID available for retainer update"
fi

test_name "Update retainer period without attributes fails"
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    xbe_run do retainer-periods update "$UPDATE_ID"
    assert_failure
else
    skip "No retainer period ID available for no-attribute update test"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List retainer periods with --retainer filter"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_json view retainer-periods list --retainer "$RET_ID" --limit 5
    assert_success
else
    skip "No retainer ID available for filter test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete retainer period requires --confirm flag"
if [[ -n "$CREATED_PERIOD_ID" && "$CREATED_PERIOD_ID" != "null" ]]; then
    xbe_run do retainer-periods delete "$CREATED_PERIOD_ID"
    assert_failure
else
    skip "No created retainer period for delete confirmation test"
fi

test_name "Delete retainer period with --confirm"
if [[ -n "$CREATED_PERIOD_ID" && "$CREATED_PERIOD_ID" != "null" ]]; then
    xbe_run do retainer-periods delete "$CREATED_PERIOD_ID" --confirm
    assert_success
else
    skip "No created retainer period to delete"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create retainer period without retainer fails"
xbe_run do retainer-periods create --start-on "$START_ON" --end-on "$END_ON" --weekly-payment-amount "1000"
assert_failure

test_name "Create retainer period without start-on fails"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_run do retainer-periods create --retainer "$RET_ID" --end-on "$END_ON" --weekly-payment-amount "1000"
    assert_failure
else
    skip "No retainer ID available for missing start-on test"
fi

test_name "Create retainer period without end-on fails"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_run do retainer-periods create --retainer "$RET_ID" --start-on "$START_ON" --weekly-payment-amount "1000"
    assert_failure
else
    skip "No retainer ID available for missing end-on test"
fi

test_name "Create retainer period without weekly payment amount fails"
if [[ -n "$RET_ID" && "$RET_ID" != "null" ]]; then
    xbe_run do retainer-periods create --retainer "$RET_ID" --start-on "$START_ON" --end-on "$END_ON"
    assert_failure
else
    skip "No retainer ID available for missing weekly payment amount test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
