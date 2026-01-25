#!/bin/bash
#
# XBE CLI Integration Tests: Trucker Shift Sets
#
# Tests list filters, show, and update operations for the
# trucker-shift-sets resource.
#
# COVERAGE: List filters + show + update + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

SAMPLE_ID=""
TRUCKER_ID=""
DRIVER_ID=""
DRIVER_NAME=""
BROKER_ID=""
TRAILER_ID=""
TRAILER_CLASSIFICATION_ID=""
TRACTOR_ID=""
START_ON=""
EARLIEST_START_AT=""
ORDERED_SHIFT_ID=""
SHIFT_COUNT=""
ODOMETER_START_VALUE=""
ODOMETER_END_VALUE=""
EXPLICIT_MOBILIZATION_BEFORE_MINUTES=""
EXPLICIT_PRE_TRIP_MINUTES=""
EXPLICIT_POST_TRIP_MINUTES=""
IS_CUSTOMER_AMOUNT_CONSTRAINT_ENABLED=""
IS_BROKER_AMOUNT_CONSTRAINT_ENABLED=""
IS_TIME_SHEET_ENABLED=""
ODOMETER_UNIT_OF_MEASURE_EXPLICIT=""
CAN_CURRENT_USER_EDIT=""
TRIP_ID=""
EXPLICIT_BROKER_AMOUNT_CONSTRAINT_ID=""
SKIP_ID_FILTERS=0

# Derived values for update tests
UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES=""
UPDATE_EXPLICIT_PRE_TRIP_MINUTES=""
UPDATE_EXPLICIT_POST_TRIP_MINUTES=""
UPDATE_ODOMETER_START_VALUE=""
UPDATE_ODOMETER_END_VALUE=""
UPDATE_ODOMETER_UNIT=""
UPDATE_IS_CUSTOMER_CONSTRAINT=""
UPDATE_IS_BROKER_CONSTRAINT=""
UPDATE_IS_TIME_SHEET_ENABLED=""

# List filters with booleans
HAS_DRIVER_FILTER=""


describe "Resource: trucker-shift-sets"

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List trucker shift sets"
xbe_json view trucker-shift-sets list --limit 5
assert_success

test_name "List trucker shift sets returns array"
xbe_json view trucker-shift-sets list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list trucker shift sets"
fi

# ==========================================================================
# Sample Data
# ==========================================================================

test_name "Find sample trucker shift set"
xbe_json view trucker-shift-sets list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    START_ON=$(json_get ".[0].start_on")

    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        SKIP_ID_FILTERS=1
        skip "No trucker shift sets available"
    fi
else
    SKIP_ID_FILTERS=1
    fail "Failed to list trucker shift sets"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show trucker shift set"
