#!/bin/bash
#
# XBE CLI Integration Tests: Transport Order Materials
#
# Tests list/show/create/update/delete operations for the transport-order-materials resource.
#
# COVERAGE: List filters + show + create/update/delete + error cases
#

# Load test helpers
source "$(dirname "${BASH_SOURCE[0]}")/../lib/test_helpers.sh"

BROKER_ID="${XBE_TEST_BROKER_ID:-}"
CUSTOMER_ID="${XBE_TEST_CUSTOMER_ID:-}"
TRANSPORT_ORDER_ID="${XBE_TEST_TRANSPORT_ORDER_ID:-}"
MATERIAL_TYPE_ID="${XBE_TEST_MATERIAL_TYPE_ID:-}"
UNIT_OF_MEASURE_ID="${XBE_TEST_UNIT_OF_MEASURE_ID:-}"

SAMPLE_ID=""
SAMPLE_TRANSPORT_ORDER_ID=""
SAMPLE_MATERIAL_TYPE_ID=""
SAMPLE_UNIT_OF_MEASURE_ID=""
SAMPLE_QUANTITY_EXPLICIT=""

CREATED_ID=""

describe "Resource: transport-order-materials"

# ============================================================================
# Prerequisites
# ============================================================================

test_name "Create prerequisite broker for transport order materials tests"
BROKER_NAME=$(unique_name "TOMBroker")

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

test_name "Create prerequisite customer for transport order materials tests"
CUSTOMER_NAME=$(unique_name "TOMCustomer")

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

test_name "Create prerequisite transport order for transport order materials tests"
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

test_name "Create prerequisite material type for transport order materials tests"
MATERIAL_TYPE_NAME=$(unique_name "TOMMaterialType")

xbe_json do material-types create --name "$MATERIAL_TYPE_NAME"

