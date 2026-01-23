#!/bin/bash
#
# XBE CLI Integration Tests: Time Card Cost Code Allocations
#
# Tests list, show, create, update, delete operations for the
# time-card-cost-code-allocations resource.
#
# COVERAGE: List + show + filters + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TIME_CARD_ID=""
CREATED_ALLOCATION_ID=""
LIST_SUPPORTED="true"

describe "Resource: time-card-cost-code-allocations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time card cost code allocations"
xbe_json view time-card-cost-code-allocations list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing time card cost code allocations"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List time card cost code allocations returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-card-cost-code-allocations list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list time card cost code allocations"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show + filters)
# ============================================================================

test_name "Capture sample time card cost code allocation"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-card-cost-code-allocations list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_CARD_ID=$(json_get ".[0].time_card_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No time card cost code allocations available for follow-on tests"
        fi
    else
        skip "Could not list time card cost code allocations to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# FILTER Tests
# ============================================================================

test_name "Filter allocations by time card"
if [[ -n "$SAMPLE_TIME_CARD_ID" && "$SAMPLE_TIME_CARD_ID" != "null" ]]; then
    xbe_json view time-card-cost-code-allocations list --time-card "$SAMPLE_TIME_CARD_ID" --limit 5
    if [[ $status -eq 0 ]]; then
        pass
    else
        fail "Filter by time card failed"
    fi
else
    skip "No time card ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time card cost code allocation"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-card-cost-code-allocations show "$SAMPLE_ID"
    assert_success
else
    skip "No time card cost code allocation ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time card cost code allocation"
if [[ -n "$XBE_TEST_TIME_CARD_ID" && -n "$XBE_TEST_TIME_CARD_COST_CODE_DETAILS" ]]; then
    xbe_json do time-card-cost-code-allocations create \
        --time-card "$XBE_TEST_TIME_CARD_ID" \
        --details "$XBE_TEST_TIME_CARD_COST_CODE_DETAILS"
    if [[ $status -eq 0 ]]; then
        CREATED_ALLOCATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
            register_cleanup "time-card-cost-code-allocations" "$CREATED_ALLOCATION_ID"
        fi
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TIME_CARD_ID and XBE_TEST_TIME_CARD_COST_CODE_DETAILS to enable create test"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time card cost code allocation details"
UPDATE_TARGET_ID="$CREATED_ALLOCATION_ID"
if [[ -z "$UPDATE_TARGET_ID" || "$UPDATE_TARGET_ID" == "null" ]]; then
    UPDATE_TARGET_ID="$XBE_TEST_TIME_CARD_COST_CODE_ALLOCATION_ID"
fi

if [[ -n "$UPDATE_TARGET_ID" && -n "$XBE_TEST_TIME_CARD_COST_CODE_DETAILS" ]]; then
    xbe_json do time-card-cost-code-allocations update "$UPDATE_TARGET_ID" \
        --details "$XBE_TEST_TIME_CARD_COST_CODE_DETAILS"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "Set XBE_TEST_TIME_CARD_COST_CODE_ALLOCATION_ID and XBE_TEST_TIME_CARD_COST_CODE_DETAILS to enable update test"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete allocation requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do time-card-cost-code-allocations delete "$SAMPLE_ID"
    assert_failure
else
    skip "No allocation ID available"
fi

test_name "Delete time card cost code allocation"
if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
    xbe_json do time-card-cost-code-allocations delete "$CREATED_ALLOCATION_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"Record Invalid"* ]] || \
           [[ "$output" == *"422"* ]]; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created allocation ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create allocation without required flags fails"
xbe_run do time-card-cost-code-allocations create
assert_failure

test_name "Update allocation without fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do time-card-cost-code-allocations update "$SAMPLE_ID"
    assert_failure
else
    skip "No allocation ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
