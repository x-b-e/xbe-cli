#!/bin/bash
#
# XBE CLI Integration Tests: Tractor Odometer Readings
#
# Tests list, show, create, update, and delete operations for the tractor-odometer-readings resource.
#
# COVERAGE: All list filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_TRUCKER_ID=""
CREATED_TRACTOR_ID=""
CREATED_READING_ID=""
UNIT_OF_MEASURE_ID=""
ALT_UNIT_OF_MEASURE_ID=""

SAMPLE_ID=""
SAMPLE_DRIVER_DAY_ID=""
SAMPLE_CREATED_BY_ID=""

READING_ON="2025-01-15"
READING_TIME="08:30"
READING_VALUE="120345.6"
STATE_CODE="IL"
UPDATED_STATE_CODE="CA"
UPDATED_READING_ON="2025-01-16"
UPDATED_READING_TIME="09:15"
UPDATED_VALUE="120400.1"


describe "Resource: tractor-odometer-readings"

# ============================================================================
# Prerequisites - Create broker, trucker, tractor, and pick unit of measure
# ============================================================================

test_name "Create prerequisite broker for tractor odometer readings tests"
BROKER_NAME=$(unique_name "TORTestBroker")

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
TRUCKER_NAME=$(unique_name "TORTestTrucker")
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
TRACTOR_NUMBER=$(unique_name "TORTractor")

xbe_json do tractors create \
    --number "$TRACTOR_NUMBER" \
    --trucker "$CREATED_TRUCKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_TRACTOR_ID=$(json_get ".id")
    if [[ -n "$CREATED_TRACTOR_ID" && "$CREATED_TRACTOR_ID" != "null" ]]; then
        register_cleanup "tractors" "$CREATED_TRACTOR_ID"
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

test_name "Find unit of measure for odometer readings"

xbe_json view unit-of-measures list --limit 200
if [[ $status -eq 0 ]]; then
    UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r '[.[] | select((.name | ascii_downcase) == "mile" or (.name | ascii_downcase) == "miles" or (.name | ascii_downcase) == "linear meter" or (.name | ascii_downcase) == "linear meters")][0].id')
    if [[ -z "$UNIT_OF_MEASURE_ID" || "$UNIT_OF_MEASURE_ID" == "null" ]]; then
        UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r '[.[] | select((.abbreviation | ascii_downcase) == "mi")][0].id')
    fi
    if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
        ALT_UNIT_OF_MEASURE_ID=$(echo "$output" | jq -r --arg primary "$UNIT_OF_MEASURE_ID" '[.[] | select((.name | ascii_downcase) == "mile" or (.name | ascii_downcase) == "miles" or (.name | ascii_downcase) == "linear meter" or (.name | ascii_downcase) == "linear meters") | select(.id != $primary)][0].id')
    fi
fi

if [[ -z "$UNIT_OF_MEASURE_ID" || "$UNIT_OF_MEASURE_ID" == "null" ]]; then
    fail "Could not find a unit of measure ID"
    echo "Cannot continue without a unit of measure"
    run_tests
else
    pass
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create tractor odometer reading with required fields"

xbe_json do tractor-odometer-readings create \
    --tractor "$CREATED_TRACTOR_ID" \
    --unit-of-measure "$UNIT_OF_MEASURE_ID" \
    --value "$READING_VALUE" \
    --state-code "$STATE_CODE" \
    --reading-on "$READING_ON" \
    --reading-time "$READING_TIME" \
    --date-sequence 1

if [[ $status -eq 0 ]]; then
    CREATED_READING_ID=$(json_get ".id")
    if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
        register_cleanup "tractor-odometer-readings" "$CREATED_READING_ID"
        pass
    else
        fail "Created reading but no ID returned"
    fi
else
    fail "Failed to create tractor odometer reading"
fi

# Only continue if we successfully created a reading
if [[ -z "$CREATED_READING_ID" || "$CREATED_READING_ID" == "null" ]]; then
    echo "Cannot continue without a valid tractor odometer reading ID"
    run_tests
fi

# ==========================================================================
# UPDATE Tests
# ==========================================================================

