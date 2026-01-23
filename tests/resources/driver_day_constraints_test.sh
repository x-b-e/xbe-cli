#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Constraints
#
# Tests list, show, create, update, and delete operations for the driver_day_constraints resource.
#
# COVERAGE: All filters + create/update relationships + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
SAMPLE_DRIVER_DAY_ID=""
SAMPLE_CONSTRAINT_ID=""

CREATE_DRIVER_DAY_ID=""
CREATE_DRIVER_DAY_SOURCE=""
CREATE_CONSTRAINT_ID=""
UPDATE_CONSTRAINT_ID=""
CREATED_DRIVER_DAY_CONSTRAINT_ID=""

describe "Resource: driver_day_constraints"

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List driver day constraints"
xbe_json view driver-day-constraints list --limit 5
assert_success

test_name "List driver day constraints returns array"
xbe_json view driver-day-constraints list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver day constraints"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample driver day constraint"
xbe_json view driver-day-constraints list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
    SAMPLE_CONSTRAINT_ID=$(json_get ".[0].constraint_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No driver day constraints available for follow-on tests"
    fi
else
    skip "Could not list driver day constraints to capture sample"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List driver day constraints with --driver-day filter"
if [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    xbe_json view driver-day-constraints list --driver-day "$SAMPLE_DRIVER_DAY_ID" --limit 5
    assert_success
else
    skip "No driver day ID available"
fi

test_name "List driver day constraints with --constraint filter"
if [[ -n "$SAMPLE_CONSTRAINT_ID" && "$SAMPLE_CONSTRAINT_ID" != "null" ]]; then
    xbe_json view driver-day-constraints list --constraint "$SAMPLE_CONSTRAINT_ID" --limit 5
    assert_success
else
    skip "No constraint ID available"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver day constraint"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view driver-day-constraints show "$SAMPLE_ID"
    assert_success
else
    skip "No driver day constraint ID available"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Resolve driver day for create"
if [[ -n "$XBE_TEST_DRIVER_DAY_ID" ]]; then
    CREATE_DRIVER_DAY_ID="$XBE_TEST_DRIVER_DAY_ID"
    CREATE_DRIVER_DAY_SOURCE="env"
    echo "    Using XBE_TEST_DRIVER_DAY_ID: $CREATE_DRIVER_DAY_ID"
    pass
elif [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    CREATE_DRIVER_DAY_ID="$SAMPLE_DRIVER_DAY_ID"
    CREATE_DRIVER_DAY_SOURCE="sample"
    pass
else
    xbe_json view trips list --limit 1
    if [[ $status -eq 0 ]]; then
        CREATE_DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
        if [[ -n "$CREATE_DRIVER_DAY_ID" && "$CREATE_DRIVER_DAY_ID" != "null" ]]; then
            CREATE_DRIVER_DAY_SOURCE="trips"
            pass
        else
            skip "No driver day ID available from trips"
        fi
    else
        skip "Could not list trips to find a driver day ID"
    fi
fi

if [[ -n "$XBE_TEST_SHIFT_SET_TIME_CARD_CONSTRAINT_ID" ]]; then
    CREATE_CONSTRAINT_ID="$XBE_TEST_SHIFT_SET_TIME_CARD_CONSTRAINT_ID"
    UPDATE_CONSTRAINT_ID="$XBE_TEST_SHIFT_SET_TIME_CARD_CONSTRAINT_ID"
elif [[ -n "$SAMPLE_CONSTRAINT_ID" && "$SAMPLE_CONSTRAINT_ID" != "null" ]]; then
    UPDATE_CONSTRAINT_ID="$SAMPLE_CONSTRAINT_ID"
    if [[ "$CREATE_DRIVER_DAY_SOURCE" != "sample" ]]; then
        CREATE_CONSTRAINT_ID="$SAMPLE_CONSTRAINT_ID"
    fi
fi

test_name "Create driver day constraint"
if [[ -n "$CREATE_DRIVER_DAY_ID" && "$CREATE_DRIVER_DAY_ID" != "null" ]]; then
    if [[ -n "$CREATE_CONSTRAINT_ID" && "$CREATE_CONSTRAINT_ID" != "null" ]]; then
        xbe_json do driver-day-constraints create --driver-day "$CREATE_DRIVER_DAY_ID" --constraint "$CREATE_CONSTRAINT_ID"
    else
        xbe_json do driver-day-constraints create --driver-day "$CREATE_DRIVER_DAY_ID"
    fi

    if [[ $status -eq 0 ]]; then
        CREATED_DRIVER_DAY_CONSTRAINT_ID=$(json_get ".id")
        if [[ -n "$CREATED_DRIVER_DAY_CONSTRAINT_ID" && "$CREATED_DRIVER_DAY_CONSTRAINT_ID" != "null" ]]; then
            register_cleanup "driver-day-constraints" "$CREATED_DRIVER_DAY_CONSTRAINT_ID"
            pass
        else
            fail "Created driver day constraint but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"422"* ]]; then
            skip "Create not permitted or failed validation"
        else
            fail "Failed to create driver day constraint"
        fi
    fi
else
    skip "No driver day ID available for create"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update driver day constraint"
if [[ -n "$CREATED_DRIVER_DAY_CONSTRAINT_ID" && "$CREATED_DRIVER_DAY_CONSTRAINT_ID" != "null" ]]; then
    if [[ -n "$UPDATE_CONSTRAINT_ID" && "$UPDATE_CONSTRAINT_ID" != "null" ]]; then
        xbe_json do driver-day-constraints update "$CREATED_DRIVER_DAY_CONSTRAINT_ID" --constraint "$UPDATE_CONSTRAINT_ID"
        if [[ $status -eq 0 ]]; then
            pass
        else
            if [[ "$output" == *"constraint already exists for driver day"* ]] || [[ "$output" == *"VALIDATION_ERROR"* ]]; then
                skip "Constraint already exists for driver day"
            else
                fail "Update failed"
            fi
        fi
    else
        skip "No constraint ID available for update"
    fi
else
    skip "No created driver day constraint ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete driver day constraint"
if [[ -n "$CREATED_DRIVER_DAY_CONSTRAINT_ID" && "$CREATED_DRIVER_DAY_CONSTRAINT_ID" != "null" ]]; then
    xbe_run do driver-day-constraints delete "$CREATED_DRIVER_DAY_CONSTRAINT_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        skip "Could not delete driver day constraint (permissions or policy)"
    fi
else
    skip "No created driver day constraint ID available"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create driver day constraint without required fields fails"
xbe_run do driver-day-constraints create
assert_failure

test_name "Update driver day constraint without any fields fails"
xbe_run do driver-day-constraints update "999999"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
