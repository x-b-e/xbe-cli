#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet Cost Code Allocations
#
# Tests CRUD operations for the time_sheet_cost_code_allocations resource.
#
# COVERAGE: All filters + all create/update attributes (best-effort with existing data)
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_ALLOCATION_ID=""
SAMPLE_ALLOCATION_ID=""
SAMPLE_TIME_SHEET_ID=""
SKIP_CREATE=0

describe "Resource: time-sheet-cost-code-allocations"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet cost code allocations"
xbe_json view time-sheet-cost-code-allocations list --limit 5
assert_success

test_name "List time sheet cost code allocations returns array"
xbe_json view time-sheet-cost-code-allocations list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list time sheet cost code allocations"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Locate time sheet cost code allocation for filters"
xbe_json view time-sheet-cost-code-allocations list --limit 20
if [[ $status -eq 0 ]]; then
    count=$(echo "$output" | jq 'length' 2>/dev/null)
    if [[ "$count" -gt 0 ]]; then
        SAMPLE_ALLOCATION_ID=$(json_get ".[0].id")
        SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
        pass
    else
        if [[ -n "$XBE_TEST_TIME_SHEET_COST_CODE_ALLOCATION_ID" ]]; then
            xbe_json view time-sheet-cost-code-allocations show "$XBE_TEST_TIME_SHEET_COST_CODE_ALLOCATION_ID"
            if [[ $status -eq 0 ]]; then
                SAMPLE_ALLOCATION_ID=$(json_get ".id")
                SAMPLE_TIME_SHEET_ID=$(json_get ".time_sheet_id")
                pass
            else
                skip "Failed to load XBE_TEST_TIME_SHEET_COST_CODE_ALLOCATION_ID"
            fi
        else
            skip "No time sheet cost code allocations found. Set XBE_TEST_TIME_SHEET_COST_CODE_ALLOCATION_ID for filter tests."
        fi
    fi
else
    fail "Failed to list time sheet cost code allocations for filters"
fi

# ============================================================================
# Show Tests
# ============================================================================

if [[ -n "$SAMPLE_ALLOCATION_ID" && "$SAMPLE_ALLOCATION_ID" != "null" ]]; then
    test_name "Show time sheet cost code allocation"
    xbe_json view time-sheet-cost-code-allocations show "$SAMPLE_ALLOCATION_ID"
    assert_success
else
    test_name "Show time sheet cost code allocation"
    skip "No sample allocation available"
fi

# ============================================================================
# Filter Tests
# ============================================================================

if [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
    test_name "Filter by time sheet"
    xbe_json view time-sheet-cost-code-allocations list --time-sheet "$SAMPLE_TIME_SHEET_ID"
    assert_success
else
    test_name "Filter by time sheet"
    skip "No time sheet ID available"
fi

test_name "List time sheet cost code allocations with --created-at-min filter"
xbe_json view time-sheet-cost-code-allocations list --created-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List time sheet cost code allocations with --created-at-max filter"
xbe_json view time-sheet-cost-code-allocations list --created-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List time sheet cost code allocations with --is-created-at filter"
xbe_json view time-sheet-cost-code-allocations list --is-created-at true --limit 5
assert_success

test_name "List time sheet cost code allocations with --updated-at-min filter"
xbe_json view time-sheet-cost-code-allocations list --updated-at-min "2024-01-01T00:00:00Z" --limit 5
assert_success

test_name "List time sheet cost code allocations with --updated-at-max filter"
xbe_json view time-sheet-cost-code-allocations list --updated-at-max "2026-12-31T23:59:59Z" --limit 5
assert_success

test_name "List time sheet cost code allocations with --is-updated-at filter"
xbe_json view time-sheet-cost-code-allocations list --is-updated-at true --limit 5
assert_success

# ============================================================================
# CREATE Tests
# ============================================================================

CREATE_TIME_SHEET_ID="${XBE_TEST_TIME_SHEET_ID:-}"
CREATE_COST_CODE_ID="${XBE_TEST_TIME_SHEET_COST_CODE_ID:-}"
UPDATE_COST_CODE_ID="${XBE_TEST_TIME_SHEET_COST_CODE_ID_2:-$CREATE_COST_CODE_ID}"

if [[ -n "$CREATE_TIME_SHEET_ID" && -n "$CREATE_COST_CODE_ID" ]]; then
    DETAILS_JSON=$(printf '[{\"cost_code_id\":\"%s\",\"percentage\":1}]' "$CREATE_COST_CODE_ID")

    test_name "Create time sheet cost code allocation"
    xbe_json do time-sheet-cost-code-allocations create \
        --time-sheet "$CREATE_TIME_SHEET_ID" \
        --details "$DETAILS_JSON"

    if [[ $status -eq 0 ]]; then
        CREATED_ALLOCATION_ID=$(json_get ".id")
        if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
            register_cleanup "time-sheet-cost-code-allocations" "$CREATED_ALLOCATION_ID"
            pass
        else
            fail "Created allocation but no ID returned"
        fi
    else
        if [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Create rejected (validation or authorization)"
            SKIP_CREATE=1
        else
            fail "Failed to create time sheet cost code allocation"
            SKIP_CREATE=1
        fi
    fi
else
    test_name "Create time sheet cost code allocation"
    skip "Set XBE_TEST_TIME_SHEET_ID and XBE_TEST_TIME_SHEET_COST_CODE_ID to enable create tests"
    SKIP_CREATE=1
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
    UPDATE_DETAILS_JSON=$(printf '[{\"cost_code_id\":\"%s\",\"percentage\":1}]' "$UPDATE_COST_CODE_ID")

    test_name "Update time sheet cost code allocation details"
    xbe_json do time-sheet-cost-code-allocations update "$CREATED_ALLOCATION_ID" --details "$UPDATE_DETAILS_JSON"
    assert_success

    test_name "Update time sheet cost code allocation without fields fails"
    xbe_json do time-sheet-cost-code-allocations update "$CREATED_ALLOCATION_ID"
    assert_failure
else
    test_name "Update time sheet cost code allocation details"
    skip "No allocation created"
    test_name "Update time sheet cost code allocation without fields fails"
    skip "No allocation created"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

if [[ -n "$CREATED_ALLOCATION_ID" && "$CREATED_ALLOCATION_ID" != "null" ]]; then
    test_name "Delete time sheet cost code allocation requires --confirm flag"
    xbe_run do time-sheet-cost-code-allocations delete "$CREATED_ALLOCATION_ID"
    assert_failure

    test_name "Delete time sheet cost code allocation with --confirm"
    xbe_run do time-sheet-cost-code-allocations delete "$CREATED_ALLOCATION_ID" --confirm
    assert_success
else
    test_name "Delete time sheet cost code allocation requires --confirm flag"
    skip "No allocation created"
    test_name "Delete time sheet cost code allocation with --confirm"
    skip "No allocation created"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
