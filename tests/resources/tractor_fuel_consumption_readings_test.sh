#!/bin/bash
#
# XBE CLI Integration Tests: Tractor Fuel Consumption Readings
#
# Tests create, update, delete operations and list filters for the
# tractor_fuel_consumption_readings resource.
#
# COVERAGE: Create, update, delete + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_READING_ID=""
EXISTING_READING_ID=""

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRACTOR_ID=""

TRACTOR_ID="${XBE_TEST_TRACTOR_ID:-}"
UNIT_OF_MEASURE_ID="${XBE_TEST_UNIT_OF_MEASURE_ID:-}"
DRIVER_DAY_ID="${XBE_TEST_DRIVER_DAY_ID:-}"

GALLON_UOM_ID=""
LITER_UOM_ID=""
ALT_UNIT_OF_MEASURE_ID=""

describe "Resource: tractor-fuel-consumption-readings"

# ==========================================================================
# Seed IDs from existing data if available
# ==========================================================================

test_name "Lookup existing tractor fuel consumption reading (if any)"
xbe_json view tractor-fuel-consumption-readings list --limit 1
if [[ $status -eq 0 ]]; then
    EXISTING_READING_ID=$(json_get ".[0].id")
    if [[ -n "$EXISTING_READING_ID" && "$EXISTING_READING_ID" != "null" ]]; then
        if [[ -z "$TRACTOR_ID" || "$TRACTOR_ID" == "null" ]]; then
            TRACTOR_ID=$(json_get ".[0].tractor_id")
        fi
        if [[ -z "$UNIT_OF_MEASURE_ID" || "$UNIT_OF_MEASURE_ID" == "null" ]]; then
            UNIT_OF_MEASURE_ID=$(json_get ".[0].unit_of_measure_id")
        fi
        if [[ -z "$DRIVER_DAY_ID" || "$DRIVER_DAY_ID" == "null" ]]; then
            DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
        fi
    fi
    pass
else
    fail "Failed to list tractor fuel consumption readings"
fi

# ==========================================================================
# Prerequisites - Create broker, trucker, tractor (if needed)
# ==========================================================================

if [[ -z "$TRACTOR_ID" || "$TRACTOR_ID" == "null" ]]; then
    test_name "Create prerequisite broker for tractor fuel consumption tests"
    BROKER_NAME=$(unique_name "FuelReadBroker")

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

    test_name "Create prerequisite trucker"
    TRUCKER_NAME=$(unique_name "FuelReadTrucker")
    TRUCKER_ADDRESS="350 N Orleans St, Chicago, IL 60654"

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

    test_name "Create prerequisite tractor"
    TRACTOR_NUMBER=$(unique_name "FuelReadTractor")

    xbe_json do tractors create \
        --number "$TRACTOR_NUMBER" \
        --trucker "$CREATED_TRUCKER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_TRACTOR_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRACTOR_ID" && "$CREATED_TRACTOR_ID" != "null" ]]; then
            register_cleanup "tractors" "$CREATED_TRACTOR_ID"
            TRACTOR_ID="$CREATED_TRACTOR_ID"
            pass
        else
            fail "Created tractor but no ID returned"
            echo "Cannot continue without a tractor"
            run_tests
        fi
    else
        fail "Failed to create tractor"
        echo "Cannot continue without a tractor"
        run_tests
    fi
fi

# ==========================================================================
# Lookup unit of measure IDs
# ==========================================================================

test_name "Lookup unit of measure IDs (gallon/liter)"
if [[ -z "$UNIT_OF_MEASURE_ID" || "$UNIT_OF_MEASURE_ID" == "null" ]]; then
    xbe_json view unit-of-measures list --name "gallon" --limit 1
    if [[ $status -eq 0 ]]; then
        GALLON_UOM_ID=$(json_get ".[0].id")
    fi
    xbe_json view unit-of-measures list --name "liter" --limit 1
    if [[ $status -eq 0 ]]; then
        LITER_UOM_ID=$(json_get ".[0].id")
    fi

    if [[ -n "$GALLON_UOM_ID" && "$GALLON_UOM_ID" != "null" ]]; then
        UNIT_OF_MEASURE_ID="$GALLON_UOM_ID"
        if [[ -n "$LITER_UOM_ID" && "$LITER_UOM_ID" != "null" ]]; then
            ALT_UNIT_OF_MEASURE_ID="$LITER_UOM_ID"
        fi
    elif [[ -n "$LITER_UOM_ID" && "$LITER_UOM_ID" != "null" ]]; then
        UNIT_OF_MEASURE_ID="$LITER_UOM_ID"
    fi
fi

if [[ -n "$GALLON_UOM_ID" && "$GALLON_UOM_ID" != "null" && -n "$LITER_UOM_ID" && "$LITER_UOM_ID" != "null" && "$GALLON_UOM_ID" != "$LITER_UOM_ID" ]]; then
    if [[ "$UNIT_OF_MEASURE_ID" == "$GALLON_UOM_ID" ]]; then
        ALT_UNIT_OF_MEASURE_ID="$LITER_UOM_ID"
    else
        ALT_UNIT_OF_MEASURE_ID="$GALLON_UOM_ID"
    fi
