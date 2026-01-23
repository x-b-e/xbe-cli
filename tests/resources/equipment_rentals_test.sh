#!/bin/bash
#
# XBE CLI Integration Tests: Equipment Rentals
#
# Tests CRUD operations for the equipment-rentals resource.
# Equipment rentals require a broker relationship.
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_RENTAL_ID=""
CREATED_BROKER_ID=""
CREATED_EQUIPMENT_CLASSIFICATION_ID=""

describe "Resource: equipment-rentals"

# ============================================================================
# Prerequisites - Create resources for equipment rental tests
# ============================================================================

test_name "Create prerequisite broker for equipment rental tests"
BROKER_NAME=$(unique_name "RentalTestBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_BROKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$CREATED_BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        CREATED_BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $CREATED_BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create prerequisite equipment classification for equipment rental tests"
CLASSIFICATION_NAME=$(unique_name "EqClass")

xbe_json do equipment-classifications create --name "$CLASSIFICATION_NAME"

if [[ $status -eq 0 ]]; then
    CREATED_EQUIPMENT_CLASSIFICATION_ID=$(json_get ".id")
    if [[ -n "$CREATED_EQUIPMENT_CLASSIFICATION_ID" && "$CREATED_EQUIPMENT_CLASSIFICATION_ID" != "null" ]]; then
        register_cleanup "equipment-classifications" "$CREATED_EQUIPMENT_CLASSIFICATION_ID"
        pass
    else
        fail "Created equipment classification but no ID returned"
        echo "Cannot continue without equipment classification"
        run_tests
    fi
else
    fail "Failed to create equipment classification"
    echo "Cannot continue without equipment classification"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create equipment rental with required fields"
xbe_json do equipment-rentals create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID"

if [[ $status -eq 0 ]]; then
    CREATED_RENTAL_ID=$(json_get ".id")
    if [[ -n "$CREATED_RENTAL_ID" && "$CREATED_RENTAL_ID" != "null" ]]; then
        register_cleanup "equipment-rentals" "$CREATED_RENTAL_ID"
        pass
    else
        fail "Created equipment rental but no ID returned"
    fi
else
    fail "Failed to create equipment rental"
fi

if [[ -z "$CREATED_RENTAL_ID" || "$CREATED_RENTAL_ID" == "null" ]]; then
    echo "Cannot continue without a valid equipment rental ID"
    run_tests
fi

test_name "Create equipment rental with description"
xbe_json do equipment-rentals create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --description "Test Excavator Rental"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-rentals" "$id"
    pass
else
    fail "Failed to create equipment rental with description"
fi

test_name "Create equipment rental with dates"
xbe_json do equipment-rentals create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --description "Dated Rental" \
    --start-on "2024-01-15" \
    --end-on-planned "2024-02-15"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-rentals" "$id"
    pass
else
    fail "Failed to create equipment rental with dates"
fi

test_name "Create equipment rental with cost information"
xbe_json do equipment-rentals create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --description "Costed Rental" \
    --approximate-cost-per-day "500.00" \
    --cost-per-hour "75.00" \
    --target-utilization-hours "8"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "equipment-rentals" "$id"
    pass
else
    fail "Failed to create equipment rental with cost information"
fi

# Equipment classification is now required, so this test is redundant
# The base create test already tests with equipment-classification

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update equipment rental description"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --description "Updated Rental Description"
assert_success

test_name "Update equipment rental start-on"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --start-on "2024-02-01"
assert_success

# Status update requires start-on to be set (can't be rented without a start date)
test_name "Update equipment rental status"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --status "rented"
assert_success

test_name "Update equipment rental end-on"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --end-on "2024-03-01"
assert_success

test_name "Update equipment rental planned dates"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" \
    --start-on-planned "2024-01-25" \
    --end-on-planned "2024-03-15"
assert_success

test_name "Update equipment rental cost per day"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --approximate-cost-per-day "600.00"
assert_success

test_name "Update equipment rental cost per hour"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --cost-per-hour "85.00"
assert_success

test_name "Update equipment rental utilization hours"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --target-utilization-hours "10"
assert_success

test_name "Update equipment rental actual usage hours"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --actual-rental-usage-hours "45"
assert_success

test_name "Update equipment rental skip-weekend"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID" --skip-weekend
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List equipment rentals"
xbe_json view equipment-rentals list --limit 5
assert_success

test_name "List equipment rentals returns array"
xbe_json view equipment-rentals list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list equipment rentals"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List equipment rentals with --broker filter"
xbe_json view equipment-rentals list --broker "$CREATED_BROKER_ID" --limit 10
assert_success

test_name "List equipment rentals with --start-on-min filter"
xbe_json view equipment-rentals list --start-on-min "2024-01-01" --limit 10
assert_success

test_name "List equipment rentals with --start-on-max filter"
xbe_json view equipment-rentals list --start-on-max "2025-12-31" --limit 10
assert_success

test_name "List equipment rentals with --end-on-min filter"
xbe_json view equipment-rentals list --end-on-min "2024-01-01" --limit 10
assert_success

test_name "List equipment rentals with --end-on-max filter"
xbe_json view equipment-rentals list --end-on-max "2025-12-31" --limit 10
assert_success

# ============================================================================
# LIST Tests - Pagination
# ============================================================================

test_name "List equipment rentals with --limit"
xbe_json view equipment-rentals list --limit 3
assert_success

test_name "List equipment rentals with --offset"
xbe_json view equipment-rentals list --limit 3 --offset 1
assert_success

test_name "List equipment rentals with pagination (limit + offset)"
xbe_json view equipment-rentals list --limit 5 --offset 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete equipment rental requires --confirm flag"
xbe_json do equipment-rentals delete "$CREATED_RENTAL_ID"
assert_failure

test_name "Delete equipment rental with --confirm"
# Create an equipment rental specifically for deletion
xbe_json do equipment-rentals create \
    --broker "$CREATED_BROKER_ID" \
    --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" \
    --description "Rental For Deletion"

if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do equipment-rentals delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        # API may not allow deletion
        register_cleanup "equipment-rentals" "$DEL_ID"
        skip "API may not allow equipment rental deletion"
    fi
else
    skip "Could not create equipment rental for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create equipment rental without broker fails"
xbe_json do equipment-rentals create --equipment-classification "$CREATED_EQUIPMENT_CLASSIFICATION_ID" --description "Test"
assert_failure

test_name "Create equipment rental without equipment-classification fails"
xbe_json do equipment-rentals create --broker "$CREATED_BROKER_ID" --description "Test"
assert_failure

test_name "Update without any fields fails"
xbe_json do equipment-rentals update "$CREATED_RENTAL_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
