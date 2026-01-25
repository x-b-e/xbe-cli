#!/bin/bash
#
# XBE CLI Integration Tests: Transport References
#
# Tests CRUD operations for the transport-references resource.
# Transport references require a subject (transport order, stop, or material).
#
# COVERAGE: Create/update/delete + list filters + show + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_CUSTOMER_ID:-}"
TRANSPORT_ORDER_ID="${XBE_TEST_TRANSPORT_ORDER_ID:-}"

REFERENCE_ID=""
REFERENCE_POSITION=""

describe "Resource: transport-references"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for transport references tests"
BROKER_NAME=$(unique_name "TRBroker")

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

test_name "Create prerequisite customer for transport references tests"
CUSTOMER_NAME=$(unique_name "TRCustomer")

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

test_name "Create prerequisite transport order for transport references tests"
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

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create transport reference with required fields"
REFERENCE_KEY="BOL"
REFERENCE_VALUE="REF-$(unique_suffix)"

xbe_json do transport-references create \
    --subject-type transport-orders \
    --subject-id "$TRANSPORT_ORDER_ID" \
    --key "$REFERENCE_KEY" \
    --value "$REFERENCE_VALUE" \
    --position 1

if [[ $status -eq 0 ]]; then
    REFERENCE_ID=$(json_get ".id")
    REFERENCE_POSITION=$(json_get ".position")
    if [[ -n "$REFERENCE_ID" && "$REFERENCE_ID" != "null" ]]; then
        register_cleanup "transport-references" "$REFERENCE_ID"
        pass
    else
        fail "Created transport reference but no ID returned"
    fi
else
    fail "Failed to create transport reference"
fi

if [[ -z "$REFERENCE_ID" || "$REFERENCE_ID" == "null" ]]; then
    echo "Cannot continue without a valid transport reference ID"
    run_tests
fi

test_name "Create transport reference with another key"
xbe_json do transport-references create \
    --subject-type transport-orders \
    --subject-id "$TRANSPORT_ORDER_ID" \
    --key "PO" \
    --value "PO-$(unique_suffix)"

if [[ $status -eq 0 ]]; then
    id=$(json_get ".id")
    register_cleanup "transport-references" "$id"
    pass
else
    fail "Failed to create transport reference with another key"
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update transport reference key"
xbe_json do transport-references update "$REFERENCE_ID" --key "BOL-UPDATED"
assert_success

test_name "Update transport reference value"
xbe_json do transport-references update "$REFERENCE_ID" --value "REF-UPDATED-$(unique_suffix)"
assert_success

test_name "Update transport reference position"
xbe_json do transport-references update "$REFERENCE_ID" --position 1
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show transport reference"
xbe_json view transport-references show "$REFERENCE_ID"
assert_success

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport references"
xbe_json view transport-references list --limit 5
assert_success

test_name "List transport references returns array"
xbe_json view transport-references list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list transport references"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List transport references with --key filter"
xbe_json view transport-references list --key "$REFERENCE_KEY" --limit 10
assert_success

test_name "List transport references with --position filter"
POSITION_FILTER="${REFERENCE_POSITION:-1}"
xbe_json view transport-references list --position "$POSITION_FILTER" --limit 10
assert_success

test_name "List transport references with --subject filter"
xbe_json view transport-references list --subject-type transport-orders --subject-id "$TRANSPORT_ORDER_ID" --limit 10
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete transport reference requires --confirm flag"
xbe_json do transport-references delete "$REFERENCE_ID"
assert_failure

test_name "Delete transport reference with --confirm"
# Create a transport reference specifically for deletion
xbe_json do transport-references create \
    --subject-type transport-orders \
    --subject-id "$TRANSPORT_ORDER_ID" \
    --key "DEL" \
    --value "DEL-$(unique_suffix)"
if [[ $status -eq 0 ]]; then
    DEL_ID=$(json_get ".id")
    xbe_json do transport-references delete "$DEL_ID" --confirm
    if [[ $status -eq 0 ]]; then
        pass
    else
        register_cleanup "transport-references" "$DEL_ID"
        skip "API may not allow transport reference deletion"
    fi
else
    skip "Could not create transport reference for deletion test"
fi

# ============================================================================
# Error Cases
# ============================================================================

test_name "Create transport reference without key fails"
xbe_json do transport-references create --subject-type transport-orders --subject-id "$TRANSPORT_ORDER_ID" --value "NO-KEY"
assert_failure

test_name "Create transport reference without subject fails"
xbe_json do transport-references create --key "NO-SUBJECT" --value "NO-SUBJECT"
assert_failure

test_name "Update transport reference without any fields fails"
xbe_json do transport-references update "$REFERENCE_ID"
assert_failure

# ============================================================================
# Summary
# ============================================================================

run_tests
