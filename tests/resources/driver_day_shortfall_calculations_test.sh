#!/bin/bash
#
# XBE CLI Integration Tests: Driver Day Shortfall Calculations
#
# Tests create behavior for the driver-day-shortfall-calculations resource.
# Requires time card IDs and constraint IDs available in staging.
#
# COVERAGE: Required attributes + optional unallocatable-time-card-ids
#

source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

describe "Resource: driver-day-shortfall-calculations"

TIME_CARD_IDS="${XBE_TEST_TIME_CARD_IDS:-}"
CONSTRAINT_IDS="${XBE_TEST_DRIVER_DAY_TIME_CARD_CONSTRAINT_IDS:-}"
UNALLOCATABLE_IDS="${XBE_TEST_UNALLOCATABLE_TIME_CARD_IDS:-}"

if [[ -z "$TIME_CARD_IDS" || -z "$CONSTRAINT_IDS" ]]; then
    test_name "Skip create tests (missing driver day shortfall prerequisites)"
    skip "Set XBE_TEST_TIME_CARD_IDS and XBE_TEST_DRIVER_DAY_TIME_CARD_CONSTRAINT_IDS to run create tests"
else
    test_name "Create driver day shortfall calculation with required IDs"
    xbe_json do driver-day-shortfall-calculations create \
        --time-card-ids "$TIME_CARD_IDS" \
        --driver-day-time-card-constraint-ids "$CONSTRAINT_IDS"

    if [[ $status -eq 0 ]]; then
        assert_json_has ".id"
    else
        fail "Failed to create driver day shortfall calculation"
    fi

    if [[ -n "$UNALLOCATABLE_IDS" ]]; then
        test_name "Create driver day shortfall calculation with unallocatable time card IDs"
        xbe_json do driver-day-shortfall-calculations create \
            --time-card-ids "$TIME_CARD_IDS" \
            --unallocatable-time-card-ids "$UNALLOCATABLE_IDS" \
            --driver-day-time-card-constraint-ids "$CONSTRAINT_IDS"

        if [[ $status -eq 0 ]]; then
            assert_json_has ".id"
        else
            fail "Failed to create driver day shortfall calculation with unallocatable IDs"
        fi
    fi
fi

test_name "Fail create without time-card-ids"
xbe_run do driver-day-shortfall-calculations create --driver-day-time-card-constraint-ids "1"
assert_failure

test_name "Fail create without driver-day-time-card-constraint-ids"
xbe_run do driver-day-shortfall-calculations create --time-card-ids "1"
assert_failure

run_tests
