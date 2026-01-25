#!/bin/bash
#
# XBE CLI Integration Tests: Transport Order Stops
#
# Tests CRUD operations for the transport_order_stops resource.
#
# COVERAGE: All filters + all create/update attributes and relationships
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_CUSTOMER_ID=""
CREATED_TRANSPORT_ORDER_ID=""
CREATED_LOCATION_ID=""
CREATED_LOCATION_ID_2=""
CREATED_STOP_ID=""

AT_MIN="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
AT_MAX="$(date -u -v+1H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")"
UPDATED_AT_MIN="$(date -u -v+2H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+2 hours" +"%Y-%m-%dT%H:%M:%SZ")"
UPDATED_AT_MAX="$(date -u -v+3H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+3 hours" +"%Y-%m-%dT%H:%M:%SZ")"

EXTERNAL_TMS_STOP_NUMBER="TMS-STOP-$(date +%s)"
UPDATED_EXTERNAL_TMS_STOP_NUMBER="TMS-STOP-UPDATED-$(date +%s)"

DIRECT_API_AVAILABLE=0
if [[ -n "$XBE_TOKEN" ]]; then
    DIRECT_API_AVAILABLE=1
fi

cleanup_api_resources() {
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        return
    fi

    if [[ -n "$CREATED_LOCATION_ID_2" && "$CREATED_LOCATION_ID_2" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/project-transport-locations/$CREATED_LOCATION_ID_2" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi

    if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
        curl -sS -X DELETE "$XBE_BASE_URL/v1/project-transport-locations/$CREATED_LOCATION_ID" \
            -H "Authorization: Bearer $XBE_TOKEN" >/dev/null 2>&1 || true
    fi
}

trap 'cleanup_api_resources; run_cleanup' EXIT

api_post() {
    local path="$1"
    local body="$2"
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        output="Missing XBE_TOKEN for direct API calls"
        status=1
        return
    fi
    run curl -sS -X POST "$XBE_BASE_URL$path" \
        -H "Authorization: Bearer $XBE_TOKEN" \
        -H "Content-Type: application/vnd.api+json" \
        -d "$body"
}

describe "Resource: transport-order-stops"

# ============================================================================
# Prerequisites - Create broker, customer, transport order, and location
# ============================================================================

test_name "Create prerequisite broker for transport order stops tests"
BROKER_NAME=$(unique_name "TOStopBroker")

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

test_name "Create prerequisite customer for transport order stops tests"
CUSTOMER_NAME=$(unique_name "TOStopCustomer")

xbe_json do customers create --name "$CUSTOMER_NAME" --broker "$CREATED_BROKER_ID"

if [[ $status -eq 0 ]]; then
    CREATED_CUSTOMER_ID=$(json_get ".id")
    if [[ -n "$CREATED_CUSTOMER_ID" && "$CREATED_CUSTOMER_ID" != "null" ]]; then
        register_cleanup "customers" "$CREATED_CUSTOMER_ID"
        pass
    else
        fail "Created customer but no ID returned"
        echo "Cannot continue without a customer"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_CUSTOMER_ID" ]]; then
        CREATED_CUSTOMER_ID="$XBE_TEST_CUSTOMER_ID"
        echo "    Using XBE_TEST_CUSTOMER_ID: $CREATED_CUSTOMER_ID"
        pass
    else
        fail "Failed to create customer and XBE_TEST_CUSTOMER_ID not set"
        echo "Cannot continue without a customer"
        run_tests
    fi
fi

test_name "Select or create transport order for stop tests"
if [[ -n "$XBE_TEST_TRANSPORT_ORDER_ID" ]]; then
    CREATED_TRANSPORT_ORDER_ID="$XBE_TEST_TRANSPORT_ORDER_ID"
    echo "    Using XBE_TEST_TRANSPORT_ORDER_ID: $CREATED_TRANSPORT_ORDER_ID"
    pass
else
    xbe_json do transport-orders create --customer "$CREATED_CUSTOMER_ID"

    if [[ $status -eq 0 ]]; then
        CREATED_TRANSPORT_ORDER_ID=$(json_get ".id")
        if [[ -n "$CREATED_TRANSPORT_ORDER_ID" && "$CREATED_TRANSPORT_ORDER_ID" != "null" ]]; then
            pass
        else
            fail "Created transport order but no ID returned"
            echo "Cannot continue without a transport order"
            run_tests
        fi
    else
        fail "Failed to create transport order"
        echo "Cannot continue without a transport order"
        run_tests
    fi
fi

if [[ -n "$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID" ]]; then
    CREATED_LOCATION_ID="$XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID"
    echo "    Using XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID: $CREATED_LOCATION_ID"
else
    if [[ $DIRECT_API_AVAILABLE -eq 0 ]]; then
        skip "Set XBE_TOKEN or XBE_TEST_PROJECT_TRANSPORT_LOCATION_ID to run transport order stop tests"
        run_tests
    fi

    test_name "Create transport location for stop tests"
    LOCATION_NAME=$(unique_name "TOStopLocation")
    api_post "/v1/project-transport-locations" "{\"data\":{\"type\":\"project-transport-locations\",\"attributes\":{\"name\":\"$LOCATION_NAME\",\"geocoding-method\":\"explicit\",\"address-latitude\":34.05,\"address-longitude\":-118.25,\"is-valid-for-stop\":true},\"relationships\":{\"broker\":{\"data\":{\"type\":\"brokers\",\"id\":\"$CREATED_BROKER_ID\"}}}}}"

    if [[ $status -eq 0 ]]; then
        CREATED_LOCATION_ID=$(json_get ".data.id")
        if [[ -n "$CREATED_LOCATION_ID" && "$CREATED_LOCATION_ID" != "null" ]]; then
            pass
        else
            fail "Created location but no ID returned"
            echo "Cannot continue without a location"
            run_tests
        fi
    else
        fail "Failed to create transport location"
        run_tests
    fi
fi

# Optional second location for update tests
if [[ $DIRECT_API_AVAILABLE -eq 1 ]]; then
    test_name "Create second transport location for update tests"
    LOCATION_NAME_2=$(unique_name "TOStopLocationTwo")
    api_post "/v1/project-transport-locations" "{\"data\":{\"type\":\"project-transport-locations\",\"attributes\":{\"name\":\"$LOCATION_NAME_2\",\"geocoding-method\":\"explicit\",\"address-latitude\":34.06,\"address-longitude\":-118.26,\"is-valid-for-stop\":true},\"relationships\":{\"broker\":{\"data\":{\"type\":\"brokers\",\"id\":\"$CREATED_BROKER_ID\"}}}}}"

    if [[ $status -eq 0 ]]; then
        CREATED_LOCATION_ID_2=$(json_get ".data.id")
        if [[ -n "$CREATED_LOCATION_ID_2" && "$CREATED_LOCATION_ID_2" != "null" ]]; then
            pass
        else
            fail "Created second location but no ID returned"
        fi
    else
        fail "Failed to create second transport location"
    fi
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create transport order stop with required fields"
xbe_json do transport-order-stops create \
    --transport-order "$CREATED_TRANSPORT_ORDER_ID" \
    --location "$CREATED_LOCATION_ID" \
    --role pickup \
    --position 1 \
    --status planned \
    --at-min "$AT_MIN" \
    --at-max "$AT_MAX" \
    --external-tms-stop-number "$EXTERNAL_TMS_STOP_NUMBER"

if [[ $status -eq 0 ]]; then
    CREATED_STOP_ID=$(json_get ".id")
    if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then
        register_cleanup "transport-order-stops" "$CREATED_STOP_ID"
        pass
    else
        fail "Created stop but no ID returned"
    fi
else
    fail "Failed to create transport order stop: $output"
fi

test_name "Create transport order stop without required fields fails"
xbe_json do transport-order-stops create
assert_failure

# ============================================================================
# UPDATE Tests
# ============================================================================

if [[ -n "$CREATED_STOP_ID" && "$CREATED_STOP_ID" != "null" ]]; then

    test_name "Update transport order stop status"
    xbe_json do transport-order-stops update "$CREATED_STOP_ID" --status started
    assert_success

    test_name "Update transport order stop role"
    xbe_json do transport-order-stops update "$CREATED_STOP_ID" --role delivery
    assert_success

    test_name "Update transport order stop position"
    xbe_json do transport-order-stops update "$CREATED_STOP_ID" --position 2
    assert_success

    test_name "Update transport order stop timing"
    xbe_json do transport-order-stops update "$CREATED_STOP_ID" \
        --at-min "$UPDATED_AT_MIN" \
        --at-max "$UPDATED_AT_MAX"
    assert_success

    test_name "Update transport order stop external TMS stop number"
    xbe_json do transport-order-stops update "$CREATED_STOP_ID" --external-tms-stop-number "$UPDATED_EXTERNAL_TMS_STOP_NUMBER"
    assert_success

    if [[ -n "$CREATED_LOCATION_ID_2" && "$CREATED_LOCATION_ID_2" != "null" ]]; then
        test_name "Update transport order stop location"
        xbe_json do transport-order-stops update "$CREATED_STOP_ID" --location "$CREATED_LOCATION_ID_2"
        assert_success
    else
        skip "Skipping location update test (second location not available)"
    fi
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show transport order stop"
xbe_json view transport-order-stops show "$CREATED_STOP_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport order stops"
xbe_json view transport-order-stops list --limit 5
assert_success

test_name "List transport order stops returns array"
xbe_json view transport-order-stops list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list transport order stops"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

TEST_TIMESTAMP="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

test_name "List transport order stops with --transport-order filter"
xbe_json view transport-order-stops list --transport-order "$CREATED_TRANSPORT_ORDER_ID"
assert_success

test_name "List transport order stops with --location filter"
xbe_json view transport-order-stops list --location "$CREATED_LOCATION_ID"
assert_success

test_name "List transport order stops with --role filter"
xbe_json view transport-order-stops list --role pickup
assert_success

test_name "List transport order stops with --at-min-min filter"
xbe_json view transport-order-stops list --at-min-min "$AT_MIN"
assert_success

test_name "List transport order stops with --at-min-max filter"
xbe_json view transport-order-stops list --at-min-max "$AT_MAX"
assert_success

test_name "List transport order stops with --is-at-min filter"
xbe_json view transport-order-stops list --is-at-min true
assert_success

test_name "List transport order stops with --at-max-min filter"
xbe_json view transport-order-stops list --at-max-min "$AT_MIN"
assert_success

test_name "List transport order stops with --at-max-max filter"
xbe_json view transport-order-stops list --at-max-max "$AT_MAX"
assert_success

test_name "List transport order stops with --is-at-max filter"
xbe_json view transport-order-stops list --is-at-max true
assert_success

test_name "List transport order stops with --external-tms-stop-number filter"
xbe_json view transport-order-stops list --external-tms-stop-number "$UPDATED_EXTERNAL_TMS_STOP_NUMBER"
assert_success

test_name "List transport order stops with --external-identification-value filter"
xbe_json view transport-order-stops list --external-identification-value "$UPDATED_EXTERNAL_TMS_STOP_NUMBER"
assert_success

test_name "List transport order stops with --created-at-min filter"
xbe_json view transport-order-stops list --created-at-min "$TEST_TIMESTAMP"
assert_success

test_name "List transport order stops with --created-at-max filter"
xbe_json view transport-order-stops list --created-at-max "$TEST_TIMESTAMP"
assert_success

test_name "List transport order stops with --is-created-at filter"
xbe_json view transport-order-stops list --is-created-at true
assert_success

test_name "List transport order stops with --updated-at-min filter"
xbe_json view transport-order-stops list --updated-at-min "$TEST_TIMESTAMP"
assert_success

test_name "List transport order stops with --updated-at-max filter"
xbe_json view transport-order-stops list --updated-at-max "$TEST_TIMESTAMP"
assert_success

test_name "List transport order stops with --is-updated-at filter"
xbe_json view transport-order-stops list --is-updated-at true
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete transport order stop requires --confirm"
xbe_json do transport-order-stops delete "$CREATED_STOP_ID"
assert_failure

run_tests
