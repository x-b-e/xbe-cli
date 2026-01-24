#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Drivers
#
# Tests create, list filters, show, and delete operations for raw transport drivers.
#
# COVERAGE: Create attributes + list filters
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_DRIVER_ID=""
CREATED_BROKER_ID=""
IMPORT_STATUS=""

EXTERNAL_DRIVER_ID=""


describe "Resource: raw-transport-drivers"

# ==========================================================================
# Prerequisites - Create broker
# ==========================================================================

test_name "Create prerequisite broker for raw transport drivers"
BROKER_NAME=$(unique_name "RawTransportDriverBroker")

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

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create raw transport driver with required fields"
EXTERNAL_DRIVER_ID=$(unique_name "DRV")

xbe_json do raw-transport-drivers create \
    --broker "$CREATED_BROKER_ID" \
    --external-driver-id "$EXTERNAL_DRIVER_ID" \
    --importer "quantix_tmw" \
    --tables '[]'

if [[ $status -eq 0 ]]; then
    CREATED_DRIVER_ID=$(json_get ".id")
    IMPORT_STATUS=$(json_get ".import_status")
    if [[ -n "$CREATED_DRIVER_ID" && "$CREATED_DRIVER_ID" != "null" ]]; then
        register_cleanup "raw-transport-drivers" "$CREATED_DRIVER_ID"
        pass
    else
        fail "Created raw transport driver but no ID returned"
    fi
else
    fail "Failed to create raw transport driver"
fi

# Only continue if we successfully created a driver
if [[ -z "$CREATED_DRIVER_ID" || "$CREATED_DRIVER_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport driver ID"
    run_tests
fi

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show raw transport driver"
xbe_json view raw-transport-drivers show "$CREATED_DRIVER_ID"
assert_success

# ==========================================================================
# LIST Tests - Basic
# ==========================================================================

test_name "List raw transport drivers"
xbe_json view raw-transport-drivers list --limit 5
assert_success


test_name "List raw transport drivers returns array"
xbe_json view raw-transport-drivers list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw transport drivers"
fi

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "List raw transport drivers with --broker filter"
xbe_json view raw-transport-drivers list --broker "$CREATED_BROKER_ID" --limit 5
assert_success


test_name "List raw transport drivers with --importer filter"
xbe_json view raw-transport-drivers list --importer "quantix_tmw" --limit 5
assert_success


test_name "List raw transport drivers with --import-status filter"
STATUS_FILTER="$IMPORT_STATUS"
if [[ -z "$STATUS_FILTER" || "$STATUS_FILTER" == "null" ]]; then
    STATUS_FILTER="pending"
fi
xbe_json view raw-transport-drivers list --import-status "$STATUS_FILTER" --limit 5
assert_success

# ==========================================================================
# DELETE Tests
# ==========================================================================

test_name "Delete raw transport driver requires --confirm flag"
xbe_run do raw-transport-drivers delete "$CREATED_DRIVER_ID"
assert_failure


test_name "Delete raw transport driver with --confirm"
xbe_run do raw-transport-drivers delete "$CREATED_DRIVER_ID" --confirm
assert_success

# ==========================================================================
# Summary
# ==========================================================================

run_tests
