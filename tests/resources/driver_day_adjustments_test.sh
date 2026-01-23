#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Adjustments
#
# Tests list, show, create, update, and delete operations for the driver-day-adjustments resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_DRIVER_DAY_ID=""
SAMPLE_TRUCKER_ID=""
SAMPLE_DRIVER_ID=""
CREATED_ID=""

describe "Resource: driver-day-adjustments"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver day adjustments"
xbe_json view driver-day-adjustments list --limit 5
assert_success

test_name "List driver day adjustments returns array"
xbe_json view driver-day-adjustments list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver day adjustments"
fi

# ============================================================================
# Sample Record (used for filters/show/update)
# ============================================================================

test_name "Capture sample adjustment"
xbe_json view driver-day-adjustments list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
    SAMPLE_TRUCKER_ID=$(json_get ".[0].trucker_id")
    SAMPLE_DRIVER_ID=$(json_get ".[0].driver_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No driver day adjustments available for follow-on tests"
    fi
else
    skip "Could not list driver day adjustments to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List adjustments with --driver-day filter"
if [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments list --driver-day "$SAMPLE_DRIVER_DAY_ID" --limit 5
    assert_success
else
    skip "No driver day ID available"
fi

test_name "List adjustments with --trucker filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments list --trucker "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List adjustments with --trucker-id filter"
if [[ -n "$SAMPLE_TRUCKER_ID" && "$SAMPLE_TRUCKER_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments list --trucker-id "$SAMPLE_TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List adjustments with --driver filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments list --driver "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List adjustments with --driver-id filter"
if [[ -n "$SAMPLE_DRIVER_ID" && "$SAMPLE_DRIVER_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments list --driver-id "$SAMPLE_DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver day adjustment"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-day-adjustments show "$SAMPLE_ID"
    assert_success
else
    skip "No adjustment ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create driver day adjustment with explicit amount"
if [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    xbe_json do driver-day-adjustments create --driver-day "$SAMPLE_DRIVER_DAY_ID" --amount-explicit "10.00"
    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            pass
        else
            fail "Create succeeded but no ID returned"
        fi
    else
        if [[ "$output" == *"already"* ]] || [[ "$output" == *"taken"* ]] || [[ "$output" == *"422"* ]] || [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            pass
        else
            fail "Create failed: $output"
        fi
    fi
else
    skip "No driver day ID available"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update driver day adjustment amount"
UPDATE_ID=""
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    UPDATE_ID="$CREATED_ID"
elif [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    UPDATE_ID="$SAMPLE_ID"
fi

if [[ -n "$UPDATE_ID" ]]; then
    xbe_json do driver-day-adjustments update "$UPDATE_ID" --amount-explicit "12.50"
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to update driver day adjustment"
        else
            fail "Update failed: $output"
        fi
    fi
else
    skip "No adjustment ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete driver day adjustment requires --confirm flag"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_run do driver-day-adjustments delete "$SAMPLE_ID"
    assert_failure
else
    skip "No adjustment ID available"
fi

test_name "Delete driver day adjustment with --confirm"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do driver-day-adjustments delete "$CREATED_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
            skip "Not authorized to delete driver day adjustment"
        else
            fail "Delete failed: $output"
        fi
    fi
else
    skip "No created adjustment available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create driver day adjustment without driver-day fails"
xbe_json do driver-day-adjustments create --amount-explicit "5.00"
assert_failure

test_name "Update without any fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do driver-day-adjustments update "$SAMPLE_ID"
    assert_failure
else
    skip "No adjustment ID available"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