if [[ $SKIP_ID_FILTERS -eq 0 && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets show "$SAMPLE_ID"
    if [[ $status -eq 0 ]]; then
        TRUCKER_ID=$(json_get ".trucker_id")
        DRIVER_ID=$(json_get ".driver_id")
        DRIVER_NAME=$(json_get ".driver_name")
        BROKER_ID=$(json_get ".broker_id")
        TRAILER_ID=$(json_get ".trailer_id")
        TRAILER_CLASSIFICATION_ID=$(json_get ".trailer_classification_id")
        TRACTOR_ID=$(json_get ".tractor_id")
        START_ON=$(json_get ".start_on")
        EARLIEST_START_AT=$(json_get ".earliest_start_at")
        ORDERED_SHIFT_ID=$(json_get ".ordered_shift_ids[0]")
        SHIFT_COUNT=$(json_get ".ordered_shift_ids | length")
        ODOMETER_START_VALUE=$(json_get ".odometer_start_value")
        ODOMETER_END_VALUE=$(json_get ".odometer_end_value")
        EXPLICIT_MOBILIZATION_BEFORE_MINUTES=$(json_get ".explicit_mobilization_before_minutes")
        EXPLICIT_PRE_TRIP_MINUTES=$(json_get ".explicit_pre_trip_minutes")
        EXPLICIT_POST_TRIP_MINUTES=$(json_get ".explicit_post_trip_minutes")
        IS_CUSTOMER_AMOUNT_CONSTRAINT_ENABLED=$(json_get ".is_customer_amount_constraint_enabled")
        IS_BROKER_AMOUNT_CONSTRAINT_ENABLED=$(json_get ".is_broker_amount_constraint_enabled")
        IS_TIME_SHEET_ENABLED=$(json_get ".is_time_sheet_enabled")
        ODOMETER_UNIT_OF_MEASURE_EXPLICIT=$(json_get ".odometer_unit_of_measure_explicit")
        CAN_CURRENT_USER_EDIT=$(json_get ".can_current_user_edit")
        TRIP_ID=$(json_get ".trip_ids[0]")
        EXPLICIT_BROKER_AMOUNT_CONSTRAINT_ID=$(json_get ".explicit_broker_amount_constraint_id")
        pass
    else
        fail "Failed to show trucker shift set"
    fi
else
    skip "No trucker shift set ID available"
fi

if [[ "$CAN_CURRENT_USER_EDIT" == "true" ]]; then
    UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES="$EXPLICIT_MOBILIZATION_BEFORE_MINUTES"
    UPDATE_EXPLICIT_PRE_TRIP_MINUTES="$EXPLICIT_PRE_TRIP_MINUTES"
    UPDATE_EXPLICIT_POST_TRIP_MINUTES="$EXPLICIT_POST_TRIP_MINUTES"
    UPDATE_ODOMETER_START_VALUE="$ODOMETER_START_VALUE"
    UPDATE_ODOMETER_END_VALUE="$ODOMETER_END_VALUE"
    UPDATE_ODOMETER_UNIT="$ODOMETER_UNIT_OF_MEASURE_EXPLICIT"
    UPDATE_IS_CUSTOMER_CONSTRAINT="$IS_CUSTOMER_AMOUNT_CONSTRAINT_ENABLED"
    UPDATE_IS_BROKER_CONSTRAINT="$IS_BROKER_AMOUNT_CONSTRAINT_ENABLED"
    UPDATE_IS_TIME_SHEET_ENABLED="$IS_TIME_SHEET_ENABLED"

    if [[ -z "$UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES" || "$UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES" == "null" ]]; then
        UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES=0
    fi
    if [[ -z "$UPDATE_EXPLICIT_PRE_TRIP_MINUTES" || "$UPDATE_EXPLICIT_PRE_TRIP_MINUTES" == "null" ]]; then
        UPDATE_EXPLICIT_PRE_TRIP_MINUTES=5
    fi
    if [[ -z "$UPDATE_EXPLICIT_POST_TRIP_MINUTES" || "$UPDATE_EXPLICIT_POST_TRIP_MINUTES" == "null" ]]; then
        UPDATE_EXPLICIT_POST_TRIP_MINUTES=5
    fi
    if [[ -z "$UPDATE_ODOMETER_START_VALUE" || "$UPDATE_ODOMETER_START_VALUE" == "null" ]]; then
        UPDATE_ODOMETER_START_VALUE=100
    fi
    if [[ -z "$UPDATE_ODOMETER_END_VALUE" || "$UPDATE_ODOMETER_END_VALUE" == "null" ]]; then
        UPDATE_ODOMETER_END_VALUE=101
    fi
    if [[ -z "$UPDATE_ODOMETER_UNIT" || "$UPDATE_ODOMETER_UNIT" == "null" ]]; then
        UPDATE_ODOMETER_UNIT="mile"
    fi
    if [[ -z "$UPDATE_IS_CUSTOMER_CONSTRAINT" || "$UPDATE_IS_CUSTOMER_CONSTRAINT" == "null" ]]; then
        UPDATE_IS_CUSTOMER_CONSTRAINT=true
    fi
    if [[ -z "$UPDATE_IS_BROKER_CONSTRAINT" || "$UPDATE_IS_BROKER_CONSTRAINT" == "null" ]]; then
        UPDATE_IS_BROKER_CONSTRAINT=true
    fi
    if [[ -z "$UPDATE_IS_TIME_SHEET_ENABLED" || "$UPDATE_IS_TIME_SHEET_ENABLED" == "null" ]]; then
        UPDATE_IS_TIME_SHEET_ENABLED=true
    fi
fi

if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
    HAS_DRIVER_FILTER=true
else
    HAS_DRIVER_FILTER=false
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List trucker shift sets with --trucker filter"
if [[ -n "$TRUCKER_ID" && "$TRUCKER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --trucker "$TRUCKER_ID" --limit 5
    assert_success
else
    skip "No trucker ID available"
fi

test_name "List trucker shift sets with --driver filter"
if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --driver "$DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List trucker shift sets with --driver-id filter"
if [[ -n "$DRIVER_ID" && "$DRIVER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --driver-id "$DRIVER_ID" --limit 5
    assert_success
else
    skip "No driver ID available"
fi

test_name "List trucker shift sets with --driver-name filter"
if [[ -n "$DRIVER_NAME" && "$DRIVER_NAME" != "null" ]]; then
    xbe_json view trucker-shift-sets list --driver-name "$DRIVER_NAME" --limit 5
    assert_success
else
    skip "No driver name available"
fi

test_name "List trucker shift sets with --has-driver filter"
xbe_json view trucker-shift-sets list --has-driver "$HAS_DRIVER_FILTER" --limit 5
assert_success

test_name "List trucker shift sets with --broker filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --broker "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List trucker shift sets with --broker-id filter"
if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --broker-id "$BROKER_ID" --limit 5
    assert_success
else
    skip "No broker ID available"
fi

test_name "List trucker shift sets with --trailer filter"
if [[ -n "$TRAILER_ID" && "$TRAILER_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --trailer "$TRAILER_ID" --limit 5
    assert_success
else
    skip "No trailer ID available"
fi

test_name "List trucker shift sets with --trailer-classification filter"
if [[ -n "$TRAILER_CLASSIFICATION_ID" && "$TRAILER_CLASSIFICATION_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --trailer-classification "$TRAILER_CLASSIFICATION_ID" --limit 5
    assert_success
else
    skip "No trailer classification ID available"
fi

test_name "List trucker shift sets with --tractor filter"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --tractor "$TRACTOR_ID" --limit 5
    assert_success
else
    skip "No tractor ID available"
fi

test_name "List trucker shift sets with --shifts filter"
if [[ -n "$ORDERED_SHIFT_ID" && "$ORDERED_SHIFT_ID" != "null" ]]; then
    xbe_json view trucker-shift-sets list --shifts "$ORDERED_SHIFT_ID" --limit 5
    assert_success
else
    skip "No shift ID available"
fi

test_name "List trucker shift sets with --start-on filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view trucker-shift-sets list --start-on "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

test_name "List trucker shift sets with --start-on-min filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view trucker-shift-sets list --start-on-min "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

test_name "List trucker shift sets with --start-on-max filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view trucker-shift-sets list --start-on-max "$START_ON" --limit 5
    assert_success
else
    skip "No start-on date available"
fi

test_name "List trucker shift sets with --has-start-on filter"
xbe_json view trucker-shift-sets list --has-start-on true --limit 5
assert_success

test_name "List trucker shift sets with --start-at filter"
if [[ -n "$EARLIEST_START_AT" && "$EARLIEST_START_AT" != "null" ]]; then
    xbe_json view trucker-shift-sets list --start-at "$EARLIEST_START_AT" --limit 5
    assert_success
else
    skip "No earliest start-at available"
fi

test_name "List trucker shift sets with --driver-day-on filter"
if [[ -n "$START_ON" && "$START_ON" != "null" ]]; then
    xbe_json view trucker-shift-sets list --driver-day-on "$START_ON" --limit 5
    assert_success
else
    skip "No driver day date available"
fi

test_name "List trucker shift sets with --number-of-shifts-eq filter"
if [[ -n "$SHIFT_COUNT" && "$SHIFT_COUNT" != "null" && "$SHIFT_COUNT" -gt 0 ]]; then
    xbe_json view trucker-shift-sets list --number-of-shifts-eq "$SHIFT_COUNT" --limit 5
    assert_success
else
    skip "No shift count available"
fi

test_name "List trucker shift sets with --number-of-shifts-gte filter"
if [[ -n "$SHIFT_COUNT" && "$SHIFT_COUNT" != "null" && "$SHIFT_COUNT" -gt 0 ]]; then
    xbe_json view trucker-shift-sets list --number-of-shifts-gte "$SHIFT_COUNT" --limit 5
    assert_success
else
    skip "No shift count available"
fi

test_name "List trucker shift sets with --is-expecting-time-sheet filter"
xbe_json view trucker-shift-sets list --is-expecting-time-sheet true --limit 5
assert_success

test_name "List trucker shift sets with --without-time-card filter"
xbe_json view trucker-shift-sets list --without-time-card true --limit 5
assert_success

test_name "List trucker shift sets with --without-approved-time-card filter"
xbe_json view trucker-shift-sets list --without-approved-time-card true --limit 5
assert_success

test_name "List trucker shift sets with --with-missing-time-card-approvals filter"
xbe_json view trucker-shift-sets list --with-missing-time-card-approvals true --limit 5
assert_success

test_name "List trucker shift sets with --without-approved-time-sheet filter"
xbe_json view trucker-shift-sets list --without-approved-time-sheet true --limit 5
assert_success

test_name "List trucker shift sets with --without-submitted-time-sheet filter"
xbe_json view trucker-shift-sets list --without-submitted-time-sheet true --limit 5
assert_success

test_name "List trucker shift sets with --has-constraint filter"
xbe_json view trucker-shift-sets list --has-constraint true --limit 5
assert_success

test_name "List trucker shift sets with --odometer-start-value filter"
if [[ -n "$ODOMETER_START_VALUE" && "$ODOMETER_START_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-start-value "$ODOMETER_START_VALUE" --limit 5
    assert_success
else
    skip "No odometer start value available"
fi

test_name "List trucker shift sets with --odometer-start-value-min filter"
if [[ -n "$ODOMETER_START_VALUE" && "$ODOMETER_START_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-start-value-min "$ODOMETER_START_VALUE" --limit 5
    assert_success
else
    skip "No odometer start value available"
fi

test_name "List trucker shift sets with --odometer-start-value-max filter"
if [[ -n "$ODOMETER_START_VALUE" && "$ODOMETER_START_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-start-value-max "$ODOMETER_START_VALUE" --limit 5
    assert_success
else
    skip "No odometer start value available"
fi

test_name "List trucker shift sets with --odometer-end-value filter"
if [[ -n "$ODOMETER_END_VALUE" && "$ODOMETER_END_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-end-value "$ODOMETER_END_VALUE" --limit 5
    assert_success
else
    skip "No odometer end value available"
fi

test_name "List trucker shift sets with --odometer-end-value-min filter"
if [[ -n "$ODOMETER_END_VALUE" && "$ODOMETER_END_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-end-value-min "$ODOMETER_END_VALUE" --limit 5
    assert_success
else
    skip "No odometer end value available"
fi

test_name "List trucker shift sets with --odometer-end-value-max filter"
if [[ -n "$ODOMETER_END_VALUE" && "$ODOMETER_END_VALUE" != "null" ]]; then
    xbe_json view trucker-shift-sets list --odometer-end-value-max "$ODOMETER_END_VALUE" --limit 5
    assert_success
else
    skip "No odometer end value available"
fi

# ==========================================================================
# LIST Tests - Pagination
# ==========================================================================

test_name "List trucker shift sets with --limit"
xbe_json view trucker-shift-sets list --limit 3
assert_success

test_name "List trucker shift sets with --offset"
xbe_json view trucker-shift-sets list --limit 3 --offset 1
assert_success

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update trucker shift set explicit-mobilization-before-minutes"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --explicit-mobilization-before-minutes "$UPDATE_EXPLICIT_MOBILIZATION_BEFORE_MINUTES"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set explicit-pre-trip-minutes"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --explicit-pre-trip-minutes "$UPDATE_EXPLICIT_PRE_TRIP_MINUTES"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set explicit-post-trip-minutes"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --explicit-post-trip-minutes "$UPDATE_EXPLICIT_POST_TRIP_MINUTES"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set is-customer-amount-constraint-enabled"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --is-customer-amount-constraint-enabled="$UPDATE_IS_CUSTOMER_CONSTRAINT"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set is-broker-amount-constraint-enabled"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --is-broker-amount-constraint-enabled="$UPDATE_IS_BROKER_CONSTRAINT"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set is-time-sheet-enabled"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --is-time-sheet-enabled="$UPDATE_IS_TIME_SHEET_ENABLED"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set odometer-start-value"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --odometer-start-value "$UPDATE_ODOMETER_START_VALUE"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set odometer-end-value"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --odometer-end-value "$UPDATE_ODOMETER_END_VALUE"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set odometer-unit-of-measure-explicit"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --odometer-unit-of-measure-explicit "$UPDATE_ODOMETER_UNIT"
    assert_success
else
    skip "No editable trucker shift set available"
fi

test_name "Update trucker shift set new-shift-ids"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$ORDERED_SHIFT_ID" && "$ORDERED_SHIFT_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --new-shift-ids "$ORDERED_SHIFT_ID"
    assert_success
else
    skip "No editable trucker shift set with shift IDs available"
fi

test_name "Update trucker shift set trips"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$TRIP_ID" && "$TRIP_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --trips "$TRIP_ID"
    assert_success
else
    skip "No editable trucker shift set with trips available"
fi

test_name "Update trucker shift set explicit-broker-amount-constraint"
if [[ "$CAN_CURRENT_USER_EDIT" == "true" && -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" && -n "$EXPLICIT_BROKER_AMOUNT_CONSTRAINT_ID" && "$EXPLICIT_BROKER_AMOUNT_CONSTRAINT_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID" --explicit-broker-amount-constraint "$EXPLICIT_BROKER_AMOUNT_CONSTRAINT_ID"
    assert_success
else
    skip "No editable trucker shift set with explicit broker amount constraint available"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Update without any fields fails"
if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
    xbe_json do trucker-shift-sets update "$SAMPLE_ID"
    assert_failure
else
    skip "No trucker shift set ID available"
fi

test_name "Update non-existent trucker shift set fails"
xbe_json do trucker-shift-sets update "99999999" --is-time-sheet-enabled=true
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
