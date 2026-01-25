#!/bin/bash
#
# XBE CLI Integration Tests: Raw Transport Orders
#
# Tests CRUD operations for the raw-transport-orders resource.
# Raw transport orders store imported TMW payloads before normalization.
#
# COVERAGE: Create (all writable attrs), list filters, show, delete
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

CREATED_BROKER_ID=""
CREATED_ORDER_ID=""
CREATED_ORDER_DELETE_ID=""
ROWVERSION_MIN=""
ROWVERSION_MAX=""


describe "Resource: raw-transport-orders"

# ============================================================================
# Prerequisites - Create broker
# ============================================================================

test_name "Create prerequisite broker for raw transport orders"
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

test_name "Create raw transport order with required fields"
ORDER_NUMBER="RAW-ORD-$(date +%s)-$RANDOM"
TABLES_JSON=$(cat <<EOT
[
  {
    "table_name": "orderheader",
    "query": "select * from orderheader",
    "primary_key_column": "ord_hdrnumber",
    "rows": [
      {
        "columns": [
          {"key": "ord_hdrnumber", "value": "$ORDER_NUMBER"}
        ]
      }
    ]
  }
]
EOT
)

xbe_json do raw-transport-orders create \
    --external-order-number "$ORDER_NUMBER" \
    --broker "$CREATED_BROKER_ID" \
    --tables "$TABLES_JSON"

if [[ $status -eq 0 ]]; then
    CREATED_ORDER_ID=$(json_get ".id")
    if [[ -n "$CREATED_ORDER_ID" && "$CREATED_ORDER_ID" != "null" ]]; then
        register_cleanup "raw-transport-orders" "$CREATED_ORDER_ID"
        pass
    else
        fail "Created raw transport order but no ID returned"
    fi
else
    fail "Failed to create raw transport order"
fi

# Only continue if we successfully created a raw transport order
if [[ -z "$CREATED_ORDER_ID" || "$CREATED_ORDER_ID" == "null" ]]; then
    echo "Cannot continue without a valid raw transport order ID"
    run_tests
fi

test_name "Create raw transport order with importer, managed flag, and rowversions"
ORDER_NUMBER2="RAW-ORD-$(date +%s)-$RANDOM"
ROWVERSION_MIN="100"
ROWVERSION_MAX="200"
TABLES_JSON2=$(cat <<EOT
[
  {
    "table_name": "orderheader",
    "query": "select * from orderheader",
    "primary_key_column": "ord_hdrnumber",
    "rows": [
      {
        "columns": [
          {"key": "ord_hdrnumber", "value": "$ORDER_NUMBER2"}
        ]
      }
    ]
  }
]
EOT
)

xbe_json do raw-transport-orders create \
    --external-order-number "$ORDER_NUMBER2" \
    --broker "$CREATED_BROKER_ID" \
    --importer quantix_tmw \
    --is-managed \
    --tables-rowversion-min "$ROWVERSION_MIN" \
    --tables-rowversion-max "$ROWVERSION_MAX" \
    --tables "$TABLES_JSON2"

if [[ $status -eq 0 ]]; then
    CREATED_ORDER_DELETE_ID=$(json_get ".id")
    if [[ -n "$CREATED_ORDER_DELETE_ID" && "$CREATED_ORDER_DELETE_ID" != "null" ]]; then
        register_cleanup "raw-transport-orders" "$CREATED_ORDER_DELETE_ID"
        pass
    else
        fail "Created raw transport order but no ID returned"
    fi
else
    fail "Failed to create raw transport order with optional fields"
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List raw transport orders"
xbe_json view raw-transport-orders list --limit 5
assert_success

test_name "List raw transport orders returns array"
xbe_json view raw-transport-orders list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list raw transport orders"
fi

# ============================================================================
# SHOW Tests
# ============================================================================

test_name "Show raw transport order"
xbe_json view raw-transport-orders show "$CREATED_ORDER_ID"
assert_success

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update raw transport order is not supported"
xbe_run do raw-transport-orders update "$CREATED_ORDER_ID"
assert_failure

# ============================================================================
# LIST Tests - Filters
# ============================================================================

test_name "List raw transport orders with --broker"
xbe_json view raw-transport-orders list --broker "$CREATED_BROKER_ID" --limit 5
assert_success

test_name "List raw transport orders with --tables-rowversion-min"
xbe_json view raw-transport-orders list --tables-rowversion-min "$ROWVERSION_MIN" --limit 5
assert_success

test_name "List raw transport orders with --tables-rowversion-max"
xbe_json view raw-transport-orders list --tables-rowversion-max "$ROWVERSION_MAX" --limit 5
assert_success

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete raw transport order requires --confirm"
if [[ -n "$CREATED_ORDER_DELETE_ID" && "$CREATED_ORDER_DELETE_ID" != "null" ]]; then
    xbe_run do raw-transport-orders delete "$CREATED_ORDER_DELETE_ID"
    assert_failure
else
    skip "No raw transport order available for delete"
fi

test_name "Delete raw transport order"
if [[ -n "$CREATED_ORDER_DELETE_ID" && "$CREATED_ORDER_DELETE_ID" != "null" ]]; then
    xbe_run do raw-transport-orders delete "$CREATED_ORDER_DELETE_ID" --confirm
    assert_success
else
    skip "No raw transport order available for delete"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
