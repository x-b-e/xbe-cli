#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Tractors
#
# Tests list, show, create, update, and delete operations for raw-transport-tractors.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_RAW_TRANSPORT_TRACTOR_ID=""

describe "Resource: raw-transport-tractors"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create broker for raw transport tractor tests"
BROKER_NAME=$(unique_name "RawTransportBroker")

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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create raw transport tractor"
RAW_EXTERNAL_TRACTOR_ID=$(unique_name "RawTransportTractor")
TABLES_JSON='[{"table_name":"tractorprofile","query":"SELECT * FROM tractorprofile","primary_key_column":"trc_number","rows":[{"columns":[{"key":"trc_number","value":"TRC-123"},{"key":"trc_make","value":"KENWORTH"},{"key":"trc_model","value":"T680"}]}]}]'

if [[ -n "$CREATED_BROKER_ID" && "$CREATED_BROKER_ID" != "null" ]]; then
    xbe_json do raw-transport-tractors create \
        --external-tractor-id "$RAW_EXTERNAL_TRACTOR_ID" \
        --broker "$CREATED_BROKER_ID" \
        --importer "quantix_tmw" \
        --tables "$TABLES_JSON"
    if [[ $status -eq 0 ]]; then
        CREATED_RAW_TRANSPORT_TRACTOR_ID=$(json_get ".id")
        if [[ -n "$CREATED_RAW_TRANSPORT_TRACTOR_ID" && "$CREATED_RAW_TRANSPORT_TRACTOR_ID" != "null" ]]; then
            register_cleanup "raw-transport-tractors" "$CREATED_RAW_TRANSPORT_TRACTOR_ID"
            pass
        else
            fail "Created raw transport tractor but no ID returned"
        fi
    else
        fail "Failed to create raw transport tractor"
    fi
else
    skip "Missing broker ID for creation"
fi

if [[ -z "$CREATED_RAW_TRANSPORT_TRACTOR_ID" || "$CREATED_RAW_TRANSPORT_TRACTOR_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport tractor ID"
    run_tests
fi

# ============================================================================
# LIST Tests
# ============================================================================

test_name "List raw transport tractors"
xbe_json view raw-transport-tractors list --limit 50
assert_json_is_array

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show raw transport tractor"
xbe_json view raw-transport-tractors show "$CREATED_RAW_TRANSPORT_TRACTOR_ID"
assert_success

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "Filter by broker"
xbe_json view raw-transport-tractors list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "Filter by importer"
xbe_json view raw-transport-tractors list --importer "quantix_tmw" --limit 5
assert_success

test_name "Filter by import status"
xbe_json view raw-transport-tractors list --import-status "pending" --limit 5
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update raw transport tractor (not supported)"
xbe_run do raw-transport-tractors update "$CREATED_RAW_TRANSPORT_TRACTOR_ID"
assert_failure

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete raw transport tractor"
xbe_run do raw-transport-tractors delete "$CREATED_RAW_TRANSPORT_TRACTOR_ID" --confirm
assert_success

run_tests