test_name "Update tractor odometer reading value"
xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --value "$UPDATED_VALUE"
assert_success

test_name "Update tractor odometer reading state code"
xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --state-code "$UPDATED_STATE_CODE"
assert_success

test_name "Update tractor odometer reading date"
xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --reading-on "$UPDATED_READING_ON"
assert_success

test_name "Update tractor odometer reading time"
xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --reading-time "$UPDATED_READING_TIME"
assert_success

test_name "Update tractor odometer reading date sequence"
xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --date-sequence 2
assert_success

test_name "Update tractor odometer reading unit of measure"
if [[ -n "$ALT_UNIT_OF_MEASURE_ID" && "$ALT_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do tractor-odometer-readings update "$CREATED_READING_ID" --unit-of-measure "$ALT_UNIT_OF_MEASURE_ID"
    assert_success
else
    skip "No alternate unit of measure available"
fi

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List tractor odometer readings"
xbe_json view tractor-odometer-readings list --limit 5
assert_success

test_name "List tractor odometer readings returns array"
xbe_json view tractor-odometer-readings list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list tractor odometer readings"
fi

# ==========================================================================
# Sample Record (used for filters/show)
# ==========================================================================

test_name "Capture sample tractor odometer reading"
xbe_json view tractor-odometer-readings list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_DRIVER_DAY_ID=$(json_get ".[0].driver_day_id")
    SAMPLE_CREATED_BY_ID=$(json_get ".[0].created_by_id")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No tractor odometer readings available for follow-on tests"
    fi
else
    skip "Could not list tractor odometer readings to capture sample"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List tractor odometer readings with --tractor filter"
xbe_json view tractor-odometer-readings list --tractor "$CREATED_TRACTOR_ID" --limit 5
assert_success

test_name "List tractor odometer readings with --trucker filter"
xbe_json view tractor-odometer-readings list --trucker "$CREATED_TRUCKER_ID" --limit 5
assert_success

test_name "List tractor odometer readings with --unit-of-measure filter"
xbe_json view tractor-odometer-readings list --unit-of-measure "$UNIT_OF_MEASURE_ID" --limit 5
assert_success

test_name "List tractor odometer readings with --reading-on-min filter"
xbe_json view tractor-odometer-readings list --reading-on-min "2020-01-01" --limit 5
assert_success

test_name "List tractor odometer readings with --reading-on-max filter"
xbe_json view tractor-odometer-readings list --reading-on-max "2030-01-01" --limit 5
assert_success

test_name "List tractor odometer readings with --driver-day filter"
if [[ -n "$SAMPLE_DRIVER_DAY_ID" && "$SAMPLE_DRIVER_DAY_ID" != "null" ]]; then
    xbe_json view tractor-odometer-readings list --driver-day "$SAMPLE_DRIVER_DAY_ID" --limit 5
    assert_success
else
    skip "No driver day ID available"
fi

test_name "List tractor odometer readings with --created-by filter"
if [[ -n "$SAMPLE_CREATED_BY_ID" && "$SAMPLE_CREATED_BY_ID" != "null" ]]; then
    xbe_json view tractor-odometer-readings list --created-by "$SAMPLE_CREATED_BY_ID" --limit 5
    assert_success
else
    skip "No created-by ID available"
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show tractor odometer reading"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_json view tractor-odometer-readings show "$CREATED_READING_ID"
    assert_success
else
    skip "No tractor odometer reading ID available"
fi

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete tractor odometer reading"
if [[ -n "$CREATED_READING_ID" && "$CREATED_READING_ID" != "null" ]]; then
    xbe_run do tractor-odometer-readings delete "$CREATED_READING_ID" --confirm
    assert_success
else
    skip "No created tractor odometer reading to delete"
fi

# ==========================================================================
# Error Cases
# ==========================================================================

test_name "Create tractor odometer reading without required fields fails"
xbe_run do tractor-odometer-readings create
assert_failure


test_name "Update tractor odometer reading without any fields fails"
xbe_run do tractor-odometer-readings update "999999"
assert_failure

# ==========================================================================
# Summary
# ==========================================================================

run_tests