fi

if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    pass
else
    fail "No unit of measure ID available"
fi

# ==========================================================================
# Optional driver day lookup
# ==========================================================================

test_name "Lookup driver day (optional)"
if [[ -z "$DRIVER_DAY_ID" || "$DRIVER_DAY_ID" == "null" ]]; then
    xbe_json view trips list --limit 1
    if [[ $status -eq 0 ]]; then
        DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
    fi
fi

if [[ -n "$DRIVER_DAY_ID" && "$DRIVER_DAY_ID" != "null" ]]; then
    pass
else
    skip "No driver day ID available"
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create tractor fuel consumption reading"
if [[ -n "$TRACTOR_ID" && "$TRACTOR_ID" != "null" && -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    READING_ON="$(date -u +%Y-%m-%d)"
    READING_TIME="$(date -u +%H:%M:%S)"
    CREATE_ARGS=(--tractor "$TRACTOR_ID" --unit-of-measure "$UNIT_OF_MEASURE_ID" --value "12.5" --reading-on "$READING_ON" --reading-time "$READING_TIME" --state-code "CA" --date-sequence "1")
    if [[ -n "$DRIVER_DAY_ID" && "$DRIVER_DAY_ID" != "null" ]]; then
        CREATE_ARGS+=(--driver-day "$DRIVER_DAY_ID")
    fi

    xbe_json do tractor-fuel-consumption-readings create "${CREATE_ARGS[@]}"

    if [[ $status -eq 0 ]]; then
        CREATED_READING_ID=$(json_get ".id")
        if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
            register_cleanup "tractor-fuel-consumption-readings" "$CREATED_READING_ID"
            pass
        else
            fail "Created reading but no ID returned"
        fi
    else
        fail "Failed to create tractor fuel consumption reading: $output"
    fi
else
    skip "Missing tractor/unit-of-measure IDs"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show tractor fuel consumption reading"
SHOW_ID="$CREATED_READING_ID"
if [[ -z "$SHOW_ID" || "$SHOW_ID" == "null" ]]; then
    SHOW_ID="$EXISTING_READING_ID"
fi
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view tractor-fuel-consumption-readings show "$SHOW_ID"
    assert_success
else
    skip "No tractor fuel consumption reading ID available"
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update tractor fuel consumption reading"
UPDATE_ID="$CREATED_READING_ID"
if [[ -z "$UPDATE_ID" || "$UPDATE_ID" == "null" ]]; then
    UPDATE_ID="$EXISTING_READING_ID"
fi
if [[ -n "$UPDATE_ID" && "$UPDATE_ID" != "null" ]]; then
    UPDATE_ARGS=(--value "9.75" --state-code "NV")
    if [[ -n "$ALT_UNIT_OF_MEASURE_ID" && "$ALT_UNIT_OF_MEASURE_ID" != "null" ]]; then
        UPDATE_ARGS+=(--unit-of-measure "$ALT_UNIT_OF_MEASURE_ID")
    fi
    xbe_json do tractor-fuel-consumption-readings update "$UPDATE_ID" "${UPDATE_ARGS[@]}"
    assert_success
else
    skip "No tractor fuel consumption reading ID available"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List tractor fuel consumption readings"
xbe_json view tractor-fuel-consumption-readings list --limit 10
assert_success

test_name "List tractor fuel consumption readings returns array"
xbe_json view tractor-fuel-consumption-readings list --limit 10
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tractor fuel consumption readings"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List tractor fuel consumption readings with --tractor filter"
xbe_json view tractor-fuel-consumption-readings list --tractor 1 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --trucker filter"
xbe_json view tractor-fuel-consumption-readings list --trucker 1 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --driver-day filter"
xbe_json view tractor-fuel-consumption-readings list --driver-day 1 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --unit-of-measure filter"
xbe_json view tractor-fuel-consumption-readings list --unit-of-measure 1 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --reading-on filter"
xbe_json view tractor-fuel-consumption-readings list --reading-on 2025-01-01 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --reading-on-min filter"
xbe_json view tractor-fuel-consumption-readings list --reading-on-min 2025-01-01 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --reading-on-max filter"
xbe_json view tractor-fuel-consumption-readings list --reading-on-max 2025-01-31 --limit 10
assert_success

test_name "List tractor fuel consumption readings with --has-reading-on filter"
xbe_json view tractor-fuel-consumption-readings list --has-reading-on true --limit 10
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete tractor fuel consumption reading requires --confirm flag"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_json do tractor-fuel-consumption-readings delete "$CREATED_READING_ID"
    assert_failure
else
    skip "No reading ID available"
fi

test_name "Delete tractor fuel consumption reading with --confirm"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_json do tractor-fuel-consumption-readings delete "$CREATED_READING_ID" --confirm
    assert_success
else
    skip "No reading ID available"
fi

# ==========================================================================
# Summary
# ==========================================================================

run_tests
