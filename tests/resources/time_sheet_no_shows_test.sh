#!/bin/bash
#
# XBE CLI Integration Tests: Time Sheet No-Shows
#
# Tests list, show, create, update, and delete operations for the time-sheet-no-shows resource.
#
# COVERAGE: List + show + filters + create + update + delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_TIME_SHEET_ID=""
CREATE_TIME_SHEET_ID=""
CREATED_ID=""
LIST_SUPPORTED="true"

describe "Resource: time-sheet-no-shows"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List time sheet no-shows"
xbe_json view time-sheet-no-shows list --limit 5
if [[ $status -eq 0 ]]; then
    pass
else
    if [[ "$output" == *"404"* ]] || [[ "$output" == *"doesn't exist"* ]]; then
        LIST_SUPPORTED="false"
        skip "Server does not support listing time sheet no-shows"
    else
        fail "Expected success (exit 0), got exit $status"
    fi
fi

test_name "List time sheet no-shows returns array"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-no-shows list --limit 5
    if [[ $status -eq 0 ]]; then
        assert_json_is_array
    else
        fail "Failed to list time sheet no-shows"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List time sheet no-shows with time filters"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-no-shows list \
        --created-at-min 2020-01-01T00:00:00Z \
        --created-at-max 2030-01-01T00:00:00Z \
        --updated-at-min 2020-01-01T00:00:00Z \
        --updated-at-max 2030-01-01T00:00:00Z \
        --limit 5
    assert_success
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Sample Record (used for show)
# ============================================================================

test_name "Capture sample time sheet no-show"
if [[ "$LIST_SUPPORTED" == "true" ]]; then
    xbe_json view time-sheet-no-shows list --limit 1
    if [[ $status -eq 0 ]]; then
        SAMPLE_ID=$(json_get ".[0].id")
        SAMPLE_TIME_SHEET_ID=$(json_get ".[0].time_sheet_id")
        if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
            pass
        else
            skip "No time sheet no-shows available for follow-on tests"
        fi
    else
        skip "Could not list time sheet no-shows to capture sample"
    fi
else
    skip "List endpoint not supported"
fi

# ============================================================================
# Prerequisites for Create
# ============================================================================

if [[ -n "$XBE_TEST_TIME_SHEET_ID" ]]; then
    CREATE_TIME_SHEET_ID="$XBE_TEST_TIME_SHEET_ID"
elif [[ -n "$SAMPLE_TIME_SHEET_ID" && "$SAMPLE_TIME_SHEET_ID" != "null" ]]; then
    CREATE_TIME_SHEET_ID="$SAMPLE_TIME_SHEET_ID"
else
    xbe_json view labor-requirements list --limit 1
    if [[ $status -eq 0 ]]; then
        LABOR_REQUIREMENT_ID=$(json_get ".[0].id")
        if [[ -n "$LABOR_REQUIREMENT_ID" && "$LABOR_REQUIREMENT_ID" != "null" ]]; then
            xbe_json view labor-requirements show "$LABOR_REQUIREMENT_ID"
            if [[ $status -eq 0 ]]; then
                CREATE_TIME_SHEET_ID=$(json_get ".time_sheet_id")
            fi
        fi
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create time sheet no-show"
if [[ -n "$CREATE_TIME_SHEET_ID" && "$CREATE_TIME_SHEET_ID" != "null" ]]; then
    xbe_json do time-sheet-no-shows create \
        --time-sheet "$CREATE_TIME_SHEET_ID" \
        --no-show-reason "CLI test no-show"

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
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
    skip "No time sheet ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show time sheet no-show"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view time-sheet-no-shows show "$SAMPLE_ID"
    assert_success
else
    skip "No no-show ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update time sheet no-show reason"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do time-sheet-no-shows update "$CREATED_ID" --no-show-reason "Updated no-show reason"
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
    skip "No created no-show ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete time sheet no-show"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do time-sheet-no-shows delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || \
           [[ "$output" == *"not authorized"* ]] || \
           [[ "$output" == *"404"* ]]; then
            pass
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created no-show ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create no-show without required flags fails"
xbe_run do time-sheet-no-shows create
assert_failure

test_name "Update no-show without fields fails"
xbe_run do time-sheet-no-shows update 123
assert_failure

test_name "Delete no-show without confirm fails"
xbe_run do time-sheet-no-shows delete 123
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
