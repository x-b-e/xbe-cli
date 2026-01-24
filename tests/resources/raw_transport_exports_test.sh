#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Exports
#
# Tests list, show, create, update, and delete operations for raw-transport-exports.
#
# COVERAGE: All list filters + writable attributes
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_ID=""
CUSTOMER_ID=""
TRANSPORT_ORDER_ID=""
RAW_TRANSPORT_EXPORT_ID=""

EXTERNAL_ORDER_NUMBER=""
TARGET_DATABASE="tmw"
TARGET_TABLE="stops"
EXPORT_TYPE="quantix_tmw"
CHECKSUM=""
SEQUENCE="1"

FIRST_SEEN_AT="2025-01-01T00:00:00Z"
THROTTLED_UNTIL="2025-01-01T12:00:00Z"
EXPORTED_AT="2025-01-02T00:00:00Z"

HEADERS_JSON='["ord_hdrnumber","stp_number"]'
ROWS_JSON='[["4001","1"],["4001","2"]]'
EXPORT_RESULTS_JSON='{"status":"queued","message":"test"}'
NOT_EXPORTABLE_REASONS_JSON='["missing_data"]'
STP_NUMBERS_JSON='["100","101"]'
FORMATTED_EXPORT="ord_hdrnumber,stp_number;4001,1"


describe "Resource: raw-transport-exports"

# ==========================================================================
# Prerequisites - Create broker, customer, transport order
# ==========================================================================

test_name "Create broker for raw transport export tests"
BROKER_NAME=$(unique_name "RawTransportExportBroker")

xbe_json do brokers create --name "$BROKER_NAME"

if [[ $status -eq 0 ]]; then
    BROKER_ID=$(json_get ".id")
    if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
        register_cleanup "brokers" "$BROKER_ID"
        pass
    else
        fail "Created broker but no ID returned"
        echo "Cannot continue without a broker"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_BROKER_ID" ]]; then
        BROKER_ID="$XBE_TEST_BROKER_ID"
        echo "    Using XBE_TEST_BROKER_ID: $BROKER_ID"
        pass
    else
        fail "Failed to create broker and XBE_TEST_BROKER_ID not set"
        echo "Cannot continue without a broker"
        run_tests
    fi
fi