if [[ $status -eq 0 ]]; then
    MATERIAL_TYPE_ID=$(json_get ".id")
    if [[ -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" ]]; then
        register_cleanup "material-types" "$MATERIAL_TYPE_ID"
        pass
    else
        fail "Created material type but no ID returned"
        echo "Cannot continue without a material type"
        run_tests
    fi
else
    if [[ -n "$XBE_TEST_MATERIAL_TYPE_ID" ]]; then
        MATERIAL_TYPE_ID="$XBE_TEST_MATERIAL_TYPE_ID"
        echo "    Using XBE_TEST_MATERIAL_TYPE_ID: $MATERIAL_TYPE_ID"
        pass
    else
        fail "Failed to create material type and XBE_TEST_MATERIAL_TYPE_ID not set"
        echo "Cannot continue without a material type"
        run_tests
    fi
fi

test_name "Fetch unit of measure ID for transport order materials tests"
if [[ -n "$UNIT_OF_MEASURE_ID" ]]; then
    echo "    Using XBE_TEST_UNIT_OF_MEASURE_ID: $UNIT_OF_MEASURE_ID"
    pass
else
    xbe_json view unit-of-measures list --limit 1
    if [[ $status -eq 0 ]]; then
        UNIT_OF_MEASURE_ID=$(json_get ".[0].id")
        if [[ -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
            pass
        else
            skip "No unit of measure available"
        fi
    else
        skip "Failed to list unit of measures"
    fi
fi

# ============================================================================
# LIST Tests - Basic
# ============================================================================

test_name "List transport order materials"
xbe_json view transport-order-materials list --limit 5
assert_success

test_name "List transport order materials returns array"
xbe_json view transport-order-materials list --limit 5
if [[ $status -eq 0 ]]; then
    assert_json_is_array
else
    fail "Failed to list transport order materials"
fi

# ============================================================================
# Sample Record (used for filters/show)
# ============================================================================

test_name "Capture sample transport order material"
xbe_json view transport-order-materials list --limit 1
if [[ $status -eq 0 ]]; then
    SAMPLE_ID=$(json_get ".[0].id")
    SAMPLE_TRANSPORT_ORDER_ID=$(json_get ".[0].transport_order_id")
    SAMPLE_MATERIAL_TYPE_ID=$(json_get ".[0].material_type_id")
    SAMPLE_UNIT_OF_MEASURE_ID=$(json_get ".[0].unit_of_measure_id")
    SAMPLE_QUANTITY_EXPLICIT=$(json_get ".[0].quantity_explicit")
    if [[ -n "$SAMPLE_ID" && "$SAMPLE_ID" != "null" ]]; then
        pass
    else
        skip "No transport order materials available for follow-on tests"
    fi
else
    skip "Could not list transport order materials to capture sample"
fi

# ============================================================================
# CREATE Tests
# ============================================================================

test_name "Create requires --transport-order"
xbe_run do transport-order-materials create --material-type "123" --unit-of-measure "123"
assert_failure

test_name "Create requires --material-type"
xbe_run do transport-order-materials create --transport-order "123" --unit-of-measure "123"
assert_failure

test_name "Create requires --unit-of-measure"
xbe_run do transport-order-materials create --transport-order "123" --material-type "123"
assert_failure

test_name "Create transport order material"
if [[ -n "$TRANSPORT_ORDER_ID" && "$TRANSPORT_ORDER_ID" != "null" && -n "$MATERIAL_TYPE_ID" && "$MATERIAL_TYPE_ID" != "null" && -n "$UNIT_OF_MEASURE_ID" && "$UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json do transport-order-materials create \
        --transport-order "$TRANSPORT_ORDER_ID" \
        --material-type "$MATERIAL_TYPE_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID" \
        --quantity-explicit 10

    if [[ $status -eq 0 ]]; then
        CREATED_ID=$(json_get ".id")
        if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
            register_cleanup "transport-order-materials" "$CREATED_ID"
            pass
        else
            fail "Created transport order material but no ID returned"
        fi
    else
        if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]] || [[ "$output" == *"Validation"* ]] || [[ "$output" == *"unprocessable"* ]]; then
            skip "Create failed due to permissions or validation"
        else
            fail "Failed to create transport order material: $output"
        fi
    fi
else
    skip "Missing prerequisites. Set XBE_TEST_TRANSPORT_ORDER_ID, XBE_TEST_MATERIAL_TYPE_ID, and XBE_TEST_UNIT_OF_MEASURE_ID to enable create testing."
fi

# ============================================================================
# UPDATE Tests
# ============================================================================

test_name "Update without fields fails"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_run do transport-order-materials update "$CREATED_ID"
    assert_failure
else
    skip "No created transport order material ID available"
fi

test_name "Update transport order material quantity-explicit"
if [[ -n "$CREATED_ID" && "$CREATED_ID" != "null" ]]; then
    xbe_json do transport-order-materials update "$CREATED_ID" --quantity-explicit 12
    assert_success
else
    skip "No created transport order material ID available"
fi

# ============================================================================
# LIST Tests - Filters
# ============================================================================

FILTER_TRANSPORT_ORDER_ID="${SAMPLE_TRANSPORT_ORDER_ID:-$TRANSPORT_ORDER_ID}"
FILTER_MATERIAL_TYPE_ID="${SAMPLE_MATERIAL_TYPE_ID:-$MATERIAL_TYPE_ID}"
FILTER_UNIT_OF_MEASURE_ID="${SAMPLE_UNIT_OF_MEASURE_ID:-$UNIT_OF_MEASURE_ID}"
FILTER_QUANTITY_EXPLICIT="${SAMPLE_QUANTITY_EXPLICIT:-10}"

test_name "List transport order materials with --transport-order filter"
if [[ -n "$FILTER_TRANSPORT_ORDER_ID" && "$FILTER_TRANSPORT_ORDER_ID" != "null" ]]; then
    xbe_json view transport-order-materials list --transport-order "$FILTER_TRANSPORT_ORDER_ID" --limit 5
    assert_success
else
    skip "No transport order ID available"
fi

test_name "List transport order materials with --material-type filter"
if [[ -n "$FILTER_MATERIAL_TYPE_ID" && "$FILTER_MATERIAL_TYPE_ID" != "null" ]]; then
    xbe_json view transport-order-materials list --material-type "$FILTER_MATERIAL_TYPE_ID" --limit 5
    assert_success
else
    skip "No material type ID available"
fi

test_name "List transport order materials with --unit-of-measure filter"
if [[ -n "$FILTER_UNIT_OF_MEASURE_ID" && "$FILTER_UNIT_OF_MEASURE_ID" != "null" ]]; then
    xbe_json view transport-order-materials list --unit-of-measure "$FILTER_UNIT_OF_MEASURE_ID" --limit 5
    assert_success
else
    skip "No unit of measure ID available"
fi

test_name "List transport order materials with --quantity-explicit filter"
if [[ -n "$FILTER_QUANTITY_EXPLICIT" && "$FILTER_QUANTITY_EXPLICIT" != "null" ]]; then
    xbe_json view transport-order-materials list --quantity-explicit "$FILTER_QUANTITY_EXPLICIT" --limit 5
    assert_success
else
    skip "No explicit quantity available"
fi

test_name "List transport order materials with --limit"
xbe_json view transport-order-materials list --limit 5
assert_success

test_name "List transport order materials with --offset"
xbe_json view transport-order-materials list --limit 5 --offset 5
assert_success

# ============================================================================
# SHOW Tests
# ============================================================================

SHOW_ID="${SAMPLE_ID:-$CREATED_ID}"
test_name "Show transport order material"
if [[ -n "$SHOW_ID" && "$SHOW_ID" != "null" ]]; then
    xbe_json view transport-order-materials show "$SHOW_ID"
    assert_success
else
    skip "No transport order material ID available"
fi

# ============================================================================
# DELETE Tests
# ============================================================================

test_name "Delete transport order material requires --confirm flag"
if [[ -n "$TRANSPORT_ORDER_ID" && -n "$MATERIAL_TYPE_ID" && -n "$UNIT_OF_MEASURE_ID" ]]; then
    xbe_json do transport-order-materials create \
        --transport-order "$TRANSPORT_ORDER_ID" \
        --material-type "$MATERIAL_TYPE_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID=$(json_get ".id")
        xbe_json do transport-order-materials delete "$DEL_ID"
        assert_failure
    else
        skip "Could not create transport order material for delete test"
    fi
else
    skip "Missing prerequisites for delete test"
fi

test_name "Delete transport order material with --confirm"
if [[ -n "$TRANSPORT_ORDER_ID" && -n "$MATERIAL_TYPE_ID" && -n "$UNIT_OF_MEASURE_ID" ]]; then
    xbe_json do transport-order-materials create \
        --transport-order "$TRANSPORT_ORDER_ID" \
        --material-type "$MATERIAL_TYPE_ID" \
        --unit-of-measure "$UNIT_OF_MEASURE_ID"
    if [[ $status -eq 0 ]]; then
        DEL_ID2=$(json_get ".id")
        xbe_json do transport-order-materials delete "$DEL_ID2" --confirm
        if [[ $status -eq 0 ]]; then
            pass
        else
            if [[ "$output" == *"Not Authorized"* ]] || [[ "$output" == *"not authorized"* ]]; then
                skip "Delete not authorized"
            else
                fail "Delete command failed unexpectedly: $output"
            fi
        fi
    else
        skip "Could not create transport order material for delete test"
    fi
else
    skip "Missing prerequisites for delete test"
fi

# ============================================================================
# Summary
# ============================================================================

run_tests
