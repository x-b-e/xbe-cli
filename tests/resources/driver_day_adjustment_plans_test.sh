#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Adjustment Plans
#
# Tests CRUD operations for the driver_day_adjustment_plans resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_PLAN_ID=""

TEST_START_AT="2025-01-15T08:00:00Z"
UPDATED_START_AT="2025-01-16T06:00:00Z"

describe "Resource: driver_day_adjustment_plans"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for driver day adjustment plan tests"
BROKER_NAME=$(unique_name "DriverDayPlanBroker")

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

test_name "Create trucker for driver day adjustment plan tests"
TRUCKER_NAME=$(unique_name "DriverDayPlanTrucker")
TRUCKER_ADDRESS="100 Plan Ave, Adjust City, AC 55555"

xbe_json do truckers create \
    --name "$TRUCKER_NAME" \
    --broker "$CREATED_BROKER_ID" \
    --company-address "$TRUCKER_ADDRESS"

if [[ $status -eq 0 ]]; then
    CREATED_TRUCKER_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRUCKER_ID" && "$CREATED_TRUCKER_ID" != "null" ]]; then
        register_cleanup "truckers" "$CREATED_TRUCKER_ID"
        pass
    else
        fail "Created trucker but no ID returned"
        echo "Cannot continue without a trucker"
        run_tests
    fi
else
    fail "Failed to create trucker"
    echo "Cannot continue without a trucker"
    run_tests
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create driver day adjustment plan with required fields"
PLAN_CONTENT=$(unique_name "DriverDayPlan")

xbe_json do driver-day-adjustment-plans create \
    --trucker "$CREATED_TRUCKER_ID" \
    --content "$PLAN_CONTENT" \
    --start-at "$TEST_START_AT"

if [[ $status -eq 0 ]]; then
    CREATED_PLAN_ID=$(json_get ".id")
    if [[ -n "$CREATED_PLAN_ID" && "$CREATED_PLAN_ID" != "null" ]]; then
        register_cleanup "driver-day-adjustment-plans" "$CREATED_PLAN_ID"
        pass
    else
        fail "Created plan but no ID returned"
    fi
else
    fail "Failed to create driver day adjustment plan"
fi

if [[ -z "$CREATED_PLAN_ID" || "$CREATED_PLAN_ID" == "null" ]]; then
    echo "Cannot continue without a valid plan ID"
    run_tests
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update driver day adjustment plan with no fields fails"

xbe_json do driver-day-adjustment-plans update "$CREATED_PLAN_ID"
assert_failure

test_name "Update driver day adjustment plan content and start-at"
UPDATED_CONTENT=$(unique_name "DriverDayPlanUpdated")

xbe_json do driver-day-adjustment-plans update "$CREATED_PLAN_ID" \
    --content "$UPDATED_CONTENT" \
    --start-at "$UPDATED_START_AT"

assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show driver day adjustment plan"

xbe_json view driver-day-adjustment-plans show "$CREATED_PLAN_ID"
assert_success

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List driver day adjustment plans filtered by trucker"

xbe_json view driver-day-adjustment-plans list --trucker "$CREATED_TRUCKER_ID"
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list driver day adjustment plans"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete driver day adjustment plan"

xbe_json do driver-day-adjustment-plans delete "$CREATED_PLAN_ID" --confirm
assert_success

# ============================================================================
# Summary
# ============================================================================

run_tests