test_name "Create customer for raw transport export tests"
CUSTOMER_NAME=$(unique_name "RawTransportExportCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$BROKER_ID"

if [[ $status -eq 0 ]]; then
    CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CUSTOMER_ID" && "$CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Create transport order for raw transport export tests"

xbe_json do transport-orders create --customer "$CUSTOMER_ID"

if [[ $status -eq 0 ]]; then
    TRANSPORT_ORDER_ID=$(json_get ".id")
    if [[ -n "$TRANSPORT_ORDER_ID" && "$TRANSPORT_ORDER_ID" != "null" ]]; then
        register_cleanup "transport-orders" "$TRANSPORT_ORDER_ID"
        pass
    else
        fail "Created transport order but no ID returned"
        echo "Cannot continue without a transport order"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_TRANSPORT_ORDER_ID" ]]; then
        TRANSPORT_ORDER_ID="$XBE_TEST_TRANSPORT_ORDER_ID"
        echo "    Using XBE_TEST_TRANSPORT_ORDER_ID: $TRANSPORT_ORDER_ID"
        pass
    else
        fail "Failed to create transport order and XBE_TEST_TRANSPORT_ORDER_ID not set"
        echo "Cannot continue without a transport order"
        run_tests
    fi
fi

# ==========================================================================
# CREATE Tests
# ==========================================================================

test_name "Create raw transport export"
EXTERNAL_ORDER_NUMBER=$(unique_name "RawTransportExport")
CHECKSUM="chk-$(unique_suffix)"

if [[ -n "$BROKER_ID" && "$BROKER_ID" != "null" ]]; then
    xbe_json do raw-transport-exports create \
        --external-order-number "$EXTERNAL_ORDER_NUMBER" \
        --target-database "$TARGET_DATABASE" \
        --target-table "$TARGET_TABLE" \
        --export-type "$EXPORT_TYPE" \
        --headers "$HEADERS_JSON" \
        --rows "$ROWS_JSON" \
        --formatted-export "$FORMATTED_EXPORT" \
        --checksum "$CHECKSUM" \
        --sequence "$SEQUENCE" \
        --stp-numbers "$STP_NUMBERS_JSON" \
        --is-exportable=false \
        --is-exported=false \
        --not-exportable-reasons "$NOT_EXPORTABLE_REASONS_JSON" \
        --issue-type "missing_data" \
        --first-seen-at "$FIRST_SEEN_AT" \
        --throttled-until "$THROTTLED_UNTIL" \
        --export-results "$EXPORT_RESULTS_JSON" \
        --exported-at "$EXPORTED_AT" \
        --broker "$BROKER_ID" \
        --transport-order "$TRANSPORT_ORDER_ID"

    if [[ $status -eq 0 ]]; then
        RAW_TRANSPORT_EXPORT_ID=$(json_get ".id")
        if [[ -n "$RAW_TRANSPORT_EXPORT_ID" && "$RAW_TRANSPORT_EXPORT_ID" != "null" ]]; then
            pass
        else
            fail "Created raw transport export but no ID returned"
        fi
    else
        fail "Failed to create raw transport export"
    fi
else
    skip "Missing broker ID for creation"
fi

if [[ -z "$RAW_TRANSPORT_EXPORT_ID" || "$RAW_TRANSPORT_EXPORT_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport export ID"
    run_tests
fi

# ==========================================================================
# LIST Tests
# ==========================================================================

test_name "List raw transport exports"
xbe_json view raw-transport-exports list --limit 50
assert_json_is_array

# ==========================================================================
# SHOW Tests
# ==========================================================================

test_name "Show raw transport export"
xbe_json view raw-transport-exports show "$RAW_TRANSPORT_EXPORT_ID"
assert_success

# ==========================================================================
# LIST Tests - Filters
# ==========================================================================

test_name "Filter by broker"
xbe_json view raw-transport-exports list --broker "$BROKER_ID" --limit 5
assert_success

test_name "Filter by transport order"
xbe_json view raw-transport-exports list --transport-order "$TRANSPORT_ORDER_ID" --limit 5
assert_success

test_name "Filter by external order number"
xbe_json view raw-transport-exports list --external-order-number "$EXTERNAL_ORDER_NUMBER" --limit 5
assert_success

test_name "Filter by export type"
xbe_json view raw-transport-exports list --export-type "$EXPORT_TYPE" --limit 5
assert_success

test_name "Filter by target table"
xbe_json view raw-transport-exports list --target-table "$TARGET_TABLE" --limit 5
assert_success

test_name "Filter by issue type"
xbe_json view raw-transport-exports list --issue-type "missing_data" --limit 5
assert_success

test_name "Filter by exportable flag"
xbe_json view raw-transport-exports list --is-exportable false --limit 5
assert_success

test_name "Filter by exported flag"
xbe_json view raw-transport-exports list --is-exported false --limit 5
assert_success

test_name "Filter by checksum"
xbe_json view raw-transport-exports list --checksum "$CHECKSUM" --limit 5
assert_success

test_name "Filter by sequence"
xbe_json view raw-transport-exports list --sequence "$SEQUENCE" --limit 5
assert_success

test_name "Filter by created at min"
xbe_json view raw-transport-exports list --created-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by created at max"
xbe_json view raw-transport-exports list --created-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by created at presence"
xbe_json view raw-transport-exports list --is-created-at true --limit 5
assert_success

test_name "Filter by exported at min"
xbe_json view raw-transport-exports list --exported-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by exported at max"
xbe_json view raw-transport-exports list --exported-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by exported at presence"
xbe_json view raw-transport-exports list --is-exported-at true --limit 5
assert_success

test_name "Filter by first seen at min"
xbe_json view raw-transport-exports list --first-seen-at-min "2020-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by first seen at max"
xbe_json view raw-transport-exports list --first-seen-at-max "2030-01-01T00:00:00Z" --limit 5
assert_success

test_name "Filter by first seen at presence"
xbe_json view raw-transport-exports list --is-first-seen-at true --limit 5
assert_success

test_name "Filter by recent"
xbe_json view raw-transport-exports list --recent 24 --limit 5
assert_success

# ==========================================================================
# UPDATE Tests (not supported)
# ==========================================================================

test_name "Update raw transport export (not supported)"
xbe_run do raw-transport-exports update "$RAW_TRANSPORT_EXPORT_ID"
assert_failure

# ==========================================================================
# DELETE Tests (not supported)
# ==========================================================================

test_name "Delete raw transport export (not supported)"
xbe_run do raw-transport-exports delete "$RAW_TRANSPORT_EXPORT_ID"
assert_failure

run_tests
